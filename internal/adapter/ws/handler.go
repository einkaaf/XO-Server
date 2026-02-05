package ws

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "github.com/gorilla/websocket"
    "github.com/google/uuid"
    "xo-server/internal/domain"
    "xo-server/internal/usecase"
)

type Handler struct {
    hub         *Hub
    auth        usecase.TokenProvider
    games       usecase.GameService
    matchmaking usecase.MatchmakingService
}

func NewHandler(hub *Hub, auth usecase.TokenProvider, games usecase.GameService, matchmaking usecase.MatchmakingService) *Handler {
    return &Handler{hub: hub, auth: auth, games: games, matchmaking: matchmaking}
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
    token := r.URL.Query().Get("token")
    if token == "" {
        http.Error(w, "missing token", http.StatusUnauthorized)
        return
    }

    user, err := h.auth.ParseToken(token)
    if err != nil {
        http.Error(w, "invalid token", http.StatusUnauthorized)
        return
    }

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }

    client := &Client{
        hub:      h.hub,
        conn:     conn,
        send:     make(chan []byte, 256),
        userID:   user.ID,
        username: user.Username,
        handler:  h,
    }

    h.hub.Register(client)

    go client.writePump()
    go client.readPump()

    h.sendSync(client)
}

func (h *Handler) sendSync(c *Client) {
    games, err := h.games.GetActiveGames(context.Background(), c.userID)
    if err != nil {
        return
    }

    payload := SyncPayload{Games: make([]*GameResponse, 0, len(games))}
    for _, g := range games {
        payload.Games = append(payload.Games, toGameResponse(g))
    }

    sendJSON(c, "sync", payload)
}

func (h *Handler) handleMessage(c *Client, msg Envelope) {
    switch msg.Type {
    case "join_queue":
        h.handleJoinQueue(c)
    case "move":
        h.handleMove(c, msg.Payload)
    case "chat":
        h.handleChat(c, msg.Payload)
    case "resign":
        h.handleResign(c, msg.Payload)
    case "draw_offer":
        h.handleDrawOffer(c, msg.Payload)
    case "draw_accept":
        h.handleDrawAccept(c, msg.Payload)
    case "draw_decline":
        h.handleDrawDecline(c, msg.Payload)
    case "sync":
        h.sendSync(c)
    default:
        sendError(c, "unknown message type")
    }
}

func (h *Handler) handleJoinQueue(c *Client) {
    matched, game, err := h.matchmaking.JoinQueue(c.userID)
    if err != nil {
        sendError(c, err.Error())
        return
    }
    if !matched {
        sendJSON(c, "queue_joined", map[string]string{"status": "waiting"})
        return
    }

    resp := toGameResponse(game)
    payload := GamePayload{Game: resp}
    msg := mustJSON(Envelope{Type: "game_found", Payload: mustRaw(payload)})
    h.hub.BroadcastToUsers([]uuid.UUID{game.PlayerX, game.PlayerO}, msg)
}

func (h *Handler) handleMove(c *Client, raw json.RawMessage) {
    var req MoveRequest
    if err := json.Unmarshal(raw, &req); err != nil {
        sendError(c, "invalid payload")
        return
    }

    gameID, err := uuid.Parse(req.GameID)
    if err != nil {
        sendError(c, "invalid game_id")
        return
    }

    game, err := h.games.MakeMove(context.Background(), c.userID, gameID, req.Position)
    if err != nil {
        sendError(c, err.Error())
        return
    }

    h.broadcastGameUpdate(game)
}

func (h *Handler) handleChat(c *Client, raw json.RawMessage) {
    var req ChatRequest
    if err := json.Unmarshal(raw, &req); err != nil {
        sendError(c, "invalid payload")
        return
    }

    gameID, err := uuid.Parse(req.GameID)
    if err != nil {
        sendError(c, "invalid game_id")
        return
    }

    if err := h.games.AddChat(context.Background(), c.userID, gameID, req.Message); err != nil {
        sendError(c, err.Error())
        return
    }

    payload := ChatPayload{
        GameID:  gameID.String(),
        UserID:  c.userID.String(),
        Message: req.Message,
        At:      time.Now().UTC().Format(time.RFC3339),
    }
    msg := mustJSON(Envelope{Type: "chat", Payload: mustRaw(payload)})
    game, err := h.games.GetGame(context.Background(), gameID)
    if err != nil {
        return
    }
    h.hub.BroadcastToUsers([]uuid.UUID{game.PlayerX, game.PlayerO}, msg)
}

