package main

import (
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"os"
)

type Email struct {
	from     *string
	password *string
	subject  *string
	htmlFile *string
	to       *string
}

func parseArgs() (Email, error) {
	// Define command-line flags
	from := flag.String("from", "", "Your email")
	password := flag.String("password", "", "Your password")
	subject := flag.String("subject", "", "Your subject")
	htmlFile := flag.String("html", "", "Path to the HTML file")
	to := flag.String("to", "", "The email you want the email sent")

	// Parse the flags
	flag.Parse()
	// Check for required flags
	if *from == "" || *password == "" || *htmlFile == "" {
		log.Fatal("All flags (from, password, html) are required")
		// TODO: improve error message
		panic("missing one of args")
	}

	return Email{from, password, subject, htmlFile, to}, nil
}

func main() {
	e, err := parseArgs()

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
