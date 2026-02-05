package ws

import (
    "encoding/json"
    "time"

    "github.com/gorilla/websocket"
    "github.com/google/uuid"
)

type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    send     chan []byte
    userID   uuid.UUID
    username string
    handler  *Handler
}

const (
    writeWait      = 10 * time.Second
    pongWait       = 60 * time.Second
    pingPeriod     = (pongWait * 9) / 10
    maxMessageSize = 4096
)

func (c *Client) readPump() {
    defer func() {
        c.hub.Unregister(c)
        _ = c.conn.Close()
    }()

    c.conn.SetReadLimit(maxMessageSize)
    _ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        _ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }

        var env Envelope
        if err := json.Unmarshal(message, &env); err != nil {
            sendError(c, "invalid message")
            continue
        }
        c.handler.handleMessage(c, env)
    }
}

func (c *Client) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        _ = c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            _ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                _ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
                return
            }
        case <-ticker.C:
            _ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
