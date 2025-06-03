package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// convertToHtml converts MJML markup to HTML using the MJML API
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
