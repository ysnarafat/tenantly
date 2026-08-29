package services

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"
	"github.com/ysnarafat/tenantly/internal/models"
)

// Brand/status colors mirrored from the frontend's CSS variables
// (--color-primary, --color-paid, --color-due, --color-pending,
// --color-overdue in payment-list.scss) so the receipt matches the app.
var (
	receiptBrandColor  = [3]int{30, 136, 229}  // #1E88E5
	receiptDarkText    = [3]int{26, 26, 26}    // #1a1a1a
	receiptMutedText   = [3]int{102, 102, 102} // #666666
	receiptZebraFill   = [3]int{245, 247, 250} // light gray-blue
	receiptBorderColor = [3]int{224, 224, 224} // #e0e0e0
)

func receiptStatusColor(status models.PaymentStatus) [3]int {
	switch status {
	case models.PaymentStatusPaid:
		return [3]int{76, 175, 80} // #4caf50
	case models.PaymentStatusPartial:
		return [3]int{255, 152, 0} // #ff9800
	case models.PaymentStatusOverdue:
		return [3]int{244, 67, 54} // #f44336
	default: // Due
		return [3]int{144, 164, 174} // #90a4ae
	}
}

// GenerateReceiptPDF renders a one-page payment receipt as PDF bytes for a
// staff member with access to the org. It is deliberately decoupled from the
// HTTP layer so future email/SMS delivery jobs can call it directly without
// going through the download endpoint.
func (s *PaymentService) GenerateReceiptPDF(paymentID, orgID int) ([]byte, error) {
	payment, err := s.GetPayment(paymentID, orgID)
	if err != nil {
		return nil, err
	}
	return renderReceiptPDF(payment)
}

func renderReceiptPDF(payment *models.PaymentWithDetails) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(fmt.Sprintf("Receipt %s", payment.ReceiptNumber), false)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	pageW, _ := pdf.GetPageSize()
	marginL, _, marginR, _ := pdf.GetMargins()
	contentW := pageW - marginL - marginR

	// ── Header band ──────────────────────────────────────────────────────
	headerH := 32.0
	pdf.SetFillColor(receiptBrandColor[0], receiptBrandColor[1], receiptBrandColor[2])
	pdf.Rect(0, 0, pageW, headerH, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(marginL, 8)
	pdf.SetFont("Helvetica", "B", 20)
	pdf.CellFormat(contentW, 10, "Tenantly", "", 1, "L", false, 0, "")

	pdf.SetXY(marginL, 18)
	pdf.SetFont("Helvetica", "", 12)
	pdf.CellFormat(contentW, 7, "Payment Receipt", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetXY(marginL, 8)
	pdf.CellFormat(contentW, 6, payment.ReceiptNumber, "", 1, "R", false, 0, "")
	if payment.PaymentDate != nil {
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetXY(marginL, 15)
		pdf.CellFormat(contentW, 6, payment.PaymentDate.Format("02 Jan 2006"), "", 1, "R", false, 0, "")
	}

	y := headerH + 8

	// ── Status badge ─────────────────────────────────────────────────────
	statusColor := receiptStatusColor(payment.Status)
	pdf.SetFillColor(statusColor[0], statusColor[1], statusColor[2])
	badgeW := 32.0
	pdf.RoundedRect(marginL, y, badgeW, 9, 2, "1234", "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetXY(marginL, y)
	pdf.CellFormat(badgeW, 9, string(payment.Status), "", 0, "C", false, 0, "")

	pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetXY(marginL+badgeW+4, y)
	pdf.CellFormat(contentW-badgeW-4, 9, fmt.Sprintf("Period: %02d/%d", payment.Month, payment.Year), "", 1, "L", false, 0, "")

	y += 16

	// ── Details table ────────────────────────────────────────────────────
	rows := [][2]string{
		{"Tenant", payment.TenantName},
		{"Property", payment.PropertyName},
		{"Building", fmt.Sprintf("%s (%s)", payment.BuildingName, payment.BuildingCode)},
		{"Unit", payment.UnitNumber},
		{"Payment Method", payment.PaymentMethod},
	}

	rowH := 9.0
	pdf.SetDrawColor(receiptBorderColor[0], receiptBorderColor[1], receiptBorderColor[2])
	pdf.SetLineWidth(0.2)
	labelW := 50.0

	for i, row := range rows {
		fill := i%2 == 0
		if fill {
			pdf.SetFillColor(receiptZebraFill[0], receiptZebraFill[1], receiptZebraFill[2])
		}
		pdf.SetXY(marginL, y)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
		pdf.CellFormat(labelW, rowH, row[0], "1", 0, "L", fill, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(receiptDarkText[0], receiptDarkText[1], receiptDarkText[2])
		pdf.CellFormat(contentW-labelW, rowH, row[1], "1", 1, "L", fill, 0, "")
		y += rowH
	}

	// ── Amount summary ───────────────────────────────────────────────────
	y += 6
	summaryH := 16.0
	pdf.SetFillColor(receiptZebraFill[0], receiptZebraFill[1], receiptZebraFill[2])
	pdf.RoundedRect(marginL, y, contentW, summaryH, 2, "1234", "F")

	colW := contentW / 2
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
	pdf.SetXY(marginL+6, y+3)
	pdf.CellFormat(colW-6, 5, "AMOUNT DUE", "", 0, "L", false, 0, "")
	pdf.SetXY(marginL+colW, y+3)
	pdf.CellFormat(colW-6, 5, "AMOUNT PAID", "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 15)
	pdf.SetTextColor(receiptDarkText[0], receiptDarkText[1], receiptDarkText[2])
	pdf.SetXY(marginL+6, y+8)
	pdf.CellFormat(colW-6, 7, fmt.Sprintf("BDT %.2f", payment.AmountDue), "", 0, "L", false, 0, "")

	paidColor := receiptStatusColor(models.PaymentStatusPaid)
	pdf.SetTextColor(paidColor[0], paidColor[1], paidColor[2])
	pdf.SetXY(marginL+colW, y+8)
	pdf.CellFormat(colW-6, 7, fmt.Sprintf("BDT %.2f", payment.AmountPaid), "", 0, "L", false, 0, "")

	y += summaryH + 8

	if payment.Notes != "" {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
		pdf.SetXY(marginL, y)
		pdf.CellFormat(labelW, 7, "Notes", "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(receiptDarkText[0], receiptDarkText[1], receiptDarkText[2])
		pdf.SetXY(marginL+labelW, y)
		pdf.MultiCell(contentW-labelW, 7, payment.Notes, "", "L", false)
		y = pdf.GetY() + 4
	}

	// ── Footer ───────────────────────────────────────────────────────────
	pdf.SetDrawColor(receiptBrandColor[0], receiptBrandColor[1], receiptBrandColor[2])
	pdf.SetLineWidth(0.6)
	footerY := y + 6
	pdf.Line(marginL, footerY, pageW-marginR, footerY)

	pdf.SetFont("Helvetica", "I", 9)
	pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
	pdf.SetXY(marginL, footerY+4)
	pdf.CellFormat(contentW, 6, "This is a system-generated receipt. Thank you for your payment.", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to render receipt PDF: %w", err)
	}
	return buf.Bytes(), nil
}
