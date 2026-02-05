-- 001_init.sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY,
    player_x UUID NOT NULL REFERENCES users(id),
    player_o UUID NOT NULL REFERENCES users(id),
    board CHAR(9) NOT NULL,
    next_turn CHAR(1) NOT NULL,
    status TEXT NOT NULL,
    winner_user_id UUID NULL REFERENCES users(id),
    draw_offered_by UUID NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_games_player_x ON games(player_x);
CREATE INDEX IF NOT EXISTS idx_games_player_o ON games(player_o);
CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);

CREATE TABLE IF NOT EXISTS game_moves (
    id BIGSERIAL PRIMARY KEY,
    game_id UUID NOT NULL REFERENCES games(id),
    user_id UUID NOT NULL REFERENCES users(id),
    position INT NOT NULL,
    symbol CHAR(1) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_game_moves_game_id ON game_moves(game_id);

CREATE TABLE IF NOT EXISTS game_messages (
    id BIGSERIAL PRIMARY KEY,
    game_id UUID NOT NULL REFERENCES games(id),
    user_id UUID NOT NULL REFERENCES users(id),
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_game_messages_game_id ON game_messages(game_id);
