package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

const Client = "mac"
const ApiKey = "test"

var conn *websocket.Conn
var done chan struct{}
var dis chan struct{}

func init() {
	dis = make(chan struct{})
	done = make(chan struct{})
}

type Message struct {
	ApiKey  string `json:"apiKey"`
	Client  string `json:"client"`
	Content string `json:"content"`
}

func ListenWS(meg chan Message) {
	connect()
	defer conn.Close()
	listen(meg)
	register()
	disconnect()
	<-done
	time.Sleep(1 * time.Second)
	ListenWS(meg)
}

func connect() *websocket.Conn {

	// Define the WebSocket URL
	serverURL := url.URL{Scheme: "wss", Host: "cheater-server-mbmu9.ondigitalocean.app", Path: "/ws"}

	// Connect to the WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(serverURL.String(), nil)
	if err != nil {
		log.Fatal("Error connecting to the WebSocket server:", err)
	}

	return conn
}

func disconnect() {
	// Handle termination signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	select {
	case <-interrupt:
		log.Println("Received interrupt signal, closing connection...")
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("Error sending close message:", err)
		}
	}
	<-dis
	<-done
}

func listen(meg chan Message) {
	// Listen for server messages and print them in the console
	go func() {
		defer close(done)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading server message:", err)
				continue
			}
			message := Message{}
			err = json.Unmarshal(data, &message)
			if err != nil {
				log.Println("Error unmarshalling JSON:", err)
				continue
			}
			meg <- message
		}
	}()
}

func register() {
	// Send a message with the JSON content
	message := Message{
		ApiKey:  ApiKey,
		Client:  Client,
		Content: "register",
	}

	SendMessage(message)
}

func SendMessage(message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshalling JSON:", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Println("Error sending message:", err)
		<-dis
		<-done

		time.Sleep(5 * time.Second)
		SendMessage(message)
	}
}
