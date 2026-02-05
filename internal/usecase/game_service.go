package usecase

import (
    "context"
    "sync"
    "time"

    "github.com/google/uuid"
    "xo-server/internal/domain"
)

type gameService struct {
    games GameRepository
    mu    sync.Mutex
    locks map[uuid.UUID]*sync.Mutex
}

func NewGameService(games GameRepository) GameService {
    return &gameService{games: games, locks: make(map[uuid.UUID]*sync.Mutex)}
}

func (s *gameService) MakeMove(ctx context.Context, userID, gameID uuid.UUID, position int) (*domain.Game, error) {
    lock := s.getLock(gameID)
    lock.Lock()
    defer lock.Unlock()

    game, err := s.games.GetGameByID(ctx, gameID)
    if err != nil {
        return nil, err
    }

    if game.Status != domain.GameInProgress && game.Status != domain.GameDrawOffer {
        return nil, domain.ErrGameNotActive
    }

    symbol, err := playerSymbol(game, userID)
    if err != nil {
        return nil, err
    }

    if game.Status == domain.GameDrawOffer {
        game.Status = domain.GameInProgress
        game.DrawOfferedBy = nil
    }

    if game.NextTurn != symbol {
        return nil, domain.ErrNotYourTurn
    }

    if position < 0 || position > 8 {
        return nil, domain.ErrInvalidPosition
    }

    if game.Board[position] != '.' {
        return nil, domain.ErrPositionTaken
    }

    game.Board[position] = rune(symbol[0])
    if symbol == "X" {
        game.NextTurn = "O"
    } else {
        game.NextTurn = "X"
    }

    winnerSymbol := checkWinner(game.Board)
    if winnerSymbol != "" {
        game.Status = domain.GameFinished
        if winnerSymbol == "X" {
            game.WinnerUserID = &game.PlayerX
        } else {
            game.WinnerUserID = &game.PlayerO
        }
    } else if isBoardFull(game.Board) {
        game.Status = domain.GameFinished
        game.WinnerUserID = nil
    }

    game.UpdatedAt = time.Now().UTC()

    if err := s.games.UpdateGame(ctx, game); err != nil {
        return nil, err
    }

    move := &domain.GameMove{
        GameID:    game.ID,
        UserID:    userID,
        Position:  position,
        Symbol:    symbol,
        CreatedAt: time.Now().UTC(),
    }
    if err := s.games.AddMove(ctx, move); err != nil {
        return nil, err
    }

    return game, nil
}

func (s *gameService) Resign(ctx context.Context, userID, gameID uuid.UUID) (*domain.Game, error) {
    lock := s.getLock(gameID)
    lock.Lock()
    defer lock.Unlock()

    game, err := s.games.GetGameByID(ctx, gameID)
    if err != nil {
        return nil, err
    }

    if game.Status != domain.GameInProgress && game.Status != domain.GameDrawOffer {
        return nil, domain.ErrGameNotActive
    }

    if userID != game.PlayerX && userID != game.PlayerO {
        return nil, domain.ErrForbidden
    }

    winner := game.PlayerX
    if userID == game.PlayerX {
        winner = game.PlayerO
    }
    game.Status = domain.GameFinished
    game.WinnerUserID = &winner
    game.DrawOfferedBy = nil
    game.UpdatedAt = time.Now().UTC()

    if err := s.games.UpdateGame(ctx, game); err != nil {
        return nil, err
    }

    return game, nil
}

func (s *gameService) OfferDraw(ctx context.Context, userID, gameID uuid.UUID) (*domain.Game, error) {
    lock := s.getLock(gameID)
    lock.Lock()
    defer lock.Unlock()

    game, err := s.games.GetGameByID(ctx, gameID)
    if err != nil {
        return nil, err
    }

    if game.Status != domain.GameInProgress {
        return nil, domain.ErrGameNotActive
    }

    if userID != game.PlayerX && userID != game.PlayerO {
        return nil, domain.ErrForbidden
    }

    game.Status = domain.GameDrawOffer
    game.DrawOfferedBy = &userID
    game.UpdatedAt = time.Now().UTC()

    if err := s.games.UpdateGame(ctx, game); err != nil {
        return nil, err
    }

    return game, nil
}

func (s *gameService) AcceptDraw(ctx context.Context, userID, gameID uuid.UUID) (*domain.Game, error) {
    lock := s.getLock(gameID)
    lock.Lock()
    defer lock.Unlock()

    game, err := s.games.GetGameByID(ctx, gameID)
    if err != nil {
        return nil, err
    }

    if game.Status != domain.GameDrawOffer || game.DrawOfferedBy == nil {
        return nil, domain.ErrDrawNotOffered
    }

    if *game.DrawOfferedBy == userID {
        return nil, domain.ErrForbidden
    }

    game.Status = domain.GameFinished
    game.WinnerUserID = nil
    game.DrawOfferedBy = nil
    game.UpdatedAt = time.Now().UTC()

    if err := s.games.UpdateGame(ctx, game); err != nil {
        return nil, err
    }

    return game, nil
}

func (s *gameService) DeclineDraw(ctx context.Context, userID, gameID uuid.UUID) (*domain.Game, error) {
    lock := s.getLock(gameID)
    lock.Lock()
    defer lock.Unlock()

    game, err := s.games.GetGameByID(ctx, gameID)
    if err != nil {
        return nil, err
    }

    if game.Status != domain.GameDrawOffer || game.DrawOfferedBy == nil {
        return nil, domain.ErrDrawNotOffered
    }

    if *game.DrawOfferedBy == userID {
        return nil, domain.ErrForbidden
    }

    game.Status = domain.GameInProgress
    game.DrawOfferedBy = nil
    game.UpdatedAt = time.Now().UTC()

    if err := s.games.UpdateGame(ctx, game); err != nil {
        return nil, err
    }

    return game, nil
}

func (s *gameService) AddChat(ctx context.Context, userID, gameID uuid.UUID, message string) error {
    msg := &domain.GameMessage{
        GameID:    gameID,
        UserID:    userID,
        Message:   message,
        CreatedAt: time.Now().UTC(),
    }
    return s.games.AddMessage(ctx, msg)
}

func (s *gameService) GetGame(ctx context.Context, gameID uuid.UUID) (*domain.Game, error) {
    return s.games.GetGameByID(ctx, gameID)
}

func (s *gameService) GetActiveGames(ctx context.Context, userID uuid.UUID) ([]*domain.Game, error) {
    return s.games.ListActiveGamesByUser(ctx, userID)
}

func (s *gameService) getLock(gameID uuid.UUID) *sync.Mutex {
    s.mu.Lock()
    defer s.mu.Unlock()
    if l, ok := s.locks[gameID]; ok {
        return l
    }
    l := &sync.Mutex{}
    s.locks[gameID] = l
    return l
}

func playerSymbol(game *domain.Game, userID uuid.UUID) (string, error) {
    if userID == game.PlayerX {
        return "X", nil
    }
    if userID == game.PlayerO {
        return "O", nil
    }
    return "", domain.ErrForbidden
}

func checkWinner(b [9]rune) string {
    lines := [8][3]int{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {0, 3, 6}, {1, 4, 7}, {2, 5, 8}, {0, 4, 8}, {2, 4, 6}}
    for _, line := range lines {
        a, c, d := b[line[0]], b[line[1]], b[line[2]]
        if a != '.' && a == c && c == d {
            if a == 'X' {
                return "X"
            }
            if a == 'O' {
                return "O"
            }
        }
    }
    return ""
}

func isBoardFull(b [9]rune) bool {
    for _, r := range b {
        if r == '.' {
            return false
        }
    }
    return true
}
