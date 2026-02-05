package usecase

import (
    "context"
    "sync"
    "time"

    "github.com/google/uuid"
    "xo-server/internal/domain"
)

type matchmakingService struct {
    games GameRepository
    mu    sync.Mutex
    queue []uuid.UUID
}

func NewMatchmakingService(games GameRepository) MatchmakingService {
    return &matchmakingService{games: games, queue: make([]uuid.UUID, 0)}
}

func (s *matchmakingService) JoinQueue(userID uuid.UUID) (bool, *domain.Game, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    for _, id := range s.queue {
        if id == userID {
            return false, nil, domain.ErrAlreadyInQueue
        }
    }

    if len(s.queue) == 0 {
        s.queue = append(s.queue, userID)
        return false, nil, nil
    }

    opponent := s.queue[0]
    s.queue = s.queue[1:]

    if opponent == userID {
        s.queue = append(s.queue, userID)
        return false, nil, nil
    }

    game := &domain.Game{
        ID:        uuid.New(),
        PlayerX:   opponent,
        PlayerO:   userID,
        Board:     domain.NewEmptyBoard(),
        NextTurn:  "X",
        Status:    domain.GameInProgress,
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
    }

    if err := s.games.CreateGame(context.Background(), game); err != nil {
        return false, nil, err
    }

    return true, game, nil
}
