package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	email, err := parseArgs()
	if err != nil {
		log.Fatal("Error parsing arguments:", err)
	}

	if *email.watch {
		// Convert MJML to HTML immediately on startup
		data, err := os.ReadFile(*email.mjmlFile)
		if err != nil {
			log.Fatal("Error reading MJML file:", err)
		}
		html := convertToHtml(string(data))

		// Save the generated HTML to a file
		err = os.WriteFile(*email.htmlFile, []byte(html), 0644)
		if err != nil {
			log.Fatal("Error saving HTML to file:", err)
		}

		htmlContent = html // Update the htmlContent variable

		go watchMJMLFile(*email.mjmlFile, *email.htmlFile)
		go startServer()
		// Print the URL for live preview
		fmt.Println("Server started! Visit http://localhost:8080/ to see the live preview.")
		select {} // Keep the program running
		return
	}

	// If not in watch mode, send email
	err = sendEmail(email)
	if err != nil {
		log.Fatal("Error sending email:", err)
	}
	fmt.Println("Email sent successfully!")
}

// parseArgs parses command-line arguments
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
		return Email{}, fmt.Errorf("watching requires a mjml file to generate from and an html file to output to")
	}

	if !*watch && (*from == "" || *password == "" || *htmlFile == "") {
		return Email{}, fmt.Errorf("all flags (from, password, html) are required for sending email")
	}

	return Email{from, password, subject, htmlFile, mjmlFile, watch, to}, nil
}
