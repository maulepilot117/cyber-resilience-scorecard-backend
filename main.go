package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/smtp"

	"github.com/jordan-wright/email"
)

// Updated RequestData with all required fields
type RequestData struct {
    HTMLContent    string   `json:"htmlContent"`
    Email          string   `json:"email"`
    Score          int      `json:"score"`
    CategoryScores []int    `json:"categoryScores"`
    Recommendations []string `json:"recommendations"`
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

    // Log and generate PDF
    log.Println("Starting PDF generation")
    pdfPath, err := generatePDF(data.HTMLContent)
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

    // Success response
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("PDF generated and emailed successfully"))
}

// PDF generation function (placeholder for your logic)
func generatePDF(htmlContent string) (string, error) {
    log.Println("Generating PDF from HTML content")
    // Replace with your actual PDF generation logic
    pdfPath := "/path/to/generated.pdf" // Example path
    return pdfPath, nil
}

// Email sending function
func sendEmail(toEmail, pdfPath string) error {
    log.Printf("Preparing to send email to: %s with attachment: %s", toEmail, pdfPath)
    e := email.NewEmail()
    e.To = []string{toEmail}
    e.Subject = "Your Results PDF"
    e.Text = []byte("Attached is your results PDF.")

    _, err := e.AttachFile(pdfPath)
    if err != nil {
        log.Printf("Error attaching file: %v", err)
        return err
    }

    // Configure your SMTP settings here
    err = e.Send("smtp.example.com:587", smtp.PlainAuth("", "user", "pass", "smtp.example.com"))
    if err != nil {
        log.Printf("Error sending email: %v", err)
        return err
    }
    return nil
}

func main() {
    // Set up handler with logging middleware
    handler := loggingMiddleware(http.HandlerFunc(generatePDFHandler))
    http.Handle("/generate-pdf", enableCors(handler))
    log.Println("Server starting on port 3000")
    log.Fatal(http.ListenAndServe(":3000", nil))
}