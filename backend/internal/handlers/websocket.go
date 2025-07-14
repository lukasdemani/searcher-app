package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

type StatusUpdate struct {
	Type   string      `json:"type"`
	URLId  int         `json:"url_id"`
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *WebSocketHandler) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("WebSocket client connected. Total: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
				log.Printf("WebSocket client disconnected. Total: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("WebSocket write error: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	h.register <- conn

	// Handle incoming messages and cleanup
	go func() {
		defer func() {
			h.unregister <- conn
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}
		}
	}()
}

func (h *WebSocketHandler) BroadcastStatusUpdate(urlId int, status string, data interface{}) {
	update := StatusUpdate{
		Type:   "status_update",
		URLId:  urlId,
		Status: status,
		Data:   data,
	}

	message, err := json.Marshal(update)
	if err != nil {
		log.Printf("Error marshaling status update: %v", err)
		return
	}

	select {
	case h.broadcast <- message:
	default:
		log.Printf("Broadcast channel full, dropping message")
	}
}