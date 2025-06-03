package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	email, err := parseArgs()
	if err != nil {
		fmt.Println("Error parsing arguments:", err)
		return
	}

	if *email.watch {
		// Convert MJML to HTML immediately on startup
		data, err := os.ReadFile(*email.mjmlFile)
		if err != nil {
			fmt.Println("Error reading MJML file:", err)
			return
		}
		html := convertToHtml(string(data))

		// Save the generated HTML to a file
		err = os.WriteFile(*email.htmlFile, []byte(html), 0644)
		if err != nil {
			fmt.Println("Error saving HTML to file:", err)
			return
		}

		htmlContent = html // Update the htmlContent variable

		go watchMJMLFile(*email.mjmlFile, *email.htmlFile)
		go startServer()
		// Print the URL for live preview
		fmt.Println("Server started! Visit http://localhost:8080/ to see the live preview.")
		select {} // Keep the program running
	}

	// If not in watch mode, send email
	err = sendEmail(email)
	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}
	fmt.Println("Email sent successfully!")
}

func parseArgs() (Email, error) {
	// Define command-line flags
	from := flag.String("from", "", "Your email address (e.g., -from \"tmp@tmp.com\")")
	password := flag.String("password", "", "Your email password (e.g., -password \"XYZ\")")
	subject := flag.String("subject", "", "Subject of the email (e.g., -subject \"Thing I Worked On\")")
	htmlFile := flag.String("html", "", "Path to the HTML file (e.g., -html index.html)")
	mjmlFile := flag.String("mjml", "", "Path to the MJML file (e.g., -mjml index.mjml)")
	watch := flag.Bool("watch", false, "Watch an MJML file for changes and generate HTML (e.g., -watch -mjml index.mjml -html index.html)")
	to := flag.String("to", "", "The recipient email address (e.g., -to \"tmp2@tmp.com\")")

	// Parse the flags
	flag.Parse()

	// Check for required flags based on mode
	if *watch {
		if *mjmlFile == "" {
			return Email{}, fmt.Errorf("missing required flag -mjml. You must specify the MJML file to watch (e.g., -watch -mjml index.mjml -html index.html)")
		}
		if *htmlFile == "" {
			return Email{}, fmt.Errorf("missing required flag -html. You must specify the HTML file to save the generated content (e.g., -watch -mjml index.mjml -html index.html)")
		}
	} else {
		if *from == "" {
			return Email{}, fmt.Errorf("missing required flag -from. You must provide your email address for sending (e.g., -from \"tmp@tmp.com\")")
		}
		if *password == "" {
			return Email{}, fmt.Errorf("missing required flag -password. You must provide your email password (e.g., -password \"XYZ\")")
		}
		if *htmlFile == "" {
			return Email{}, fmt.Errorf("missing required flag -html. You must specify the HTML file to send (e.g., -html index.html)")
		}
		if *to == "" {
			return Email{}, fmt.Errorf("missing required flag -to. You must specify the recipient email address (e.g., -to \"tmp2@tmp.com\")")
		}
	}

	return Email{from, password, subject, htmlFile, mjmlFile, watch, to}, nil
}
