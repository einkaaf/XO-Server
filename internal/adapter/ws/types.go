package ws

import (
    "encoding/json"
)

type Envelope struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

type ErrorPayload struct {
    Message string `json:"message"`
}

type GamePayload struct {
    Game *GameResponse `json:"game"`
}

type GameResponse struct {
    ID            string  `json:"id"`
    PlayerX       string  `json:"player_x"`
    PlayerO       string  `json:"player_o"`
    Board         string  `json:"board"`
    NextTurn      string  `json:"next_turn"`
    Status        string  `json:"status"`
    WinnerUserID  *string `json:"winner_user_id"`
    DrawOfferedBy *string `json:"draw_offered_by"`
}

type SyncPayload struct {
    Games []*GameResponse `json:"games"`
}

type ChatPayload struct {
    GameID  string `json:"game_id"`
    UserID  string `json:"user_id"`
    Message string `json:"message"`
    At      string `json:"at"`
}

type MoveRequest struct {
    GameID   string `json:"game_id"`
    Position int    `json:"position"`
}

type GameIDRequest struct {
    GameID string `json:"game_id"`
}

type ChatRequest struct {
    GameID  string `json:"game_id"`
    Message string `json:"message"`
}
