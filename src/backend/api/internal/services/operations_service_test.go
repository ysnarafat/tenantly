package services

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
)

func TestSanitizePlanFeatures(t *testing.T) {
	features, err := sanitizePlanFeatures([]string{"  Manage properties  ", "", "Track leases"})
	if err != nil {
		t.Fatalf("sanitizePlanFeatures returned an error: %v", err)
	}
	if len(features) != 2 || features[0] != "Manage properties" || features[1] != "Track leases" {
		t.Fatalf("unexpected sanitized features: %#v", features)
	}
}

func TestSanitizePlanFeaturesRejectsTooManyOrLongItems(t *testing.T) {
	many := make([]string, 21)
	if _, err := sanitizePlanFeatures(many); err == nil {
		t.Fatal("expected more than 20 features to be rejected")
	}

	if _, err := sanitizePlanFeatures([]string{strings.Repeat("x", 161)}); err == nil {
		t.Fatal("expected a feature over 160 characters to be rejected")
	}
}

func TestValidateSubscriptionDates(t *testing.T) {
	service := &OperationsService{}
	start := time.Now().UTC()
	renewal := start.Add(-time.Hour)
	if err := service.ValidateSubscriptionDates(&models.UpdateSubscriptionRequest{
		StartedAt: &start,
		RenewsAt:  &renewal,
	}); err == nil {
		t.Fatal("expected renewal before start to be rejected")
	}

	future := time.Now().UTC().Add(time.Hour)
	if err := service.ValidateSubscriptionDates(&models.UpdateSubscriptionRequest{
		StartedAt: &future,
	}); err == nil {
		t.Fatal("expected a future start date to be rejected")
	}
}

func TestOnboardOrganizationRejectsMalformedSlugBeforeDatabaseAccess(t *testing.T) {
	service := &OperationsService{}
	_, err := service.OnboardOrganization(&models.CreateOrganizationOnboardingRequest{
		Name:       "New Organization",
		Slug:       "Bad Slug",
		PlanCode:   "basic",
		AdminEmail: "admin@example.com",
	}, 1)
	if err == nil || !errors.Is(err, ErrInvalidOperationsInput) {
		t.Fatalf("expected invalid slug error, got %v", err)
	}
}
