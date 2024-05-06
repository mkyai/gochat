package websocket
import (
	"net/http"
	"log"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopoc/handlers"
)

var HubInstance *Hub

type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

func newHub() *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        broadcast:  make(chan []byte),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
}

func (h *Hub) run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
            fmt.Println("Client registered")
        case message := <-h.broadcast:
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
				fmt.Println("Client unregistered")
                delete(h.clients, client)
                close(client.send)
            }
        }
    }
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	params := r.URL.Query()
    authorization := params.Get("token")
    if authorization == "" {
        http.Error(w, "Token is required", http.StatusUnauthorized)
        return
    }

	claims, err := handlers.ParseToken(authorization)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}
	channelID, err := primitive.ObjectIDFromHex(vars["channelID"])
	if err != nil {
		http.Error(w, "Invalid channel ID format", http.StatusBadRequest)
		return
	}

	if HubInstance == nil {
		fmt.Println("Creating new hub")
		HubInstance = newHub()
		go HubInstance.run()
	}

	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{hub: HubInstance, conn: conn, send: make(chan []byte, 256), UserID: userID, ChannelID: channelID }

	client.hub.register <- client

	go client.readPump()
	go client.writePump()
}
