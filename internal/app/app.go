package app

import (
    "context"
    "net/http"
    "time"

    "xo-server/internal/adapter/auth"
    httpadapter "xo-server/internal/adapter/http"
    "xo-server/internal/adapter/repo/postgres"
    "xo-server/internal/adapter/ws"
    "xo-server/internal/config"
    "xo-server/internal/usecase"
)

type App struct {
    Server *http.Server
    Hub    *ws.Hub
}

func Build(ctx context.Context, cfg *config.Config) (*App, error) {
    db, err := postgres.NewPool(ctx, cfg.DB.ConnString, cfg.DB.MaxConns)
    if err != nil {
        return nil, err
    }

    userRepo := postgres.NewUserRepo(db)
    gameRepo := postgres.NewGameRepo(db)

    tokenProvider := auth.NewJWTProvider(cfg.JWT.Secret, cfg.JWT.ParsedTTL)
    authSvc := usecase.NewAuthService(userRepo, tokenProvider)
    gameSvc := usecase.NewGameService(gameRepo)
    matchmaking := usecase.NewMatchmakingService(gameRepo)

    hub := ws.NewHub()
    wsHandler := ws.NewHandler(hub, tokenProvider, gameSvc, matchmaking)
    httpHandler := httpadapter.NewHandler(authSvc)

    mux := http.NewServeMux()
    httpHandler.RegisterRoutes(mux)
    mux.HandleFunc("/ws", wsHandler.ServeWS)

    server := &http.Server{
        Addr:              httpAddress(cfg.Server.HTTPPort),
        Handler:           mux,
        ReadHeaderTimeout: 5 * time.Second,
    }

    return &App{Server: server, Hub: hub}, nil
}

func httpAddress(port int) string {
    return ":" + itoa(port)
}

func itoa(v int) string {
    if v == 0 {
        return "0"
    }

    neg := v < 0
    if neg {
        v = -v
    }

    var b [20]byte
    i := len(b)
    for v > 0 {
        i--
        b[i] = byte('0' + v%10)
        v /= 10
    }
    if neg {
        i--
        b[i] = '-'
    }
    return string(b[i:])
}
