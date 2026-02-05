package main

import (
    "context"
    "flag"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "xo-server/internal/app"
    "xo-server/internal/config"
)

func main() {
    cfgPath := flag.String("config", "config.yaml", "path to config")
    flag.Parse()

    cfg, err := config.Load(*cfgPath)
    if err != nil {
        log.Fatalf("config error: %v", err)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    built, err := app.Build(ctx, cfg)
    if err != nil {
        log.Fatalf("app build error: %v", err)
    }

    go built.Hub.Run()

    go func() {
        log.Printf("server listening on %s", built.Server.Addr)
        if err := built.Server.ListenAndServe(); err != nil && err != httpErrClosed {
            log.Fatalf("server error: %v", err)
        }
    }()

    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
    <-sigs

    shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancelShutdown()
    _ = built.Server.Shutdown(shutdownCtx)
}

var httpErrClosed = errString("http: Server closed")

type errString string

func (e errString) Error() string { return string(e) }
