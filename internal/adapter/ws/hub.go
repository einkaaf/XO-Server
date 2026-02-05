package ws

import (
    "sync"

    "github.com/google/uuid"
)

type Hub struct {
    register   chan *Client
    unregister chan *Client

    mu      sync.RWMutex
    clients map[uuid.UUID]*Client
}

func NewHub() *Hub {
    return &Hub{
        register:   make(chan *Client),
        unregister: make(chan *Client),
        clients:    make(map[uuid.UUID]*Client),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case c := <-h.register:
            h.mu.Lock()
            h.clients[c.userID] = c
            h.mu.Unlock()
        case c := <-h.unregister:
            h.mu.Lock()
            if existing, ok := h.clients[c.userID]; ok && existing == c {
                delete(h.clients, c.userID)
                close(c.send)
            }
            h.mu.Unlock()
        }
    }
}

func (h *Hub) Register(c *Client) {
    h.register <- c
}

func (h *Hub) Unregister(c *Client) {
    h.unregister <- c
}

func (h *Hub) SendToUser(userID uuid.UUID, msg []byte) {
    h.mu.RLock()
    c := h.clients[userID]
    h.mu.RUnlock()
    if c == nil {
        return
    }
    select {
    case c.send <- msg:
    default:
    }
}

func (h *Hub) BroadcastToUsers(users []uuid.UUID, msg []byte) {
    for _, id := range users {
        h.SendToUser(id, msg)
    }
}
