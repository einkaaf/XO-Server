package memory

import (
    "context"
    "sync"

    "github.com/google/uuid"
    "xo-server/internal/domain"
)

type UserRepo struct {
    mu     sync.RWMutex
    byID   map[uuid.UUID]*domain.User
    byName map[string]*domain.User
}

func NewUserRepo() *UserRepo {
    return &UserRepo{byID: make(map[uuid.UUID]*domain.User), byName: make(map[string]*domain.User)}
}

func (r *UserRepo) CreateUser(ctx context.Context, user *domain.User) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.byName[user.Username]; ok {
        return domain.ErrInvalidInput
    }
    copy := *user
    r.byID[user.ID] = &copy
    r.byName[user.Username] = &copy
    return nil
}

func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    u, ok := r.byName[username]
    if !ok {
        return nil, domain.ErrNotFound
    }
    copy := *u
    return &copy, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    u, ok := r.byID[id]
    if !ok {
        return nil, domain.ErrNotFound
    }
    copy := *u
    return &copy, nil
}

type GameRepo struct {
    mu      sync.RWMutex
    games   map[uuid.UUID]*domain.Game
    moves   []*domain.GameMove
    messages []*domain.GameMessage
}

func NewGameRepo() *GameRepo {
    return &GameRepo{games: make(map[uuid.UUID]*domain.Game)}
}

func (r *GameRepo) CreateGame(ctx context.Context, game *domain.Game) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    copy := *game
    r.games[game.ID] = &copy
    return nil
}

func (r *GameRepo) GetGameByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    g, ok := r.games[id]
    if !ok {
        return nil, domain.ErrNotFound
    }
    copy := *g
    return &copy, nil
}

func (r *GameRepo) UpdateGame(ctx context.Context, game *domain.Game) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.games[game.ID]; !ok {
        return domain.ErrNotFound
    }
    copy := *game
    r.games[game.ID] = &copy
    return nil
}

func (r *GameRepo) ListActiveGamesByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Game, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    out := make([]*domain.Game, 0)
    for _, g := range r.games {
        if (g.PlayerX == userID || g.PlayerO == userID) && (g.Status == domain.GameWaiting || g.Status == domain.GameInProgress || g.Status == domain.GameDrawOffer) {
            copy := *g
            out = append(out, &copy)
        }
    }
    return out, nil
}

func (r *GameRepo) AddMove(ctx context.Context, move *domain.GameMove) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    copy := *move
    r.moves = append(r.moves, &copy)
    return nil
}

func (r *GameRepo) AddMessage(ctx context.Context, msg *domain.GameMessage) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    copy := *msg
    r.messages = append(r.messages, &copy)
    return nil
}
