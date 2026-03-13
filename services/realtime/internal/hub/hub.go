package hub

import (
	"encoding/json"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/machinelearningprodigy/OmniBase/realtime/internal/wal"
	"go.uber.org/zap"
)

// Subscription defines what a client is listening to
type Subscription struct {
	Schema    string // "public"
	Table     string // "*" for all tables, or specific table name
	EventType string // "*", "INSERT", "UPDATE", "DELETE"
	Filter    string // Optional WHERE-like filter
}

// Client represents a connected WebSocket client
type Client struct {
	ID            string
	UserID        string
	Role          string
	Conn          *websocket.Conn
	Subscriptions []Subscription
	Send          chan []byte
}

// Hub manages all WebSocket clients and fan-out of WAL events
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
	log     *zap.Logger
}

// NewHub creates a new connection hub
func NewHub(log *zap.Logger) *Hub {
	return &Hub{
		clients: make(map[string]*Client),
		log:     log,
	}
}

// Register adds a new WebSocket client to the hub
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.ID] = client
	h.log.Info("client connected",
		zap.String("client_id", client.ID),
		zap.String("user_id", client.UserID),
	)
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if client, ok := h.clients[clientID]; ok {
		close(client.Send)
		delete(h.clients, clientID)
	}
}

// BroadcastChange fans out a WAL change event to all subscribed clients
func (h *Hub) BroadcastChange(event wal.ChangeEvent) {
	payload, err := json.Marshal(map[string]any{
		"type":    "broadcast",
		"event":   event.EventType,
		"schema":  event.Schema,
		"table":   event.Table,
		"payload": event,
	})
	if err != nil {
		h.log.Error("failed to marshal change event", zap.Error(err))
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if h.clientIsSubscribed(client, event) {
			select {
			case client.Send <- payload:
			default:
				// Client is too slow — drop and log
				h.log.Warn("client send buffer full, dropping event",
					zap.String("client_id", client.ID),
				)
			}
		}
	}
}

// clientIsSubscribed checks if a client should receive this change event
func (h *Hub) clientIsSubscribed(client *Client, event wal.ChangeEvent) bool {
	for _, sub := range client.Subscriptions {
		schemaMatch := sub.Schema == "*" || sub.Schema == event.Schema
		tableMatch := sub.Table == "*" || sub.Table == event.Table
		eventMatch := sub.EventType == "*" || sub.EventType == event.EventType

		if schemaMatch && tableMatch && eventMatch {
			return true
		}
	}
	return false
}

// Broadcast sends an ephemeral message to clients in a specific channel (presence/broadcast)
func (h *Hub) Broadcast(channel string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := map[string]any{
		"type":    "presence",
		"channel": channel,
		"payload": json.RawMessage(payload),
	}
	data, _ := json.Marshal(msg)

	for _, client := range h.clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}

// ConnectedCount returns the number of active connections
func (h *Hub) ConnectedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
