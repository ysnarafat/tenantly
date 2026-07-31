package services

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"
)

// GenerateReceiptPDF renders a one-page payment receipt as PDF bytes. It is
// deliberately decoupled from the HTTP layer so future email/SMS delivery
// jobs can call it directly without going through the download endpoint.
func (s *PaymentService) GenerateReceiptPDF(paymentID, orgID int) ([]byte, error) {
	payment, err := s.GetPayment(paymentID, orgID)
	if err != nil {
		return nil, err
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(fmt.Sprintf("Receipt %s", payment.ReceiptNumber), false)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 10, "Payment Receipt", "", 1, "C", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 7, fmt.Sprintf("Receipt No: %s", payment.ReceiptNumber), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 7, fmt.Sprintf("Status: %s", payment.Status), "", 1, "L", false, 0, "")
	pdf.Ln(4)

	rows := [][2]string{
		{"Tenant", payment.TenantName},
		{"Property", payment.PropertyName},
		{"Building", fmt.Sprintf("%s (%s)", payment.BuildingName, payment.BuildingCode)},
		{"Unit", payment.UnitNumber},
		{"Period", fmt.Sprintf("%02d/%d", payment.Month, payment.Year)},
		{"Amount Due", fmt.Sprintf("BDT %.2f", payment.AmountDue)},
		{"Amount Paid", fmt.Sprintf("BDT %.2f", payment.AmountPaid)},
		{"Payment Method", payment.PaymentMethod},
	}
	if payment.PaymentDate != nil {
		rows = append(rows, [2]string{"Payment Date", payment.PaymentDate.Format("2006-01-02")})
	}
	if payment.Notes != "" {
		rows = append(rows, [2]string{"Notes", payment.Notes})
	}

	pdf.SetFont("Helvetica", "B", 11)
	for _, row := range rows {
		pdf.CellFormat(45, 8, row[0], "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 11)
		pdf.CellFormat(0, 8, row[1], "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "B", 11)
	}

	pdf.Ln(10)
	pdf.SetFont("Helvetica", "I", 9)
	pdf.CellFormat(0, 6, "This is a system-generated receipt.", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to render receipt PDF: %w", err)
	}
	return buf.Bytes(), nil
}
