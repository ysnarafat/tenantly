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
	return newReceiptBuilder(payment).
		Header().
		BilledToSummary().
		Divider().
		ItemizedTable().
		Notes().
		Footer().
		Build()
}

// receiptBuilder assembles a receipt PDF one section at a time. Each section
// method reads/advances b.y (the running vertical cursor) so sections never
// need to know each other's heights up front — a section whose content grows
// (e.g. BilledToSummary wrapping a long address, or Notes running long)
// simply pushes b.y further down, and whatever comes next just continues
// from there instead of colliding with it.
type receiptBuilder struct {
	pdf      *fpdf.Fpdf
	payment  *models.PaymentWithDetails
	pageW    float64
	marginL  float64
	marginR  float64
	contentW float64
	y        float64
}

func newReceiptBuilder(payment *models.PaymentWithDetails) *receiptBuilder {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(fmt.Sprintf("Receipt %s", payment.ReceiptNumber), false)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	pageW, _ := pdf.GetPageSize()
	marginL, _, marginR, _ := pdf.GetMargins()

	return &receiptBuilder{
		pdf:      pdf,
		payment:  payment,
		pageW:    pageW,
		marginL:  marginL,
		marginR:  marginR,
		contentW: pageW - marginL - marginR,
	}
}

// Header draws the brand band with the receipt number and date right-aligned.
func (b *receiptBuilder) Header() *receiptBuilder {
	pdf, payment := b.pdf, b.payment
	headerH := 36.0

	pdf.SetFillColor(receiptBrandColor[0], receiptBrandColor[1], receiptBrandColor[2])
	pdf.Rect(0, 0, b.pageW, headerH, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(b.marginL, 10)
	pdf.SetFont("Helvetica", "B", 22)
	pdf.CellFormat(b.contentW/2, 9, "Tenantly", "", 1, "L", false, 0, "")

	pdf.SetTextColor(receiptBrandTintText[0], receiptBrandTintText[1], receiptBrandTintText[2])
	pdf.SetXY(b.marginL, 21)
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(b.contentW/2, 6, "Payment Receipt", "", 1, "L", false, 0, "")

	metaColW := 75.0
	metaX := b.pageW - b.marginR - metaColW
	pdf.SetTextColor(receiptBrandTintText[0], receiptBrandTintText[1], receiptBrandTintText[2])
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetXY(metaX, 10)
	pdf.CellFormat(metaColW, 5, "RECEIPT NO.", "", 1, "R", false, 0, "")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetXY(metaX, 15)
	pdf.CellFormat(metaColW, 7, payment.ReceiptNumber, "", 1, "R", false, 0, "")

	dateStr := payment.CreatedAt.Format("02 Jan 2006")
	if payment.PaymentDate != nil {
		dateStr = payment.PaymentDate.Format("02 Jan 2006")
	}
	pdf.SetTextColor(receiptBrandTintText[0], receiptBrandTintText[1], receiptBrandTintText[2])
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetXY(metaX, 24)
	pdf.CellFormat(metaColW, 5, dateStr, "", 1, "R", false, 0, "")

	b.y = headerH + 10
	return b
}

// BilledToSummary draws the two-column tenant/address block on the left and
// the status badge/period/payment-method block on the right. Both columns
// wrap or grow independently; b.y ends up at whichever column ran longer, so
// a long property/building/unit name can never overlap the summary column
// (CellFormat neither wraps nor clips — it just draws past its stated width
// if given text wider than that — which is what caused that overlap before).
func (b *receiptBuilder) BilledToSummary() *receiptBuilder {
	pdf, payment := b.pdf, b.payment
	y := b.y
	statusColor := receiptStatusColor(payment.Status)

	leftColW := b.contentW * 0.55
	summaryX := b.marginL + leftColW
	summaryColW := b.contentW - leftColW
	gutter := 8.0
	leftTextW := leftColW - gutter

	pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetXY(b.marginL, y)
	pdf.CellFormat(leftTextW, 4, "BILLED TO", "", 0, "L", false, 0, "")
	pdf.SetXY(summaryX, y)
	pdf.CellFormat(summaryColW, 4, "SUMMARY", "", 1, "L", false, 0, "")

	pdf.SetTextColor(receiptDarkText[0], receiptDarkText[1], receiptDarkText[2])
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetXY(b.marginL, y+6)
	pdf.CellFormat(leftTextW, 7, payment.TenantName, "", 0, "L", false, 0, "")

	badgeW, badgeH := 26.0, 7.0
	pdf.SetFillColor(statusColor[0], statusColor[1], statusColor[2])
	pdf.RoundedRect(summaryX, y+5, badgeW, badgeH, 1.5, "1234", "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetXY(summaryX, y+5)
	pdf.CellFormat(badgeW, badgeH, strings.ToUpper(string(payment.Status)), "", 0, "C", false, 0, "")

	pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetXY(summaryX+badgeW+4, y+5)
	pdf.CellFormat(summaryColW-badgeW-4, badgeH, fmt.Sprintf("Period %02d/%d", payment.Month, payment.Year), "", 1, "L", false, 0, "")

	rightY := y + 5 + badgeH + 3
	if payment.PaymentMethod != "" {
		pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetXY(summaryX, rightY)
		pdf.CellFormat(summaryColW, 6, "Paid via "+payment.PaymentMethod, "", 1, "L", false, 0, "")
		rightY = pdf.GetY()
	}

	pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetXY(b.marginL, y+14)
	pdf.MultiCell(leftTextW, 5, fmt.Sprintf("%s, %s (%s), Unit %s", payment.PropertyName, payment.BuildingName, payment.BuildingCode, payment.UnitNumber), "", "L", false)
	leftY := pdf.GetY()

	b.y = leftY
	if rightY > b.y {
		b.y = rightY
	}
	b.y += 6
	return b
}

// Divider draws a thin full-width rule and advances past it.
func (b *receiptBuilder) Divider() *receiptBuilder {
	b.pdf.SetDrawColor(receiptBorderColor[0], receiptBorderColor[1], receiptBorderColor[2])
	b.pdf.SetLineWidth(0.2)
	b.pdf.Line(b.marginL, b.y, b.pageW-b.marginR, b.y)
	b.y += 6
	return b
}

// ItemizedTable draws the amount-due/amount-paid/balance-due breakdown as a
// simple bordered table with a header row.
func (b *receiptBuilder) ItemizedTable() *receiptBuilder {
	pdf, payment := b.pdf, b.payment
	statusColor := receiptStatusColor(payment.Status)

	amountColW := 45.0
	descColW := b.contentW - amountColW
	rowH := 9.0

	pdf.SetFillColor(receiptZebraFill[0], receiptZebraFill[1], receiptZebraFill[2])
	pdf.Rect(b.marginL, b.y, b.contentW, rowH, "F")
	pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetXY(b.marginL+3, b.y)
	pdf.CellFormat(descColW-3, rowH, "DESCRIPTION", "", 0, "L", false, 0, "")
	pdf.SetXY(b.marginL+descColW, b.y)
	pdf.CellFormat(amountColW-3, rowH, "AMOUNT", "", 0, "R", false, 0, "")
	b.y += rowH

	drawRow := func(label string, amount float64, bold bool, color [3]int) {
		pdf.SetXY(b.marginL+3, b.y)
		if bold {
			pdf.SetFont("Helvetica", "B", 10)
		} else {
			pdf.SetFont("Helvetica", "", 10)
		}
		pdf.SetTextColor(receiptDarkText[0], receiptDarkText[1], receiptDarkText[2])
		pdf.CellFormat(descColW-3, rowH, label, "", 0, "L", false, 0, "")
		pdf.SetXY(b.marginL+descColW, b.y)
		pdf.SetTextColor(color[0], color[1], color[2])
		pdf.CellFormat(amountColW-3, rowH, formatBDT(amount), "", 0, "R", false, 0, "")
		b.y += rowH
		pdf.SetDrawColor(receiptBorderColor[0], receiptBorderColor[1], receiptBorderColor[2])
		pdf.Line(b.marginL, b.y, b.pageW-b.marginR, b.y)
	}

	drawRow(fmt.Sprintf("Rent & charges due for %02d/%d", payment.Month, payment.Year), payment.AmountDue, false, receiptDarkText)
	drawRow("Amount Paid", payment.AmountPaid, true, statusColor)

	if balance := payment.AmountDue - payment.AmountPaid; balance > 0.005 {
		drawRow("Balance Due", balance, true, receiptBalanceColor)
	}

	b.y += 8
	return b
}

// Notes draws the free-text notes block, wrapped, when present. A no-op
// (advances nothing) when there are no notes.
func (b *receiptBuilder) Notes() *receiptBuilder {
	if b.payment.Notes == "" {
		return b
	}

	pdf := b.pdf
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
	pdf.SetXY(b.marginL, b.y)
	pdf.CellFormat(b.contentW, 5, "NOTES", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(receiptDarkText[0], receiptDarkText[1], receiptDarkText[2])
	pdf.SetXY(b.marginL, b.y+5)
	pdf.MultiCell(b.contentW, 6, b.payment.Notes, "", "L", false)
	b.y = pdf.GetY() + 4
	return b
}

// Footer draws the brand-colored closing rule and the closing note.
func (b *receiptBuilder) Footer() *receiptBuilder {
	pdf := b.pdf
	pdf.SetDrawColor(receiptBrandColor[0], receiptBrandColor[1], receiptBrandColor[2])
	pdf.SetLineWidth(0.6)
	footerY := b.y + 6
	pdf.Line(b.marginL, footerY, b.pageW-b.marginR, footerY)

	pdf.SetFont("Helvetica", "I", 9)
	pdf.SetTextColor(receiptMutedText[0], receiptMutedText[1], receiptMutedText[2])
	pdf.SetXY(b.marginL, footerY+4)
	pdf.CellFormat(b.contentW, 6, "This is a system-generated receipt. Thank you for your payment.", "", 1, "C", false, 0, "")
	return b
}

// Build finalizes the document and returns its PDF bytes.
func (b *receiptBuilder) Build() ([]byte, error) {
	var buf bytes.Buffer
	if err := b.pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to render receipt PDF: %w", err)
	}
	return buf.Bytes(), nil
}
