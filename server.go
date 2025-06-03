package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// notifySubscribers sends updated HTML content to all subscribers
func notifySubscribers(html HTML) {
	for _, ch := range subscribers {
		ch <- html
	}
}

// htmlHandler serves the raw HTML content
func htmlHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	fmt.Fprintf(w, string(htmlContent)) // Serve the HTML content directly
	mu.Unlock()
}

// serveTemplate serves the HTML template with live preview functionality
func serveTemplate(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Live Preview</title>
    <style>
        body { font-family: Arial, sans-serif; }
        #preview { border: 1px solid #ccc; padding: 10px; margin-top: 20px; }
    </style>
</head>
<body>
    <h1>MJML Live Preview</h1>
    <div id="preview"></div>
    <script>
        const ws = new WebSocket("ws://localhost:8080/ws");

        ws.onmessage = function(event) {
            document.getElementById("preview").innerHTML = event.data;
        };

        ws.onclose = function() {
            console.log("WebSocket closed");
        };
    </script>
</body>
</html> `
	t, err := template.New("preview").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template with the current HTML content
	t.Execute(w, htmlContent)
}

// broadcast sends a message to all connected WebSocket clients
func broadcast(message HTML) {
	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("Error sending message to client:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// serveWs handles WebSocket connections
func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error during connection upgrade:", err)
		return
	}
	defer conn.Close()

	clients[conn] = true

	// Send the current HTML content to the new client immediately
	mu.Lock()
	err = conn.WriteMessage(websocket.TextMessage, []byte(htmlContent))
	mu.Unlock()
	if err != nil {
		log.Println("Error sending current HTML content:", err)
		return
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	delete(clients, conn)
}

// startServer starts the HTTP server for live preview
func startServer() {
	http.HandleFunc("/ws", serveWs)
	http.HandleFunc("/html", htmlHandler) // Serve HTML content at /html
	http.HandleFunc("/", serveTemplate)   // Serve the HTML template at /
	http.ListenAndServe(":8080", nil)
}
