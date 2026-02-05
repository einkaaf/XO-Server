# XO Server

Clean-architecture Tic-Tac-Toe (XO) server with WebSocket realtime play, JWT auth, matchmaking, chat, resign, and draw.

This document is the full project guide: setup, architecture, business rules, HTTP and WebSocket APIs, and a step-by-step client flow.

## Contents

1. What This Server Does
2. Architecture Overview
3. Setup and Run
4. Configuration
5. Database Schema and Migrations
6. Business Rules
7. HTTP API
8. WebSocket API
9. Client Flow (Step-by-Step)
10. Operational Notes

## 1) What This Server Does

- Users register and login.
- Clients connect via WebSocket using JWT.
- A simple matchmaking queue pairs two users into a game.
- Players exchange moves, chat, resign, or agree on a draw.
- Game state and history are persisted in Postgres.
- Reconnect is supported by re-syncing active games on connect.
- No spectators are supported.

## 2) Architecture Overview

Clean architecture layers:

- Domain: core entities and errors. `internal/domain/*`
- Use cases: business logic. `internal/usecase/*`
- Adapters: HTTP + WebSocket + repositories + JWT. `internal/adapter/*`
- App wiring: dependency injection and server setup. `internal/app/app.go`
- Entry point: `cmd/server/main.go`

Key components:

- `AuthService` handles registration and login.
- `MatchmakingService` is an in-memory FIFO queue.
- `GameService` enforces rules, validates moves, and manages draw/resign.
- `Hub` tracks active WebSocket connections for delivery.

## 3) Setup and Run

1. Start Postgres and run migrations.
1. Update `config.yaml`.
1. Start the server.

### Docker (recommended)

```bash
docker compose up -d
```

### Migrations

- PowerShell: `scripts\migrate.ps1`
- Bash: `scripts/migrate.sh`

### Run the server

```bash
go mod tidy
go run .\cmd\server --config config.yaml
```

Server listens on `http://localhost:8080` by default.

## 4) Configuration

File: `config.yaml`

```yaml
server:
  http_port: 8080
jwt:
  secret: "change-me"
  ttl: "24h"
db:
  conn_string: "postgres://postgres:postgres@localhost:5432/xo?sslmode=disable"
  max_conns: 10
```

Notes:

- `jwt.secret` must be non-empty.
- `jwt.ttl` uses Go duration format like `24h`, `1h30m`.
- `db.conn_string` must be valid.

## 5) Database Schema and Migrations

Migration: `migrations/001_init.sql`

Tables:

- `users` user accounts.
- `games` current and historical games.
- `game_moves` move history.
- `game_messages` chat history.

## 6) Business Rules

Registration and Auth:

- Username required.
- Password must be at least 6 characters.
- JWT is required to connect to WebSocket.

Matchmaking:

- Queue is FIFO.
- If a user is already in queue, `ErrAlreadyInQueue` is returned.
- When two users are matched, a game is created with:
  - PlayerX = first in queue
  - PlayerO = second
  - Status = `in_progress`
  - Next turn = `X`

Game Rules:

- Only `player_x` or `player_o` can interact with a game.
- Valid board positions are `0` to `8`.
- A position can only be played once.
- Turn order is enforced by `next_turn`.
- Game ends when:
  - A player wins, or
  - Board is full (draw), or
  - A player resigns.

Draw Flow:

- Any active player can offer a draw.
- If a draw is offered, status becomes `draw_offered`.
- The other player can accept or decline.
- Accept => game finished (draw).
- Decline => game returns to `in_progress`.

Reconnect:

- On WS connect, server sends `sync` with active games for that user.

No spectators are allowed.

## 7) HTTP API

Base URL: `http://host:8080`

### Health

`GET /health`

Response:

```json
{ "status": "ok" }
```

### Register

`POST /api/register`

Request:

```json
{ "username": "alice", "password": "secret123" }
```

Response:

```json
{ "user_id": "uuid" }
```

### Login

`POST /api/login`

Request:

