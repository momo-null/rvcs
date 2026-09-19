package websocket

import (
	"sync"
	"time"

	"rvcs/internal/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Hub manages websocket clients and message routing between device and web clients.
type Hub struct {
	clients        map[string]*Client
	deviceMap      map[string]*Client
	userMap        map[string][]*Client
	broadcast      chan Message
	register       chan *Client
	unregister     chan *Client
	mu             sync.RWMutex
	MessageHandler MessageHandlerFunc
}

// MessageHandlerFunc handles incoming websocket messages.
type MessageHandlerFunc func(client *Client, msg *Message)

// NewHub creates a new websocket hub.
func NewHub() *Hub {
	hub := &Hub{
		clients:        make(map[string]*Client),
		deviceMap:      make(map[string]*Client),
		userMap:        make(map[string][]*Client),
		broadcast:      make(chan Message, 256),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		MessageHandler: nil,
	}

	hub.MessageHandler = hub.handleDeviceMessage
	return hub
}

// handleDeviceMessage forwards device-originated messages to web clients.
func (h *Hub) handleDeviceMessage(client *Client, msg *Message) {
	msgType := msg.Type

	logger.Info("Received WebSocket message",
		zap.String("type", string(msgType)),
		zap.String("device_id", client.DeviceID),
		zap.String("user_id", client.UserID),
		zap.String("state", msg.State),
		zap.String("status", msg.Status),
		zap.Int("data_length", len(msg.Data)),
	)

	if client.DeviceID == "" {
		return
	}

	data := map[string]interface{}{}
	for k, v := range msg.Data {
		data[k] = v
	}

	if msg.State != "" {
		data["state"] = msg.State
	}
	if msg.Status != "" {
		data["status"] = msg.Status
	}
	if msg.Command != "" {
		data["command"] = msg.Command
	}

	if msg.Error != nil {
		data["error"] = map[string]interface{}{
			"code":    msg.Error.Code,
			"message": msg.Error.Message,
		}
	}

	msgID := msg.MsgID
	if msgID == "" {
		msgID = uuid.New().String()
	}
	ts := msg.Timestamp
	if ts <= 0 {
		ts = time.Now().UnixMilli()
	}

	outMsg := Message{
		MsgID:     msgID,
		Type:      msgType,
		Command:   msg.Command,
		Status:    msg.Status,
		State:     msg.State,
		Timestamp: ts,
		Data:      data,
	}

	logger.Debug("Forwarding message to Web clients",
		zap.String("type", string(msgType)),
		zap.String("msg_id", msgID),
		zap.Int64("timestamp", ts),
		zap.Any("data", data),
	)

	h.broadcastToDeviceUsers(outMsg, client.DeviceID)
}

// broadcastToDeviceUsers sends a message to all web users (frontend filters by device id).
func (h *Hub) broadcastToDeviceUsers(msg Message, deviceID string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, clients := range h.userMap {
		for _, client := range clients {
			client.SendMsg(msg)
		}
	}

	logger.Debug("Message broadcasted to Web clients",
		zap.String("type", string(msg.Type)),
		zap.String("msg_id", msg.MsgID),
		zap.String("device_id", deviceID),
	)
}

// Run starts hub event loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a websocket client.
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client

	if client.DeviceID != "" {
		h.deviceMap[client.DeviceID] = client
	}

	if client.UserID != "" && client.DeviceID == "" {
		h.userMap[client.UserID] = append(h.userMap[client.UserID], client)
	}

	logger.Info("WebSocket client registered",
		zap.String("client_id", client.ID),
		zap.String("user_id", client.UserID),
		zap.String("device_id", client.DeviceID),
	)
}

// unregisterClient unregisters a websocket client.
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.ID]; !ok {
		return
	}

	delete(h.clients, client.ID)

	if client.DeviceID != "" {
		delete(h.deviceMap, client.DeviceID)
	}

	if client.UserID != "" && client.DeviceID == "" {
		if clients, exists := h.userMap[client.UserID]; exists {
			newClients := make([]*Client, 0, len(clients))
			for _, c := range clients {
				if c.ID != client.ID {
					newClients = append(newClients, c)
				}
			}
			if len(newClients) == 0 {
				delete(h.userMap, client.UserID)
			} else {
				h.userMap[client.UserID] = newClients
			}
		}
	}

	close(client.Send)

	logger.Info("WebSocket client unregistered",
		zap.String("client_id", client.ID),
		zap.String("user_id", client.UserID),
		zap.String("device_id", client.DeviceID),
	)
}

// broadcastMessage broadcasts message to all clients.
func (h *Hub) broadcastMessage(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		client.SendMsg(message)
	}
}

// Register registers a client.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// GetClientByDeviceID finds websocket client by device id.
func (h *Hub) GetClientByDeviceID(deviceID string) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, exists := h.deviceMap[deviceID]
	if !exists {
		logger.Debug("Device client not found in deviceMap", zap.String("device_id", deviceID))
		return nil
	}

	if client == nil {
		logger.Warn("Device client is nil in deviceMap", zap.String("device_id", deviceID))
		return nil
	}

	if _, clientExists := h.clients[client.ID]; !clientExists {
		logger.Warn("Device client connection not found", zap.String("device_id", deviceID), zap.String("client_id", client.ID))
		return nil
	}

	return client
}

// SendToDevice sends message to a specific device.
func (h *Hub) SendToDevice(deviceID string, message Message) bool {
	client := h.GetClientByDeviceID(deviceID)
	if client == nil {
		logger.Warn("Device client not found", zap.String("device_id", deviceID))
		return false
	}

	client.SendMsg(message)
	return true
}

// Broadcast broadcasts to all clients.
func (h *Hub) Broadcast(message Message) {
	h.broadcast <- message
}

// BroadcastToUsers broadcasts to web user clients.
func (h *Hub) BroadcastToUsers(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, clients := range h.userMap {
		for _, client := range clients {
			client.SendMsg(message)
		}
	}
}

// BroadcastToDeviceUsers broadcasts to all web users (frontend filters by device id).
func (h *Hub) BroadcastToDeviceUsers(deviceID string, message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, clients := range h.userMap {
		for _, client := range clients {
			client.SendMsg(message)
		}
	}
}

// defaultMessageHandler is a fallback message handler.
func defaultMessageHandler(client *Client, msg *Message) {
	logger.Info("Received message",
		zap.String("device_id", client.DeviceID),
		zap.String("msg_type", string(msg.Type)),
	)
}
