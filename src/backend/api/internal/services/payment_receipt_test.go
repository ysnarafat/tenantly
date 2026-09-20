package services

import (
	"bytes"
	"testing"

	"github.com/ysnarafat/tenantly/internal/models"
)

func TestFormatBDT(t *testing.T) {
	tests := []struct {
		amount float64
		want   string
	}{
		{0, "BDT 0.00"},
		{5, "BDT 5.00"},
		{999, "BDT 999.00"},
		{1000, "BDT 1,000.00"},
		{12500, "BDT 12,500.00"},
		{100000, "BDT 1,00,000.00"},      // 1 lakh
		{1234567.89, "BDT 12,34,567.89"}, // 12 lakh 34 thousand 567
		{-500, "-BDT 500.00"},
	}
	for _, tt := range tests {
		if got := formatBDT(tt.amount); got != tt.want {
			t.Errorf("formatBDT(%v) = %q, want %q", tt.amount, got, tt.want)
		}
	}
}

func TestGenerateReceiptPDF_ProducesValidPDF(t *testing.T) {
	svc, payRepo, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

	unitRepo.addUnit(sampleUnit(1, 2, 3))
	bldgRepo.addBuilding(sampleBuilding(2, 3))
	propRepo.addProperty(sampleProperty(3))
	created, err := svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	// Fill in the display fields a real GetByIDWithDetails join would
	// populate, so the renderer has something non-empty to lay out.
	pwd := payRepo.payments[created.ID]
	pwd.TenantName = "Rahim Uddin"
	pwd.PropertyName = "Sunrise Apartments"
	pwd.BuildingName = "Block A"
	pwd.BuildingCode = "BLK-A"
	pwd.UnitNumber = "3B"
	pwd.PaymentMethod = "bKash"
	pwd.ReceiptNumber = "RCP-202606-0001"
	pwd.AmountPaid = pwd.AmountDue
	pwd.Status = models.PaymentStatusPaid

	pdfBytes, err := svc.GenerateReceiptPDF(created.ID, 1)
	if err != nil {
		t.Fatalf("unexpected error generating receipt: %v", err)
	}
	if len(pdfBytes) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		t.Error("expected output to start with the PDF magic bytes (%PDF)")
	}
}

func TestGenerateReceiptPDF_CrossOrgPaymentNotFound(t *testing.T) {
	svc, _, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

	unitRepo.addUnit(sampleUnit(1, 2, 3))
	bldgRepo.addBuilding(sampleBuilding(2, 3))
	propRepo.addProperty(sampleProperty(3))
	created, err := svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	// orgID 999 does not own this payment (it was created under org 1).
	if _, err := svc.GenerateReceiptPDF(created.ID, 999); err == nil {
		t.Fatal("expected an error generating a receipt for another org's payment, got nil")
	}
}

func TestGenerateReceiptPDF_RendersEveryStatusColor(t *testing.T) {
	statuses := []models.PaymentStatus{
		models.PaymentStatusPaid,
		models.PaymentStatusDue,
		models.PaymentStatusPartial,
		models.PaymentStatusOverdue,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			svc, payRepo, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

			unitRepo.addUnit(sampleUnit(1, 2, 3))
			bldgRepo.addBuilding(sampleBuilding(2, 3))
			propRepo.addProperty(sampleProperty(3))
			created, err := svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)
			if err != nil {
				t.Fatalf("failed to create payment: %v", err)
			}

			payRepo.payments[created.ID].Status = status

			pdfBytes, err := svc.GenerateReceiptPDF(created.ID, 1)
			if err != nil {
				t.Fatalf("unexpected error generating receipt for status %s: %v", status, err)
			}
			if len(pdfBytes) == 0 {
				t.Fatalf("expected non-empty PDF bytes for status %s", status)
			}
		})
	}
}