func (h *Handler) handleResign(c *Client, raw json.RawMessage) {
    gameID, ok := parseGameID(c, raw)
    if !ok {
        return
    }
    game, err := h.games.Resign(context.Background(), c.userID, gameID)
    if err != nil {
        sendError(c, err.Error())
        return
    }
    h.broadcastGameUpdate(game)
}

func (h *Handler) handleDrawOffer(c *Client, raw json.RawMessage) {
    gameID, ok := parseGameID(c, raw)
    if !ok {
        return
    }
    game, err := h.games.OfferDraw(context.Background(), c.userID, gameID)
    if err != nil {
        sendError(c, err.Error())
        return
    }
    h.broadcastGameUpdate(game)
}

func (h *Handler) handleDrawAccept(c *Client, raw json.RawMessage) {
    gameID, ok := parseGameID(c, raw)
    if !ok {
        return
    }
    game, err := h.games.AcceptDraw(context.Background(), c.userID, gameID)
    if err != nil {
        sendError(c, err.Error())
        return
    }
    h.broadcastGameUpdate(game)
}

func (h *Handler) handleDrawDecline(c *Client, raw json.RawMessage) {
    gameID, ok := parseGameID(c, raw)
    if !ok {
        return
    }
    game, err := h.games.DeclineDraw(context.Background(), c.userID, gameID)
    if err != nil {
        sendError(c, err.Error())
        return
    }
    h.broadcastGameUpdate(game)
}

func (h *Handler) broadcastGameUpdate(game *domain.Game) {
    payload := GamePayload{Game: toGameResponse(game)}
    msg := mustJSON(Envelope{Type: "game_update", Payload: mustRaw(payload)})
    h.hub.BroadcastToUsers([]uuid.UUID{game.PlayerX, game.PlayerO}, msg)
}

func toGameResponse(game *domain.Game) *GameResponse {
    var winner *string
    if game.WinnerUserID != nil {
        w := game.WinnerUserID.String()
        winner = &w
    }
    var drawBy *string
    if game.DrawOfferedBy != nil {
        d := game.DrawOfferedBy.String()
        drawBy = &d
    }
    return &GameResponse{
        ID:            game.ID.String(),
        PlayerX:       game.PlayerX.String(),
        PlayerO:       game.PlayerO.String(),
        Board:         domain.BoardToString(game.Board),
        NextTurn:      game.NextTurn,
        Status:        string(game.Status),
        WinnerUserID:  winner,
        DrawOfferedBy: drawBy,
    }
}

func parseGameID(c *Client, raw json.RawMessage) (uuid.UUID, bool) {
    var req GameIDRequest
    if err := json.Unmarshal(raw, &req); err != nil {
        sendError(c, "invalid payload")
        return uuid.UUID{}, false
    }
    id, err := uuid.Parse(req.GameID)
    if err != nil {
        sendError(c, "invalid game_id")
        return uuid.UUID{}, false
    }
    return id, true
}

func mustJSON(v interface{}) []byte {
    b, _ := json.Marshal(v)
    return b
}

func mustRaw(v interface{}) json.RawMessage {
    b, _ := json.Marshal(v)
    return json.RawMessage(b)
}

func sendJSON(c *Client, msgType string, payload interface{}) {
    env := Envelope{Type: msgType, Payload: mustRaw(payload)}
    c.send <- mustJSON(env)
}

func sendError(c *Client, message string) {
    env := Envelope{Type: "error", Payload: mustRaw(ErrorPayload{Message: message})}
    c.send <- mustJSON(env)
}
