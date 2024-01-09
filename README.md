# GoWebsockets
HTTP+Websockets server that implements a Pub-Sub system.
- The main.go file consists of an HTTP+WebSockets server that implements an in-memory, realtime, bi-directional PubSub system.

The main.go file creates an HTTP+WebSockets server that implements an in-memory, realtime, bi-directional Pub-Sub system.

Current Functionality:

The main function calls the setupRoutes function which in turn calls the HTTP handler methods.
The HTTP ListenAndServe method uses the port 8080 and the DefaultServeMux handler to start a HTTP server.
The /ws address implements the webSocketHandler function.
The webSocketHandler function upgrades the basic HTTP connection to a websocket connection using the Upgrader method.
It then creates a client with a unique ID for every websocket connection.
AddClient function adds the new client to a list of clients that the Pub-Sub system keeps a track of.
RemoveClient removes a particular client from the list.
The PubSub struct consists of a list of clients and subscriptions.
Currently the server sends a message to the client when it connects and adds it to the list of clients.
When the client subscribes to a particular topic, it will receive all the messages being sent to that particular topic.


Google Chrome browser is Client1, it connects to the server by using localhost:8080. We see the following messages from the server
Server messsage: Hi Client! Server messsage: Hello Client IDea808062-128a-4f97-9d96-050d8a7a1b2d
Microsoft Edge browser is Client2, it connects to the server in the same way and sees the above messages.
Client1 sends a message to the server to subscribe to a topic.
Similarly, Client 2 sends a message to the server to subscribe to a topic.
When a message is published to the above topic, both the clients will receive it.
This is the basic implementation of a bi-directional Pub-Sub system where the server has 2 subscribers and with the current functionality, it sends a message to all its subscribers subscribed to a topic.
