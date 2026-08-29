package services

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/ysnarafat/tenantly/internal/models"
)

// Brand/status colors mirrored from the frontend's CSS variables
// (--color-primary, --color-paid, --color-due, --color-pending,
// --color-overdue in payment-list.scss) so the receipt matches the app.
var (
	receiptBrandColor    = [3]int{30, 136, 229}  // #1E88E5
	receiptBrandTintText = [3]int{214, 234, 253} // header subtitle — brand color at low contrast against white
	receiptDarkText      = [3]int{26, 26, 26}    // #1a1a1a
	receiptMutedText     = [3]int{102, 102, 102} // #666666
	receiptZebraFill     = [3]int{245, 247, 250} // light gray-blue
	receiptBorderColor   = [3]int{224, 224, 224} // #e0e0e0
	receiptBalanceColor  = [3]int{244, 67, 54}   // #f44336 — any unpaid balance, regardless of status
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

// formatBDT renders an amount using Bangladeshi digit grouping (lakh/crore —
// e.g. 1,234,567.89 is written 12,34,567.89), matching how tenants actually
// read currency amounts rather than the international 3-digit grouping.
func formatBDT(amount float64) string {
	neg := amount < 0
	if neg {
		amount = -amount
	}

	whole := fmt.Sprintf("%.2f", amount)
	intPart, decPart, _ := strings.Cut(whole, ".")

	grouped := intPart
	if len(intPart) > 3 {
		head, tail := intPart[:len(intPart)-3], intPart[len(intPart)-3:]
		var segments []string
		for len(head) > 2 {
			segments = append([]string{head[len(head)-2:]}, segments...)
			head = head[:len(head)-2]
		}
		if head != "" {
			segments = append([]string{head}, segments...)
		}
		grouped = strings.Join(append(segments, tail), ",")
	}

	result := "BDT " + grouped + "." + decPart
	if neg {
		result = "-" + result
	}
	return result
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

	statusColor := receiptStatusColor(payment.Status)

	// ── Header band ──────────────────────────────────────────────────────
	headerH := 36.0
	pdf.SetFillColor(receiptBrandColor[0], receiptBrandColor[1], receiptBrandColor[2])
	pdf.Rect(0, 0, pageW, headerH, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(marginL, 10)
	pdf.SetFont("Helvetica", "B", 22)
	pdf.CellFormat(contentW/2, 9, "Tenantly", "", 1, "L", false, 0, "")

	pdf.SetTextColor(receiptBrandTintText[0], receiptBrandTintText[1], receiptBrandTintText[2])
	pdf.SetXY(marginL, 21)
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(contentW/2, 6, "Payment Receipt", "", 1, "L", false, 0, "")

	rightColW := 75.0
	rightX := pageW - marginR - rightColW
	pdf.SetTextColor(receiptBrandTintText[0], receiptBrandTintText[1], receiptBrandTintText[2])
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetXY(rightX, 10)
	pdf.CellFormat(rightColW, 5, "RECEIPT NO.", "", 1, "R", false, 0, "")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetXY(rightX, 15)
	pdf.CellFormat(rightColW, 7, payment.ReceiptNumber, "", 1, "R", false, 0, "")

	dateStr := payment.CreatedAt.Format("02 Jan 2006")
	if payment.PaymentDate != nil {
		dateStr = payment.PaymentDate.Format("02 Jan 2006")
	}
	pdf.SetTextColor(receiptBrandTintText[0], receiptBrandTintText[1], receiptBrandTintText[2])
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetXY(rightX, 24)
	pdf.CellFormat(rightColW, 5, dateStr, "", 1, "R", false, 0, "")

	// ── Billed To / Summary ──────────────────────────────────────────────
	y := headerH + 10
	leftColW := contentW * 0.55
	rightSummaryX := marginL + leftColW

	pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetXY(marginL, y)
	pdf.CellFormat(leftColW, 4, "BILLED TO", "", 0, "L", false, 0, "")
	pdf.SetXY(rightSummaryX, y)
	pdf.CellFormat(contentW-leftColW, 4, "SUMMARY", "", 1, "L", false, 0, "")

	pdf.SetTextColor(receiptDarkText[0], receiptDarkText[1], receiptDarkText[2])
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetXY(marginL, y+6)
	pdf.CellFormat(leftColW, 7, payment.TenantName, "", 0, "L", false, 0, "")

	badgeW, badgeH := 26.0, 7.0
	pdf.SetFillColor(statusColor[0], statusColor[1], statusColor[2])
	pdf.RoundedRect(rightSummaryX, y+5, badgeW, badgeH, 1.5, "1234", "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetXY(rightSummaryX, y+5)
	pdf.CellFormat(badgeW, badgeH, strings.ToUpper(string(payment.Status)), "", 0, "C", false, 0, "")

	pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetXY(rightSummaryX+badgeW+4, y+5)
	pdf.CellFormat(contentW-leftColW-badgeW-4, badgeH, fmt.Sprintf("Period %02d/%d", payment.Month, payment.Year), "", 1, "L", false, 0, "")

	pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetXY(marginL, y+14)
	pdf.CellFormat(leftColW, 6, fmt.Sprintf("%s, %s (%s), Unit %s", payment.PropertyName, payment.BuildingName, payment.BuildingCode, payment.UnitNumber), "", 0, "L", false, 0, "")

	if payment.PaymentMethod != "" {
		pdf.SetXY(rightSummaryX, y+16)
		pdf.CellFormat(contentW-leftColW, 6, "Paid via "+payment.PaymentMethod, "", 0, "L", false, 0, "")
	}

	y += 28

	// ── Divider ──────────────────────────────────────────────────────────
	pdf.SetDrawColor(receiptBorderColor[0], receiptBorderColor[1], receiptBorderColor[2])
	pdf.SetLineWidth(0.2)
	pdf.Line(marginL, y, pageW-marginR, y)
	y += 6

	// ── Itemized summary table ───────────────────────────────────────────
	amountColW := 45.0
	descColW := contentW - amountColW
	rowH := 9.0

	pdf.SetFillColor(receiptZebraFill[0], receiptZebraFill[1], receiptZebraFill[2])
	pdf.Rect(marginL, y, contentW, rowH, "F")
	pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetXY(marginL+3, y)
	pdf.CellFormat(descColW-3, rowH, "DESCRIPTION", "", 0, "L", false, 0, "")
	pdf.SetXY(marginL+descColW, y)
	pdf.CellFormat(amountColW-3, rowH, "AMOUNT", "", 0, "R", false, 0, "")
	y += rowH

	drawRow := func(label string, amount float64, bold bool, color [3]int) {
		pdf.SetXY(marginL+3, y)
		if bold {
			pdf.SetFont("Helvetica", "B", 10)
		} else {
			pdf.SetFont("Helvetica", "", 10)
		}
		pdf.SetTextColor(receiptDarkText[0], receiptDarkText[1], receiptDarkText[2])
		pdf.CellFormat(descColW-3, rowH, label, "", 0, "L", false, 0, "")
		pdf.SetXY(marginL+descColW, y)
		pdf.SetTextColor(color[0], color[1], color[2])
		pdf.CellFormat(amountColW-3, rowH, formatBDT(amount), "", 0, "R", false, 0, "")
		y += rowH
		pdf.SetDrawColor(receiptBorderColor[0], receiptBorderColor[1], receiptBorderColor[2])
		pdf.Line(marginL, y, pageW-marginR, y)
	}

	drawRow(fmt.Sprintf("Rent & charges due for %02d/%d", payment.Month, payment.Year), payment.AmountDue, false, receiptDarkText)
	drawRow("Amount Paid", payment.AmountPaid, true, statusColor)

	balance := payment.AmountDue - payment.AmountPaid
	if balance > 0.005 {
		drawRow("Balance Due", balance, true, receiptBalanceColor)
	}

	y += 8

	// ── Notes ────────────────────────────────────────────────────────────
	if payment.Notes != "" {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
		pdf.SetXY(marginL, y)
		pdf.CellFormat(contentW, 5, "NOTES", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(receiptDarkText[0], receiptDarkText[1], receiptDarkText[2])
		pdf.SetXY(marginL, y+5)
		pdf.MultiCell(contentW, 6, payment.Notes, "", "L", false)
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
