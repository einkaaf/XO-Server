package usecase

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "xo-server/internal/adapter/auth"
    "xo-server/internal/adapter/repo/memory"
    "xo-server/internal/domain"
)

func TestAuthRegisterAndLogin(t *testing.T) {
    userRepo := memory.NewUserRepo()
    tokenProvider := auth.NewJWTProvider("secret", 0)
    svc := NewAuthService(userRepo, tokenProvider)

    user, err := svc.Register(context.Background(), "alice", "password")
    if err != nil {
        t.Fatalf("register error: %v", err)
    }

    token, u2, err := svc.Login(context.Background(), "alice", "password")
    if err != nil {
        t.Fatalf("login error: %v", err)
    }
    if token == "" {
        t.Fatalf("empty token")
    }
    if u2.ID != user.ID {
        t.Fatalf("expected same user")
    }
}

func TestGameMovesWin(t *testing.T) {
    repo := memory.NewGameRepo()
    svc := NewGameService(repo)

    game := &domain.Game{
        ID:       uuid.New(),
        PlayerX:  uuid.New(),
        PlayerO:  uuid.New(),
        Board:    domain.NewEmptyBoard(),
        NextTurn: "X",
        Status:   domain.GameInProgress,
    }
    if err := repo.CreateGame(context.Background(), game); err != nil {
        t.Fatalf("create game: %v", err)
    }

    _, err := svc.MakeMove(context.Background(), game.PlayerX, game.ID, 0)
    if err != nil {
        t.Fatalf("move 1: %v", err)
    }
    _, _ = svc.MakeMove(context.Background(), game.PlayerO, game.ID, 3)
    _, _ = svc.MakeMove(context.Background(), game.PlayerX, game.ID, 1)
    _, _ = svc.MakeMove(context.Background(), game.PlayerO, game.ID, 4)
    updated, err := svc.MakeMove(context.Background(), game.PlayerX, game.ID, 2)
    if err != nil {
        t.Fatalf("move 5: %v", err)
    }
    if updated.Status != domain.GameFinished || updated.WinnerUserID == nil || *updated.WinnerUserID != game.PlayerX {
        t.Fatalf("expected X win")
    }
}

func TestMatchmaking(t *testing.T) {
    repo := memory.NewGameRepo()
    mm := NewMatchmakingService(repo)

    u1 := uuid.New()
    u2 := uuid.New()

    matched, game, err := mm.JoinQueue(u1)
    if err != nil || matched || game != nil {
        t.Fatalf("expected waiting")
    }

    matched, game, err = mm.JoinQueue(u2)
    if err != nil || !matched || game == nil {
        t.Fatalf("expected match")
    }
}
