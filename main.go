package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

type MJMLResponse struct {
	HTML string `json:"html"`
}

type Email struct {
	from     *string
	password *string
	subject  *string
	htmlFile *string
	mjmlFile *string
	watch    *bool
	to       *string
}

type HTML string

var upgrader = websocket.Upgrader{}
var clients = make(map[*websocket.Conn]bool)
var htmlContent = HTML("")
var mu sync.Mutex
var subscribers = make([]chan HTML, 0)

// Function to monitor the MJML file for changes
func watchMJMLFile(filePath string, htmlFilePath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Add the MJML file to watch
	err = watcher.Add(filePath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Watching for changes to", filePath)

	// Run an infinite loop to watch for events
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Check if the event is a write event (file modified)
			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Println("File modified:", event.Name)

				// Regenerate the HTML
				data, err := os.ReadFile(filePath)
				html := convertToHtml(string(data))
				if err != nil {
					fmt.Println("Error regenerating HTML:", err)
					continue
				}
				fmt.Print(html)

				// Save the generated HTML to a file
				err = os.WriteFile(htmlFilePath, []byte(html), 0644)
				if err != nil {
					fmt.Println("Error saving HTML to file:", err)
					continue
				}

				htmlContent = html
				broadcast(htmlContent) // Send updated content to all connected clients
				notifySubscribers(html)
				fmt.Println("HTML saved successfully to", htmlFilePath)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				fmt.Println("Error:", err)
				panic("missing one of args")
			}
		}
	}
}

func notifySubscribers(html HTML) {
	for _, ch := range subscribers {
		ch <- html
	}
}

func htmlHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	fmt.Fprintf(w, string(htmlContent)) // Serve the HTML content directly
	mu.Unlock()
}
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
func startServer() {
	http.HandleFunc("/ws", serveWs)
	http.HandleFunc("/html", htmlHandler) // Serve HTML content at /html
	http.HandleFunc("/", serveTemplate)   // Serve the HTML template at /
	http.ListenAndServe(":8080", nil)

}

func convertToHtml(mjml string) HTML {

	fmt.Print(mjml)
	url := "https://api.mjml.io/v1/render"
	username := "f37b292d-a818-4c75-8ff5-8f8bc77b2271"
	password := "fdf40c67-c707-4053-9380-b9fad77ae8f0"

	// JSON payload
	jsonData := map[string]string{
		"mjml": mjml,
	}
	// Marshal the map into a JSON string
	jsonBytes, err := json.Marshal(jsonData)

	// Create a new request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)
	}

	// Set headers
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-APPLICATION-ID", password) // Assuming APP ID is the same as password
	req.Header.Set("X-Access-Key", password)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		panic("error sending")
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error reading response:", err)
		panic("error reading response")
	}

	// Unmarshal the JSON response
	var mjmlResponse MJMLResponse
	err = json.Unmarshal(body, &mjmlResponse)
	if err != nil {
		fmt.Println("error unmarshaling JSON: %w", err)
		panic("error unmarshelling")
	}
	return HTML(mjmlResponse.HTML)
}

func parseArgs() (Email, error) {
	// Define command-line flags
	from := flag.String("from", "", "Your email")
	password := flag.String("password", "", "Your password")
	subject := flag.String("subject", "", "Your subject")
	htmlFile := flag.String("html", "", "Path to the HTML file")
	mjmlFile := flag.String("mjml", "", "Path to the MJML file")
	watch := flag.Bool("watch", false, "If we should watch an MJML file for generating html")
	to := flag.String("to", "", "The email you want the email sent")

	// Parse the flags
	flag.Parse()
	// Check for required flags
	if *watch && (*htmlFile == "" || *mjmlFile == "") {
		log.Fatal("Watching requires a mjml file to generate from and an html file to output to")
		// TODO: improve error message
		panic("missing one of args")
	}
	if !*watch && (*from == "" || *password == "" || *htmlFile == "") {
		log.Fatal("All flags (from, password, html) are required")
		// TODO: improve error message
		panic("missing one of args")
	}

	return Email{from, password, subject, htmlFile, mjmlFile, watch, to}, nil
}

func main() {
	e, err := parseArgs()

	if *e.watch {
		go watchMJMLFile(*e.mjmlFile, *e.htmlFile)
		go startServer()
		// Print the URL for live preview
		fmt.Println("Server started! Visit http://localhost:8080/ to see the live preview.")
		select {}
		return
	}
	// from := "gavinfreeborn@gmail.com"
	// password := "ecuj jjhe plas ccly"
	// to := "gavinfreeborn@gmail.com"
	smtpServer := "smtp.gmail.com"
	port := "587"

	// Email headers and HTML content
	// subject := "Subject: Hello from Go\n"
	mime := "MIME-Version: 1.0\nContent-Type: text/html; charset=\"UTF-8\"\n\n"

	data, err := os.ReadFile(*e.htmlFile) // Read entire file
	if err != nil {
		fmt.Println("Error:", err)
		panic("error reading file")
	}
	body := data

	// TODO integrate mjml
	// html := convertToHtml("<mjml><mj-body><mj-container><mj-section><mj-column><mj-text>Hello World</mj-text></mj-column></mj-section></mj-container></mj-body></mjml>")

	msg := []byte("Subject: " + *e.subject + "\n" + mime + string(body))

	// SMTP Auth
	auth := smtp.PlainAuth("", *e.from, *e.password, smtpServer)

	// Send email
	err = smtp.SendMail(smtpServer+":"+port, auth, *e.from, []string{*e.to}, msg)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Email sent successfully!")
}
