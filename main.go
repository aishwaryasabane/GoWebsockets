// This file creates an HTTP Websockets server that implements an in-memory, realtime, bi-directional PubSub system.
package main

import (
	"encoding/json"
	"fmt"
	"sync"

	//"goproject/go-chan/pubsub"
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

type PubSub struct {
	Clients       []Client
	Subscriptions []Subscription
	mu            sync.Mutex
}

type Client struct {
	Id         string
	Connection *websocket.Conn
}

type Message struct {
	Action  string          `json:"action"`
	Topic   string          `json:"topic"`
	Message json.RawMessage `json:"message"`
}

type Subscription struct {
	Topic  string
	Client *Client
}

const (
	PUBLISH     = "publish"
	SUBSCRIBE   = "subscribe"
	UNSUBSCRIBE = "unsubscribe"
)

// Function to generate a unique ID for every client.
// Returns:
// string - A unique identifier string.
func autoId() string {
	return uuid.Must(uuid.NewV4(), nil).String()
}

/*func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the Home Page of the Server!")
}*/

var ps = &PubSub{}

// Function to set up a basic HTTP server that listens on port 8080
// and upgrade incoming WebSocket connections. It handles WebSocket 
// connection requests and upgrades them using the Upgrader method.
// Parameters:
// w: http.ResponseWriter - The response writer to write HTTP responses.
// r: *http.Request - The incoming HTTP request.
func webSocketHandler(w http.ResponseWriter, r *http.Request) {

	//fmt.Fprintf(w, "Hello WebSocket!")
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
		// Read in a message
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

		//fmt.Printf("New message from client:%s", p)

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
	fmt.Println("This is the main function of the server")
	setupRoutes()
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
	ps.mu.Lock()
	defer ps.mu.Unlock()
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
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// first remove all subscriptions by this client

	for index, sub := range ps.Subscriptions {

		if client.Id == sub.Client.Id {
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}

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
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for _, client := range ps.Clients {
		err := client.Connection.WriteMessage(1, message)
		if err != nil {
			log.Println("Error writing message:", err)
			ps.RemoveClient(client)
		}
	}
}

// Function to get the client subscriptions and add subscriptions
func (ps *PubSub) GetSubscriptions(topic string, client *Client) []Subscription {

	var subscriptionList []Subscription

	for _, subscription := range ps.Subscriptions {

		if client != nil {

			if subscription.Client.Id == client.Id && subscription.Topic == topic {
				subscriptionList = append(subscriptionList, subscription)

			}
		} else {

			if subscription.Topic == topic {
				subscriptionList = append(subscriptionList, subscription)
			}
		}
	}

	return subscriptionList
}

// Function to subscribe to a topic
func (ps *PubSub) Subscribe(client *Client, topic string) *PubSub {

	clientSubs := ps.GetSubscriptions(topic, client)

	if len(clientSubs) > 0 {

		// client is subscribed this topic before

		return ps
	}

	newSubscription := Subscription{
		Topic:  topic,
		Client: client,
	}

	ps.Subscriptions = append(ps.Subscriptions, newSubscription)

	return ps
}

// Function to publish to a topic
func (ps *PubSub) Publish(topic string, message []byte, excludeClient *Client) {

	subscriptions := ps.GetSubscriptions(topic, nil)

	for _, sub := range subscriptions {

		fmt.Printf("Sending to client id %s message is %s \n", sub.Client.Id, message)
		//sub.Client.Connection.WriteMessage(1, message)

		sub.Client.Send(message)
	}

}

// Function to send a message 
func (client *Client) Send(message []byte) error {

	return client.Connection.WriteMessage(1, message)

}

// Function to unsubscribe to a topic
func (ps *PubSub) Unsubscribe(client *Client, topic string) *PubSub {

	//clientSubscriptions := ps.GetSubscriptions(topic, client)
	for index, sub := range ps.Subscriptions {

		if sub.Client.Id == client.Id && sub.Topic == topic {
			// found this subscription from client and we do need remove it
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}

	return ps

}

// Function to handle the messages received.
// Parameters:
// client: Client - The client from which the message was received.
// messageType: int - The type of the received message (e.g., TextMessage, BinaryMessage).
// payload: []byte - The payload of the received message.
// Returns:
// *PubSub - A pointer to the PubSub instance after handling the received message.
func (ps *PubSub) HandleRecvdMessage(client Client, messageType int, payload []byte) *PubSub {
	m := Message{}

	err := json.Unmarshal(payload, &m)
	if err != nil {
		fmt.Println("This is not correct message payload")
		return ps
	}

	switch m.Action {

	case PUBLISH:

		fmt.Println("This is publish new message")

		ps.Publish(m.Topic, m.Message, nil)

		break

	case SUBSCRIBE:

		ps.Subscribe(&client, m.Topic)

		fmt.Println("new subscriber to topic", m.Topic, len(ps.Subscriptions), client.Id)

		break

	case UNSUBSCRIBE:

		fmt.Println("Client want to unsubscribe the topic", m.Topic, client.Id)

		ps.Unsubscribe(&client, m.Topic)

		break

	default:
		break
	}

	return ps
	/*fmt.Printf("Client message payload: %s", payload)
	broadcastmsg := []byte("This is a Broadcast message sent by the Server! HELLO Clients!")
	ps.broadcast(broadcastmsg)
	return ps*/
}
