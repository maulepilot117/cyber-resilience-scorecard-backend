package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jordan-wright/email"
)

// RequestData holds the data sent from the frontend
type RequestData struct {
	Email       string `json:"email"`
	HTMLContent string `json:"htmlContent"`
}

func generatePDFHandler(w http.ResponseWriter, r *http.Request) {
	var data RequestData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate unique PDF filename
	pdfFileName := uuid.New().String() + ".pdf"
	pdfPath := filepath.Join("temp", pdfFileName)

	// Ensure temp directory exists
	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		os.Mkdir("temp", 0755)
	}

	// Run wkhtmltopdf to generate PDF from HTML content
	cmd := exec.Command("wkhtmltopdf", "-", pdfPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Println("Error creating stdin pipe:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	go func() {
		defer stdin.Close()
		fmt.Fprint(stdin, data.HTMLContent)
	}()

	err = cmd.Run()
	if err != nil {
		log.Println("Error running wkhtmltopdf:", err)
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}

	// Send email with PDF attachment
	err = sendEmail(data.Email, pdfPath)
	if err != nil {
		log.Println("Error sending email:", err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	// Delete temporary PDF file
	os.Remove(pdfPath)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "PDF generated and emailed successfully"})
}

func sendEmail(toEmail, pdfPath string) error {
	// Load SMTP configuration from environment variables
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("FROM_EMAIL")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || fromEmail == "" {
		return fmt.Errorf("missing SMTP configuration in environment variables")
	}

	// Create new email
	e := email.NewEmail()
	e.From = fromEmail
	e.To = []string{toEmail}
	e.Subject = "Your Results PDF"
	e.Text = []byte("Attached is your results PDF. Enjoy!")

	// Attach the PDF
	_ , err := e.AttachFile(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to attach PDF: %w", err)
	}

	// Send the email
	err = e.Send(smtpHost+":"+smtpPort, smtp.PlainAuth("", smtpUser, smtpPass, smtpHost))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func main() {
	// Set up HTTP server
	http.HandleFunc("/generate-pdf", generatePDFHandler)

	// Start server
	port := "3000"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}