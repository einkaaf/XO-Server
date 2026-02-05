package postgres

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "xo-server/internal/domain"
)

type UserRepo struct {
    db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
    return &UserRepo{db: db}
}

func (r *UserRepo) CreateUser(ctx context.Context, user *domain.User) error {
    _, err := r.db.Exec(ctx, `
        INSERT INTO users (id, username, password_hash, created_at)
        VALUES ($1, $2, $3, $4)
    `, user.ID, user.Username, user.PasswordHash, user.CreatedAt)
    return err
}

func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
    row := r.db.QueryRow(ctx, `
        SELECT id, username, password_hash, created_at
        FROM users
        WHERE username = $1
    `, username)

    var u domain.User
    if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt); err != nil {
        return nil, domain.ErrNotFound
    }
    return &u, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    row := r.db.QueryRow(ctx, `
        SELECT id, username, password_hash, created_at
        FROM users
        WHERE id = $1
    `, id)

    var u domain.User
    if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt); err != nil {
        return nil, domain.ErrNotFound
    }
    return &u, nil
}

type GameRepo struct {
    db *pgxpool.Pool
}

func NewGameRepo(db *pgxpool.Pool) *GameRepo {
    return &GameRepo{db: db}
}

func (r *GameRepo) CreateGame(ctx context.Context, game *domain.Game) error {
    _, err := r.db.Exec(ctx, `
        INSERT INTO games (id, player_x, player_o, board, next_turn, status, winner_user_id, draw_offered_by, created_at, updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
    `, game.ID, game.PlayerX, game.PlayerO, domain.BoardToString(game.Board), game.NextTurn, string(game.Status), game.WinnerUserID, game.DrawOfferedBy, game.CreatedAt, game.UpdatedAt)
    return err
}

func (r *GameRepo) GetGameByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
    row := r.db.QueryRow(ctx, `
        SELECT id, player_x, player_o, board, next_turn, status, winner_user_id, draw_offered_by, created_at, updated_at
        FROM games
        WHERE id = $1
    `, id)

    var g domain.Game
    var boardStr string
    var status string
    if err := row.Scan(&g.ID, &g.PlayerX, &g.PlayerO, &boardStr, &g.NextTurn, &status, &g.WinnerUserID, &g.DrawOfferedBy, &g.CreatedAt, &g.UpdatedAt); err != nil {
        return nil, domain.ErrNotFound
    }

    board, err := domain.StringToBoard(boardStr)
    if err != nil {
        return nil, err
    }
    g.Board = board
    g.Status = domain.GameStatus(status)

    return &g, nil
}

func (r *GameRepo) UpdateGame(ctx context.Context, game *domain.Game) error {
    _, err := r.db.Exec(ctx, `
        UPDATE games
        SET board=$2, next_turn=$3, status=$4, winner_user_id=$5, draw_offered_by=$6, updated_at=$7
        WHERE id=$1
    `, game.ID, domain.BoardToString(game.Board), game.NextTurn, string(game.Status), game.WinnerUserID, game.DrawOfferedBy, game.UpdatedAt)
    return err
}

func (r *GameRepo) ListActiveGamesByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Game, error) {
    rows, err := r.db.Query(ctx, `
        SELECT id, player_x, player_o, board, next_turn, status, winner_user_id, draw_offered_by, created_at, updated_at
        FROM games
        WHERE (player_x = $1 OR player_o = $1) AND status IN ('waiting','in_progress','draw_offered')
        ORDER BY updated_at DESC
    `, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []*domain.Game
    for rows.Next() {
        var g domain.Game
        var boardStr string
        var status string
        if err := rows.Scan(&g.ID, &g.PlayerX, &g.PlayerO, &boardStr, &g.NextTurn, &status, &g.WinnerUserID, &g.DrawOfferedBy, &g.CreatedAt, &g.UpdatedAt); err != nil {
            return nil, err
        }
        board, err := domain.StringToBoard(boardStr)
        if err != nil {
            return nil, err
        }
        g.Board = board
        g.Status = domain.GameStatus(status)
        out = append(out, &g)
    }
    return out, nil
}

func (r *GameRepo) AddMove(ctx context.Context, move *domain.GameMove) error {
    _, err := r.db.Exec(ctx, `
        INSERT INTO game_moves (game_id, user_id, position, symbol, created_at)
        VALUES ($1,$2,$3,$4,$5)
    `, move.GameID, move.UserID, move.Position, move.Symbol, move.CreatedAt)
    return err
}

func (r *GameRepo) AddMessage(ctx context.Context, msg *domain.GameMessage) error {
    _, err := r.db.Exec(ctx, `
        INSERT INTO game_messages (game_id, user_id, message, created_at)
        VALUES ($1,$2,$3,$4)
    `, msg.GameID, msg.UserID, msg.Message, msg.CreatedAt)
    return err
}

func NewPool(ctx context.Context, conn string, max int32) (*pgxpool.Pool, error) {
    cfg, err := pgxpool.ParseConfig(conn)
    if err != nil {
        return nil, err
    }
    if max > 0 {
        cfg.MaxConns = max
    }
    cfg.MaxConnLifetime = 2 * time.Hour
    return pgxpool.NewWithConfig(ctx, cfg)
}