```json
{ "username": "alice", "password": "secret123" }
```

Response:

```json
{
  "token": "JWT",
  "user_id": "uuid",
  "username": "alice"
}
```

Errors:

- `400` invalid input
- `401` unauthorized

## 8) WebSocket API

Connect:

`ws://host:8080/ws?token=JWT`

All messages use this envelope:

```json
{
  "type": "message_type",
  "payload": { }
}
```

### Client -> Server message types

`join_queue`

Payload:

```json
{}
```

`move`

Payload:

```json
{ "game_id": "uuid", "position": 0 }
```

`chat`

Payload:

```json
{ "game_id": "uuid", "message": "hi" }
```

`resign`

Payload:

```json
{ "game_id": "uuid" }
```

`draw_offer`

Payload:

```json
{ "game_id": "uuid" }
```

`draw_accept`

Payload:

```json
{ "game_id": "uuid" }
```

`draw_decline`

Payload:

```json
{ "game_id": "uuid" }
```

`sync`

Payload:

```json
{}
```

### Server -> Client event types

`queue_joined`

Payload:

```json
{ "status": "waiting" }
```

`game_found`

Payload:

```json
{ "game": Game }
```

`game_update`

Payload:

```json
{ "game": Game }
```

`chat`

Payload:

```json
{ "game_id": "uuid", "user_id": "uuid", "message": "hi", "at": "RFC3339" }
```

`sync`

Payload:

```json
{ "games": [Game] }
```

`error`

Payload:

```json
{ "message": "error details" }
```

### Game type

```json
{
  "id": "uuid",
  "player_x": "uuid",
  "player_o": "uuid",
  "board": "X..O.....",
  "next_turn": "X",
  "status": "in_progress",
  "winner_user_id": null,
  "draw_offered_by": null
}
```

`board` is 9 chars. `.` is empty.

Status values:

- `waiting`
- `in_progress`
- `draw_offered`
- `finished`

## 9) Client Flow (Step-by-Step)

This is the full flow a client should implement.

### 1) Register

Client calls:

`POST /api/register`

Keep the returned `user_id` for display.

### 2) Login

Client calls:

`POST /api/login`

Store the returned `token`.

### 3) Connect WebSocket

Open:

`ws://host:8080/ws?token=JWT`

Server will immediately send `sync` containing any active games for this user.

### 4) Join Matchmaking Queue

Send:

```json
{ "type": "join_queue", "payload": {} }
```

You will receive:

- `queue_joined` if waiting.
- `game_found` when a match is created.

### 5) Render Game

When you receive `game_found` or `game_update`, render the board:

- Use `board` string positions 0..8.
- Use `next_turn` to highlight whose move it is.
- If `status` is `finished`, show winner or draw.

### 6) Make a Move

Send:

```json
{ "type": "move", "payload": { "game_id": "uuid", "position": 0 } }
```

If valid, server broadcasts `game_update` to both players.
If invalid, server sends `error`.

### 7) Chat

Send:

```json
{ "type": "chat", "payload": { "game_id": "uuid", "message": "gg" } }
```

Both players receive `chat`.

### 8) Draw Offer

To offer:

```json
{ "type": "draw_offer", "payload": { "game_id": "uuid" } }
```

Other player receives `game_update` with `status = draw_offered`.

To accept:

```json
{ "type": "draw_accept", "payload": { "game_id": "uuid" } }
```

To decline:

```json
{ "type": "draw_decline", "payload": { "game_id": "uuid" } }
```

### 9) Resign

Send:

```json
{ "type": "resign", "payload": { "game_id": "uuid" } }
```

Both players receive `game_update` with `status = finished`.

### 10) Reconnect

On reconnect, open WebSocket with the same JWT:

`ws://host:8080/ws?token=JWT`

Server sends `sync` with active games.

## 10) Operational Notes

- The matchmaking queue is in-memory, so it resets on server restart.
- Games and history are persisted to Postgres.
- This server does not include rate limiting or admin APIs.
- This server has no spectators by design.
