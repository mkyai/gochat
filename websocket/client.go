package websocket

import (
	"encoding/json"
	"net/http"
	"time"
	"log"
	"gopoc/handlers"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
    UserID  string `json:"userId"`
    Message string `json:"message"`
}

type Client struct {
	hub *Hub
	conn *websocket.Conn
	send chan []byte
	UserID primitive.ObjectID    
	ChannelID primitive.ObjectID 
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		handlers.WriteMessageToDB(c.ChannelID, c.UserID, string(message))
		msg := Message{
            UserID:  c.UserID.Hex(),
            Message: string(message),
        }
        jsonMessage, err := json.Marshal(msg)
        if err != nil {
            log.Printf("Failed to marshal message: %v", err)
            continue
        }
		
		c.hub.broadcast <- jsonMessage
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(time.Second * 10)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
