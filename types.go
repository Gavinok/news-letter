package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

// MJMLResponse represents the response from the MJML API
type MJMLResponse struct {
	HTML string `json:"html"`
}

// Email represents the email configuration
type Email struct {
	from     *string
	password *string
	subject  *string
	htmlFile *string
	mjmlFile *string
	watch    *bool
	to       *string
}

// HTML represents HTML content as a string
type HTML string

// Global variables
var upgrader = websocket.Upgrader{}
var clients = make(map[*websocket.Conn]bool)
var htmlContent = HTML("")
var mu sync.Mutex
var subscribers = make([]chan HTML, 0)
