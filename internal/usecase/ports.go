package usecase

import (
    "context"

    "github.com/google/uuid"
    "xo-server/internal/domain"
)

type UserRepository interface {
    CreateUser(ctx context.Context, user *domain.User) error
    GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
    GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type GameRepository interface {
    CreateGame(ctx context.Context, game *domain.Game) error
    GetGameByID(ctx context.Context, id uuid.UUID) (*domain.Game, error)
    UpdateGame(ctx context.Context, game *domain.Game) error
    ListActiveGamesByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Game, error)
    AddMove(ctx context.Context, move *domain.GameMove) error
    AddMessage(ctx context.Context, msg *domain.GameMessage) error
}

type TokenProvider interface {
    IssueToken(user *domain.User) (string, error)
    ParseToken(token string) (*domain.User, error)
}

type AuthService interface {
    Register(ctx context.Context, username, password string) (*domain.User, error)
    Login(ctx context.Context, username, password string) (string, *domain.User, error)
}

type GameService interface {
    MakeMove(ctx context.Context, userID, gameID uuid.UUID, position int) (*domain.Game, error)
    Resign(ctx context.Context, userID, gameID uuid.UUID) (*domain.Game, error)
    OfferDraw(ctx context.Context, userID, gameID uuid.UUID) (*domain.Game, error)
    AcceptDraw(ctx context.Context, userID, gameID uuid.UUID) (*domain.Game, error)
    DeclineDraw(ctx context.Context, userID, gameID uuid.UUID) (*domain.Game, error)
    AddChat(ctx context.Context, userID, gameID uuid.UUID, message string) error
    GetGame(ctx context.Context, gameID uuid.UUID) (*domain.Game, error)
    GetActiveGames(ctx context.Context, userID uuid.UUID) ([]*domain.Game, error)
}

type MatchmakingService interface {
    JoinQueue(userID uuid.UUID) (bool, *domain.Game, error)
}
