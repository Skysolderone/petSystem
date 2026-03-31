package ws

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.RWMutex
	devices map[uuid.UUID]map[*websocket.Conn]struct{}
	users   map[uuid.UUID]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{
		devices: map[uuid.UUID]map[*websocket.Conn]struct{}{},
		users:   map[uuid.UUID]map[*websocket.Conn]struct{}{},
	}
}

func (h *Hub) Subscribe(deviceID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.devices[deviceID] == nil {
		h.devices[deviceID] = map[*websocket.Conn]struct{}{}
	}
	h.devices[deviceID][conn] = struct{}{}
}

func (h *Hub) Unsubscribe(deviceID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	connections := h.devices[deviceID]
	if connections == nil {
		return
	}

	delete(connections, conn)
	if len(connections) == 0 {
		delete(h.devices, deviceID)
	}
}

func (h *Hub) Broadcast(deviceID uuid.UUID, message any) {
	h.mu.RLock()
	connections := h.devices[deviceID]
	h.mu.RUnlock()

	for conn := range connections {
		if err := conn.WriteJSON(message); err != nil {
			conn.Close()
			h.Unsubscribe(deviceID, conn)
		}
	}
}

func (h *Hub) SubscribeUser(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.users[userID] == nil {
		h.users[userID] = map[*websocket.Conn]struct{}{}
	}
	h.users[userID][conn] = struct{}{}
}

func (h *Hub) UnsubscribeUser(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	connections := h.users[userID]
	if connections == nil {
		return
	}

	delete(connections, conn)
	if len(connections) == 0 {
		delete(h.users, userID)
	}
}

func (h *Hub) BroadcastUser(userID uuid.UUID, message any) {
	h.mu.RLock()
	connections := h.users[userID]
	h.mu.RUnlock()

	for conn := range connections {
		if err := conn.WriteJSON(message); err != nil {
			conn.Close()
			h.UnsubscribeUser(userID, conn)
		}
	}
}
