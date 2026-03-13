package hub

import (
	"encoding/json"

	"go.uber.org/zap"
)

// HandleMessage processes incoming WebSocket messages from a client
func (h *Hub) HandleMessage(client *Client, msg map[string]any) {
	msgType, _ := msg["type"].(string)

	switch msgType {
	case "subscribe":
		// Handle joining a channel and subscribing to events
		channel, _ := msg["channel"].(string)
		config, _ := msg["config"].(map[string]any)
		
		if config != nil {
			postgresChanges, _ := config["postgres_changes"].([]any)
			for _, pcAny := range postgresChanges {
				if pc, ok := pcAny.(map[string]any); ok {
					event, _ := pc["event"].(string)
					schema, _ := pc["schema"].(string)
					table, _ := pc["table"].(string)
					filter, _ := pc["filter"].(string)

					if schema == "" {
						schema = "public"
					}

					client.Subscriptions = append(client.Subscriptions, Subscription{
						Schema:    schema,
						Table:     table,
						EventType: event,
						Filter:    filter,
					})
				}
			}
		}

		h.log.Info("client subscribed", zap.String("client_id", client.ID), zap.String("channel", channel))

		// Send confirmation
		select {
		case client.Send <- []byte(`{"event":"system","payload":{"status":"ok"},"ref":null}`):
		default:
		}

	case "unsubscribe":
		// Handle leaving a channel
		client.Subscriptions = []Subscription{} // simplify by clearing all for now
		h.log.Info("client unsubscribed", zap.String("client_id", client.ID))

	case "presence":
		// Handle tracking presence
		// Broadcast to everyone else
		msgBytes, _ := json.Marshal(msg)
		h.Broadcast(msg["channel"].(string), msgBytes)

	case "broadcast":
		// Handle broadcasting an ephemeral message
		msgBytes, _ := json.Marshal(msg)
		h.Broadcast(msg["channel"].(string), msgBytes)

	default:
		h.log.Debug("unknown message type", zap.String("type", msgType))
	}
}
