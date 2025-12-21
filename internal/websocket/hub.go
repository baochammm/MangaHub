package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID       int64
	Username string
	Conn     *websocket.Conn
	Send     chan []byte
	Rooms    map[string]bool
}

type Room struct {
	Name    string
	Clients map[*Client]bool
	mu      sync.RWMutex
}
type ChatHub struct {
	Clients map[*Client]bool
	Rooms   map[string]*Room
	mu      sync.RWMutex

	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan ChatMessage
	JoinRoom   chan RoomAction
	LeaveRoom  chan RoomAction
}
type ChatMessage struct {
	Room      string `json:"room"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type RoomAction struct {
	Client *Client
	Room   string
}

func NewChatHub() *ChatHub {
	return &ChatHub{
		Clients:    make(map[*Client]bool),
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan ChatMessage),
		mu:         sync.RWMutex{},
		JoinRoom:   make(chan RoomAction),
		LeaveRoom:  make(chan RoomAction),
	}
}
func (hub *ChatHub) Run() {
	for {
		select {
		case client := <-hub.Register:
			hub.Clients[client] = true
		case client := <-hub.Unregister:
			if _, ok := hub.Clients[client]; ok {
				delete(hub.Clients, client)
				for room := range client.Rooms {
					hub.removeFromRoom(client, room)

				}
			}
		case action := <-hub.JoinRoom:
			hub.addToRoom(action.Client, action.Room)

		case action := <-hub.LeaveRoom:
			hub.removeFromRoom(action.Client, action.Room)
		case msg := <-hub.Broadcast:
			hub.broadcastToRoom(msg)
		}
	}
}
func (h *ChatHub) addToRoom(c *Client, roomName string) {
	room, ok := h.Rooms[roomName]
	if !ok {
		room = &Room{
			Name:    roomName,
			Clients: make(map[*Client]bool),
		}
		h.Rooms[roomName] = room
	}

	room.mu.Lock()
	room.Clients[c] = true
	room.mu.Unlock()
	log.Printf("[WS] Client %s joined room %s", c.Username, roomName)
	log.Printf("Current clients in room %s: %d", roomName, len(room.Clients))

	c.Rooms[roomName] = true
}
func (hub *ChatHub) removeFromRoom(client *Client, roomName string) {
	hub.Rooms[roomName].mu.Lock()
	// Check if the client is in the room
	if _, ok := hub.Rooms[roomName].Clients[client]; ok {
		delete(hub.Rooms[roomName].Clients, client)
		close(client.Send)
	}
	hub.Rooms[roomName].mu.Unlock()
	log.Printf("[WS] Client %s left room %s", client.Username, roomName)

}
func (hub *ChatHub) broadcastToRoom(msg ChatMessage) {
	hub.mu.RLock()
	// Find the room
	room, exists := hub.Rooms[msg.Room]
	hub.mu.RUnlock()
	if !exists {
		return
	}
	data, _ := json.Marshal(msg)
	room.mu.RLock()
	for client := range room.Clients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(room.Clients, client)
		}
	}
	room.mu.RUnlock()
}

func firstRoom(c *Client) string {
	for r := range c.Rooms {
		return r
	}
	return "general"
}
