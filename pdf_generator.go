package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// generatePDF creates a comprehensive PDF report from the scorecard data
func generatePDF(data RequestData) (string, error) {
	// Create a new PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	
	// Add a page
	pdf.AddPage()
	
	// Set up colors
	// Primary blue color for headers
	pdf.SetFillColor(59, 130, 246) // Blue-600
	pdf.SetTextColor(255, 255, 255) // White text
	
	// Add header with gradient effect
	pdf.Rect(0, 0, 210, 40, "F")
	
	// Title
	pdf.SetFont("Arial", "B", 24)
	pdf.SetXY(10, 10)
	pdf.CellFormat(190, 10, "Cyber Resilience Scorecard", "0", 1, "C", false, 0, "")
	
	// Subtitle
	pdf.SetFont("Arial", "", 12)
	pdf.SetXY(10, 22)
	pdf.CellFormat(190, 8, "Assessment Results Report", "0", 1, "C", false, 0, "")
	
	// Reset text color for body
	pdf.SetTextColor(0, 0, 0)
	
	// Add date and email info
	pdf.SetY(50)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Generated on: %s", time.Now().Format("January 2, 2006")))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Report for: %s", data.Email))
	pdf.Ln(12)
	
	// Overall Score Section
	addScoreSection(pdf, data.Score)
	
	// Category Scores Section
	pdf.Ln(10)
	addCategoryScoresSection(pdf, data.CategoryScores)
	
	// Recommendations Section
	if len(data.Recommendations) > 0 {
		pdf.AddPage()
		addRecommendationsSection(pdf, data.Recommendations)
	}
	
	// Add footer on all pages
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.SetTextColor(128, 128, 128)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pdf.PageNo()), "", 0, "C", false, 0, "")
	})
	
	// Create output directory if it doesn't exist
	outputDir := "pdf_output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Generate filename with timestamp
	filename := fmt.Sprintf("cyber_resilience_report_%s.pdf", time.Now().Format("20060102_150405"))
	outputPath := filepath.Join(outputDir, filename)
	
	// Save the PDF
	err := pdf.OutputFileAndClose(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to save PDF: %w", err)
	}
	
	log.Printf("PDF generated successfully: %s", outputPath)
	return outputPath, nil
}

// addScoreSection adds the overall score display
func addScoreSection(pdf *gofpdf.Fpdf, score int) {
	// Score header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Overall Resilience Score")
	pdf.Ln(12)
	
	// Create a visual score representation
	scoreColor := getScoreColor(score)
	pdf.SetFillColor(scoreColor.R, scoreColor.G, scoreColor.B)
	
	// Score box
	boxWidth := 100.0
	boxHeight := 40.0
	boxX := (210 - boxWidth) / 2 // Center the box
	
	// Draw score background
	pdf.RoundedRect(boxX, pdf.GetY(), boxWidth, boxHeight, 3, "1234", "F")
	
	// Add score text
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 36)
	pdf.SetXY(boxX, pdf.GetY()+8)
	pdf.CellFormat(boxWidth, 20, fmt.Sprintf("%d%%", score), "0", 1, "C", false, 0, "")
	
	// Add score interpretation
	pdf.SetY(pdf.GetY() + boxHeight - 28)
	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(0, 0, 0)
	interpretation := getScoreInterpretation(score)
	pdf.Cell(0, 8, interpretation)
	pdf.Ln(8)
}

