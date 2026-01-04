package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now, can be restricted later
	},
}

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Manager struct {
	clients map[*Client]bool
	mu      sync.RWMutex
	broadcast chan Event
}

type Client struct {
	manager *Manager
	conn    *websocket.Conn
	send    chan Event
}

func NewManager() *Manager {
	return &Manager{
		clients:   make(map[*Client]bool),
		broadcast: make(chan Event),
	}
}

func (m *Manager) Start() {
	slog.Info("Starting WebSocket Manager")
	for {
		select {
		case event := <-m.broadcast:
			m.mu.RLock()
			for client := range m.clients {
				select {
				case client.send <- event:
				default:
					close(client.send)
					delete(m.clients, client)
				}
			}
			m.mu.RUnlock()
		}
	}
}

func (m *Manager) Broadcast(eventType string, payload interface{}) {
	m.broadcast <- Event{Type: eventType, Payload: payload}
}

func (m *Manager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", "error", err)
		return
	}

	client := &Client{
		manager: m,
		conn:    conn,
		send:    make(chan Event, 256),
	}

	m.mu.Lock()
	m.clients[client] = true
	m.mu.Unlock()

	go client.readPump()
	go client.writePump()
}

func (c *Client) readPump() {
	defer func() {
		c.manager.mu.Lock()
		delete(c.manager.clients, c)
		c.manager.mu.Unlock()
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("WebSocket read error", "error", err)
			}
			break
		}
		
		var event Event
		if err := json.Unmarshal(message, &event); err != nil {
			slog.Warn("WebSocket received invalid JSON", "error", err)
			continue
		}

		slog.Debug("WebSocket received event", "type", event.Type)
		// Route message back if needed
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case event, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(event); err != nil {
				slog.Error("WebSocket write error", "error", err)
				return
			}
		}
	}
}
