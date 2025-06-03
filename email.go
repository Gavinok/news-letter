package main

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

// sendEmail sends an email with HTML content
func sendEmail(email Email) error {
	smtpServer := "smtp.gmail.com"
	port := "587"

	// Email headers and HTML content
	mime := "MIME-Version: 1.0\nContent-Type: text/html; charset=\"UTF-8\"\n\n"

	data, err := os.ReadFile(*email.htmlFile) // Read entire file
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	body := data

	msg := []byte("Subject: " + *email.subject + "\n" + mime + string(body))

	// SMTP Auth
	auth := smtp.PlainAuth("", *email.from, *email.password, smtpServer)

	// Send email
	err = smtp.SendMail(smtpServer+":"+port, auth, *email.from, strings.Split(*email.to, ","), msg)
	if err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	return nil
}
