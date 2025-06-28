package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"

	"github.com/jordan-wright/email"
)

// Recommendation struct to match the JavaScript object structure
type Recommendation struct {
	Category string `json:"category"`
	Question string `json:"question"`
	Text     string `json:"text"`
	Status   string `json:"status"`
}

// CategoryScore struct to match the JavaScript categoryScore structure
type CategoryScore struct {
	Name       string  `json:"name"`
	Score      float64 `json:"score"`
	Max        float64 `json:"max"`
	Percentage float64 `json:"percentage"`
}

// Updated RequestData with proper types matching the frontend
type RequestData struct {
	HTMLContent    string           `json:"htmlContent"`
	Email          string           `json:"email"`
	Score          int              `json:"score"`
	CategoryScores []CategoryScore  `json:"categoryScore"`    // Note: frontend sends "categoryScore" not "categoryScores"
	Recommendations []Recommendation `json:"recommendations"`
}

// Logging middleware to log incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Adjust for production
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Handler for generating PDF and sending email
func generatePDFHandler(w http.ResponseWriter, r *http.Request) {
	// Use RequestData
	var data RequestData

	// Log the request details for debugging
	log.Printf("Request method: %s, Content-Length: %d", r.Method, r.ContentLength)

	// Check if the body is empty
	if r.ContentLength == 0 {
		log.Println("Request body is empty")
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	// Decode the request body
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	// Log the received data for debugging
	log.Printf("Received data - Email: %s, Score: %d", data.Email, data.Score)
	log.Printf("Number of recommendations: %d", len(data.Recommendations))
	log.Printf("Number of category scores: %d", len(data.CategoryScores))

	// Log and generate PDF
	log.Println("Starting PDF generation")
	pdfPath, err := generatePDF(data)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}
	log.Println("PDF generated successfully")

	// Log and send email
	log.Printf("Sending email to: %s", data.Email)
	err = sendEmail(data.Email, pdfPath)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}
	log.Println("Email sent successfully")

	// Success response - send JSON response instead of plain text
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message": "PDF generated and emailed successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// SMTP configuration from environment variables
type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Pass     string
	FromEmail string
}

// loadSMTPConfig loads SMTP configuration from environment variables
func loadSMTPConfig() (*SMTPConfig, error) {
	config := &SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),
		User:     os.Getenv("SMTP_USER"),
		Pass:     os.Getenv("SMTP_PASS"),
		FromEmail: os.Getenv("FROM_EMAIL"),
	}

	// Validate required fields
	if config.Host == "" {
		return nil, fmt.Errorf("SMTP_HOST environment variable is required")
	}
	if config.FromEmail == "" {
		return nil, fmt.Errorf("FROM_EMAIL environment variable is required")
	}

	// Parse port with default
	portStr := os.Getenv("SMTP_PORT")
	if portStr == "" {
		config.Port = 587 // Default SMTP port for TLS
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid SMTP_PORT: %v", err)
		}
		config.Port = port
	}

	log.Printf("SMTP configured: host=%s, port=%d, from=%s", config.Host, config.Port, config.FromEmail)
	return config, nil
}

// Global SMTP configuration
var smtpConfig *SMTPConfig

// Email sending function
func sendEmail(toEmail, pdfPath string) error {
	log.Printf("Preparing to send email to: %s with attachment: %s", toEmail, pdfPath)
	e := email.NewEmail()
	e.From = smtpConfig.FromEmail
	e.To = []string{toEmail}
	e.Subject = "Your Cyber Resilience Scorecard Results"
	e.Text = []byte("Please find attached your Cyber Resilience Scorecard results.")
	e.HTML = []byte(`
		<h2>Your Cyber Resilience Scorecard Results</h2>
		<p>Thank you for completing the assessment. Your detailed results are attached as a PDF.</p>
		<p>If you have any questions, please don't hesitate to contact us.</p>
	`)

	_, err := e.AttachFile(pdfPath)
	if err != nil {
		log.Printf("Error attaching file: %v", err)
		return err
	}

	// Create SMTP auth
	var auth smtp.Auth
	if smtpConfig.User != "" && smtpConfig.Pass != "" {
		auth = smtp.PlainAuth("", smtpConfig.User, smtpConfig.Pass, smtpConfig.Host)
	}

	// Send email
	addr := fmt.Sprintf("%s:%d", smtpConfig.Host, smtpConfig.Port)
	err = e.Send(addr, auth)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return err
	}
	
	log.Printf("Email sent successfully to %s", toEmail)
	return nil
}

func main() {
	// Load SMTP configuration
	var err error
	smtpConfig, err = loadSMTPConfig()
	if err != nil {
		log.Fatalf("Failed to load SMTP configuration: %v", err)
	}

	// Set up handler with logging middleware
	handler := loggingMiddleware(http.HandlerFunc(generatePDFHandler))
	http.Handle("/generate-pdf", enableCors(handler))
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}