package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestAutoID(t *testing.T) {
	// Test if autoId generates a non-empty string
	id := autoId()
	assert.NotEmpty(t, id, "autoId should generate a non-empty string")
}

func TestWebSocketHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(webSocketHandler))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err, "Failed to connect to WebSocket")
	defer ws.Close()

	// Write a message to WebSocket
	message := []byte("Test message")
	err = ws.WriteMessage(websocket.TextMessage, message)
	assert.NoError(t, err, "Failed to write message to WebSocket")

	// Read the response from WebSocket
	_, response, err := ws.ReadMessage()
	assert.NoError(t, err, "Failed to read message from WebSocket")

	expectedResponse := []byte("Server received the message!")
	assert.Equal(t, expectedResponse, response, "Unexpected response from WebSocket")
}

func TestSetupRoutes(t *testing.T) {
	// Test if setupRoutes sets up routes correctly
	//router := http.NewServeMux()
	setupRoutes()
	assert.NotNil(t, http.DefaultServeMux, "DefaultServeMux should be set up")
	setupRoutes() // Call again to check for idempotence

	// Test if the static route is registered
	requestStatic, _ := http.NewRequest("GET", "/", nil)
	responseStatic := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(responseStatic, requestStatic)
	assert.Equal(t, http.StatusOK, responseStatic.Code, "Static route should return status OK")

	// Test if the WebSocket route is registered
	requestWS, _ := http.NewRequest("GET", "/ws", nil)
	responseWS := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(responseWS, requestWS)
	assert.Equal(t, http.StatusOK, responseWS.Code, "WebSocket route should return status OK")
}

func TestAddClientAndRemoveClient(t *testing.T) {
	ps := PubSub{}

	// Create a mock WebSocket connection
	mockConn := &websocket.Conn{}
	client := Client{
		Id:         autoId(),
		Connection: mockConn,
	}

	// Test AddClient
	ps.AddClient(client)
	assert.Len(t, ps.Clients, 1, "Number of clients should be 1 after adding")

	// Test RemoveClient
	ps.RemoveClient(client)
	assert.Len(t, ps.Clients, 0, "Number of clients should be 0 after removing")
}

func TestBroadcast(t *testing.T) {
	ps := PubSub{}

	// Create two mock WebSocket connections
	mockConn1 := &websocket.Conn{}
	mockConn2 := &websocket.Conn{}

	// Add clients to PubSub
	client1 := Client{Id: autoId(), Connection: mockConn1}
	client2 := Client{Id: autoId(), Connection: mockConn2}
	ps.AddClient(client1)
	ps.AddClient(client2)

	// Test Broadcast
	message := []byte("Test Broadcast")
	ps.broadcast(message)

	// Check if both clients received the message
	_, message1, _ := mockConn1.ReadMessage()
	_, message2, _ := mockConn2.ReadMessage()

	assert.Equal(t, message, message1, "Client1 should receive the broadcasted message")
	assert.Equal(t, message, message2, "Client2 should receive the broadcasted message")
}

func TestHandleRecvdMessage(t *testing.T) {
	ps := PubSub{}

	// Create a mock WebSocket connection
	mockConn := &websocket.Conn{}

	// Add a client to PubSub
	client := Client{Id: autoId(), Connection: mockConn}
	ps.AddClient(client)

	// Test HandleRecvdMessage
	messageType := websocket.TextMessage
	payload := []byte("Test HandleRecvdMessage")
	ps.HandleRecvdMessage(client, messageType, payload)

	// Check if the broadcasted message is received
	_, message, _ := mockConn.ReadMessage()
	expectedBroadcast := []byte("This is a Broadcast message sent by the Server! HELLO Clients!")
	assert.Equal(t, expectedBroadcast, message, "Client should receive the broadcasted message")
}

func TestMainFunction(t *testing.T) {
	// Test the main function by running it in a goroutine and checking if it starts without errors
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic in main function: %v", r)
			}
		}()
		main()
	}()
	// Allow some time for the server to start before testing
	time.Sleep(100 * time.Millisecond)

	// Send a request to the server to check if it is running
	response, err := http.Get("http://localhost:8080")
	assert.NoError(t, err, "Failed to send HTTP request to server")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Server should return status OK")

	// Stop the server by closing the default listener
	http.DefaultServeMux = nil
}
