package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

// watchMJMLFile monitors the MJML file for changes and updates HTML content
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
