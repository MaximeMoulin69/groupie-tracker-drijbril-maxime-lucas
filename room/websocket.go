package room

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Conn   *websocket.Conn
	RoomID int
	UserID int
	Pseudo string
	Send   chan []byte
}

type Hub struct {
	Rooms      map[int]map[int]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
}

type BroadcastMessage struct {
	RoomID  int
	Message []byte
	Exclude int
}

type Message struct {
	Type    string      `json:"type"`
	From    string      `json:"from"`
	UserID  int         `json:"user_id"`
	Content interface{} `json:"content"`
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[int]map[int]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *BroadcastMessage),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.Rooms[client.RoomID] == nil {
				h.Rooms[client.RoomID] = make(map[int]*Client)
			}
			h.Rooms[client.RoomID][client.UserID] = client
			h.mu.Unlock()
			log.Printf("Client %s (%d) connecte a la salle %d", client.Pseudo, client.UserID, client.RoomID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if room, ok := h.Rooms[client.RoomID]; ok {
				if _, ok := room[client.UserID]; ok {
					delete(room, client.UserID)
					close(client.Send)
					if len(room) == 0 {
						delete(h.Rooms, client.RoomID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("Client %s (%d) deconnecte de la salle %d", client.Pseudo, client.UserID, client.RoomID)

		case message := <-h.Broadcast:
			h.mu.RLock()
			if room, ok := h.Rooms[message.RoomID]; ok {
				for userID, client := range room {
					if message.Exclude != 0 && userID == message.Exclude {
						continue
					}
					select {
					case client.Send <- message.Message:
					default:
						close(client.Send)
						delete(room, userID)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Erreur WebSocket: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Erreur decodage message: %v", err)
			continue
		}

		msg.UserID = c.UserID
		msg.From = c.Pseudo

		encodedMsg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Erreur encodage message: %v", err)
			continue
		}

		hub.Broadcast <- &BroadcastMessage{
			RoomID:  c.RoomID,
			Message: encodedMsg,
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.Send
		if !ok {
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request, roomID int, userID int, pseudo string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erreur upgrade WebSocket: %v", err)
		return
	}

	client := &Client{
		Conn:   conn,
		RoomID: roomID,
		UserID: userID,
		Pseudo: pseudo,
		Send:   make(chan []byte, 256),
	}

	hub.Register <- client

	go client.WritePump()
	go client.ReadPump(hub)
}