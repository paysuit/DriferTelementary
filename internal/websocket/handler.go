package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	h.hub.AddClient(conn)

	log.Printf("Client connected: %s", conn.RemoteAddr())

	go h.readLoop(conn)
}

func (h *Handler) readLoop(conn *websocket.Conn) {

	defer func() {
		h.hub.RemoveClient(conn)
		conn.Close()
	}()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		log.Printf("Received: %s", msg)

		err = conn.WriteMessage(msgType, []byte("ACK"))
		if err != nil {
			break
		}

		h.hub.Broadcast(msg)
	}
}