// addCategoryScoresSection adds the category breakdown
func addCategoryScoresSection(pdf *gofpdf.Fpdf, categories []CategoryScore) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Category Breakdown")
	pdf.Ln(10)
	
	// Table header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(80, 8, "Category", "1", 0, "L", true, 0, "")
	pdf.CellFormat(30, 8, "Score", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, "Maximum", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 8, "Percentage", "1", 0, "C", true, 0, "")
	pdf.Ln(8)
	
	// Table rows
	pdf.SetFont("Arial", "", 10)
	for i, category := range categories {
		// Alternate row colors
		if i%2 == 0 {
			pdf.SetFillColor(250, 250, 250)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		
		pdf.CellFormat(80, 8, category.Name, "1", 0, "L", true, 0, "")
		pdf.CellFormat(30, 8, fmt.Sprintf("%.1f", category.Score), "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, 8, fmt.Sprintf("%.1f", category.Max), "1", 0, "C", true, 0, "")
		
		// Add visual percentage bar
		barX := pdf.GetX()
		barY := pdf.GetY()
		
		// Draw percentage text
		pdf.CellFormat(50, 8, fmt.Sprintf("%.1f%%", category.Percentage), "1", 0, "C", true, 0, "")
		
		// Draw mini progress bar
		barWidth := 40.0
		barHeight := 4.0
		barStartX := barX + 5
		barStartY := barY + 2
		
		// Background
		pdf.SetFillColor(230, 230, 230)
		pdf.Rect(barStartX, barStartY, barWidth, barHeight, "F")
		
		// Progress
		scoreColor := getScoreColor(int(category.Percentage))
		pdf.SetFillColor(scoreColor.R, scoreColor.G, scoreColor.B)
		pdf.Rect(barStartX, barStartY, barWidth*(category.Percentage/100), barHeight, "F")
		
		pdf.Ln(8)
	}
}

// addRecommendationsSection adds the recommendations
func addRecommendationsSection(pdf *gofpdf.Fpdf, recommendations []Recommendation) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Recommendations for Improvement")
	pdf.Ln(10)
	
	// Group recommendations by status
	var criticalRecs, partialRecs []Recommendation
	for _, rec := range recommendations {
		if rec.Status == "missing" {
			criticalRecs = append(criticalRecs, rec)
		} else if rec.Status == "partial" {
			partialRecs = append(partialRecs, rec)
		}
	}
	
	// Critical improvements (missing controls)
	if len(criticalRecs) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.SetTextColor(220, 38, 38) // Red color
		pdf.Cell(0, 8, "Critical Improvements Needed")
		pdf.SetTextColor(0, 0, 0)
		pdf.Ln(8)
		
		addRecommendationsList(pdf, criticalRecs, 220, 38, 38) // Red
		pdf.Ln(8)
	}
	
	// Partial improvements
	if len(partialRecs) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.SetTextColor(245, 158, 11) // Amber color
		pdf.Cell(0, 8, "Areas for Enhancement")
		pdf.SetTextColor(0, 0, 0)
		pdf.Ln(8)
		
		addRecommendationsList(pdf, partialRecs, 245, 158, 11) // Amber
	}
}

// addRecommendationsList adds a list of recommendations with bullets
func addRecommendationsList(pdf *gofpdf.Fpdf, recs []Recommendation, r, g, b int) {
	pdf.SetFont("Arial", "", 10)
	
	// Group by category
	categoryMap := make(map[string][]Recommendation)
	for _, rec := range recs {
		categoryMap[rec.Category] = append(categoryMap[rec.Category], rec)
	}
	
	for category, catRecs := range categoryMap {
		// Category name
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 6, category)
		pdf.Ln(6)
		pdf.SetFont("Arial", "", 10)
		
		// Recommendations
		for _, rec := range catRecs {
			// Check if we need a new page
			if pdf.GetY() > 250 {
				pdf.AddPage()
			}
			
			// Bullet point
			pdf.SetX(15)
			pdf.SetFillColor(r, g, b)
			pdf.Circle(pdf.GetX()+2, pdf.GetY()+2.5, 1.5, "F")
			
			// Text with wrapping
			pdf.SetX(20)
			pdf.MultiCell(170, 5, rec.Text, "", "L", false)
			pdf.Ln(2)
		}
		pdf.Ln(4)
	}
}

// Helper structures for colors
type RGB struct {
	R, G, B int
}

// getScoreColor returns appropriate color based on score
func getScoreColor(score int) RGB {
	switch {
	case score >= 80:
		return RGB{34, 197, 94} // Green
	case score >= 60:
		return RGB{251, 191, 36} // Amber
	case score >= 40:
		return RGB{251, 146, 60} // Orange
	default:
		return RGB{239, 68, 68} // Red
	}
}

// getScoreInterpretation returns text interpretation of the score
func getScoreInterpretation(score int) string {
	switch {
	case score >= 80:
		return "Excellent - Your organization demonstrates strong cyber resilience"
	case score >= 60:
		return "Good - Your organization has solid foundations with room for improvement"
	case score >= 40:
		return "Fair - Several areas require attention to improve resilience"
	default:
		return "Needs Improvement - Significant gaps identified in cyber resilience"
	}
}