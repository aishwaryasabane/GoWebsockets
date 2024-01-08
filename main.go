// This file creates an HTTP+Websockets server that implements an in-memory, realtime, bi-directional PubSub system.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"github.com/satori/uuid"
)

// Define an upgrader to upgrade the basic HTTP connection to a websocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// The Pub-Sub system maintains a list of subscribers (clients)
type PubSub struct {
	Clients []Client
}

// Every client maintains a unique ID and a websocket connection with the server  
type Client struct {
	Id         string
	Connection *websocket.Conn
}

// The message consists of action, topic and message fields. The topic functionality can be added to the Pub-Sub system using this struct.
type Message struct {
	Action  string          `json:"action"`
	Topic   string          `json:"topic"`
	Message json.RawMessage `json:"message"`
}

// Function to generate a unique ID for every client.
// Returns:
// string - A unique identifier string.
func autoId() string {
	return uuid.Must(uuid.NewV4(), nil).String()
}

var ps = &PubSub{}

// Function to set up a basic HTTP server that listens on port 8080
// and upgrade incoming WebSocket connections. It handles WebSocket 
// connection requests and upgrades them using the Upgrader method.
// Parameters:
// w: http.ResponseWriter - The response writer to write HTTP responses.
// r: *http.Request - The incoming HTTP request.
func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	//Upgrade this connection to a WebSocket connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	// Create a client and assign it a Unique ID
	client := Client{
		Id:         autoId(),
		Connection: ws,
	}

	// Send a message to the client
	fmt.Printf("Client Connected:%s", client.Id)
	err = ws.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		log.Println(err)
	}

	// Add client to the list of clients
	ps.AddClient(client)

	// Listen indefinitely for new messages coming through on our WebSocket connection
	for {
		// Read a message
		messageType, p, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// Print out the message for clarity
		log.Println(string(p))

		// Send a message indicating the message was received
		response := []byte("Server received the message!")
		if err := ws.WriteMessage(messageType, response); err != nil {
			log.Println(err)
			return
		}

		// Call the handler to handle the received message from the client
		ps.HandleRecvdMessage(client, messageType, p)
	}

}

// Function to configure and handle the HTTP routes for the server.
// It sets up two routes: one for serving static files and another for handling
// WebSocket connections. The static route serves files from the "static" directory
// and the WebSocket route uses the webSocketHandler function to handle incoming
// WebSocket connections. 
func setupRoutes() {
	// Serve static files from the static directory
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static")
	})
	// Handle WebSocket connections using the webSocketHandler function
	http.HandleFunc("/ws", webSocketHandler)
}

func main() {
	fmt.Println("This is the main function of the Server")
	setupRoutes()
        // The server will listen on port 8080
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// Function to add a new client to the list
// Parameters:
// client: Client - The client to be added to the list.
// Returns:
// *PubSub - A pointer to the updated PubSub instance after adding the client.
func (ps *PubSub) AddClient(client Client) *PubSub {
	ps.Clients = append(ps.Clients, client)
	fmt.Println("Adding new client to the list", client.Id, len(ps.Clients))
	payload := []byte("Hello Client ID" + client.Id)
	client.Connection.WriteMessage(1, payload)
	return ps
}

// Function to remove a client from the list
// Parameters:
// client: Client - The client to be removed from the list.
// Returns:
// *PubSub - A pointer to the updated PubSub instance after removing the client.
func (ps *PubSub) RemoveClient(client Client) *PubSub {
	for i, cl := range ps.Clients {
		if cl.Id == client.Id {
			ps.Clients = append(ps.Clients[:i], ps.Clients[i+1:]...)
		}
	}
	return ps
}

// Function to send a message to all the clients in the Pub-Sub system when any client sends a message.
// Parameters:
// message: []byte - The message to be broadcasted to all clients.
func (ps *PubSub) broadcast(message []byte) {
	for _, client := range ps.Clients {
		err := client.Connection.WriteMessage(1, message)
		if err != nil {
			log.Println("Error writing message:", err)
			ps.RemoveClient(client)
		}
	}
}

// Function to handle the messages received.
// Parameters:
// client: Client - The client from which the message was received.
// messageType: int - The type of the received message (e.g., TextMessage, BinaryMessage).
// payload: []byte - The payload of the received message.
// Returns:
// *PubSub - A pointer to the PubSub instance after handling the received message.
func (ps *PubSub) HandleRecvdMessage(client Client, messageType int, payload []byte) *PubSub {
	fmt.Printf("Client message payload: %s", payload)
        // Just broadcast a message to all the subscribers (clients) whenever any client sends a message
	broadcastmsg := []byte("This is a Broadcast message sent by the Server! HELLO Clients!")
	ps.broadcast(broadcastmsg)
	return ps
}
