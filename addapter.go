package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

type Message struct {
	APIKey  string `json:"apiKey"`
	Client  string `json:"client"`
	Content string `json:"content"`
}

const Client = "mac"

var apiKey = "test"

//var url = "ws://localhost:8080/ws"
var url = "wss://cheater-server-mbmu9.ondigitalocean.app/ws"

func Listen(req chan Message, res chan Message, done chan bool) {
	// Connect to the WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		showErrorDialog(err)
	}
	defer conn.Close()

	// Send a message to register the client
	message := Message{APIKey: apiKey, Client: Client, Content: "register"}
	if err := conn.WriteJSON(message); err != nil {
		showErrorDialog(err)
	}

	// Send periodic ping messages to keep the connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
					showErrorDialog(err)
				}
			case msg := <-res:
				if err := conn.WriteJSON(msg); err != nil {
					showErrorDialog(err)
					continue
				}
			}
		}
	}()

	// Wait for incoming messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			showErrorDialog(err)
			done <- true
			break
		}

		// Decode the incoming message
		var message Message
		if err := json.Unmarshal(msg, &message); err != nil {
			showErrorDialog(err)
			continue
		}

		fmt.Println("Received: ", message)
		req <- message
	}
}
