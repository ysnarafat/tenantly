package database

import (
	"encoding/json"
	"strings"
	"testing"
)

// toJSON masks the value and returns it as a JSON string for easy assertions.
func maskToJSON(t *testing.T, v interface{}) string {
	t.Helper()
	masked, err := maskAuditValue(v)
	if err != nil {
		t.Fatalf("maskAuditValue error: %v", err)
	}
	b, err := json.Marshal(masked)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	return string(b)
}

func TestMaskAuditValue_RedactsSensitiveFields(t *testing.T) {
	in := map[string]interface{}{
		"id":            7,
		"name":          "John Doe",
		"email":         "john@example.com",
		"phone_number":  "01700000000",
		"nid_number":    "1234567890",
		"address":       "12 Gulshan Ave, Dhaka",
		"password_hash": "$2a$10$secrethash",
		"tenant_type":   "Individual",
	}

	got := maskToJSON(t, in)

	// Sensitive values must not appear anywhere in the serialized audit payload.
	for _, leaked := range []string{
		"John Doe", "john@example.com", "01700000000", "1234567890",
		"12 Gulshan Ave, Dhaka", "$2a$10$secrethash",
	} {
		if contains(got, leaked) {
			t.Errorf("sensitive value %q leaked into audit payload: %s", leaked, got)
		}
	}
	// Non-sensitive values must be preserved.
	for _, kept := range []string{`"id":7`, "Individual"} {
		if !contains(got, kept) {
			t.Errorf("expected %q to be preserved in %s", kept, got)
		}
	}
	// Redaction marker present.
	if !contains(got, "***REDACTED***") {
		t.Errorf("expected redaction marker in %s", got)
	}
}

func TestMaskAuditValue_KeyNormalization(t *testing.T) {
	in := map[string]interface{}{
		"phoneNumber":  "01711111111", // camelCase
		"PHONE_NUMBER": "01722222222",
		"First_Name":   "Jane",
	}
	got := maskToJSON(t, in)
	for _, leaked := range []string{"01711111111", "01722222222", "Jane"} {
		if contains(got, leaked) {
			t.Errorf("value %q leaked despite key-casing variance: %s", leaked, got)
		}
	}
}

func TestMaskAuditValue_PreservesNullAndNonSensitive(t *testing.T) {
	in := map[string]interface{}{
		"email":         nil, // empty field stays null, not redacted
		"ip_address":    "203.0.113.9",
		"user_agent":    "Mozilla/5.0",
		"username":      "acme_admin",
		"building_name": "Tower A",
	}
	got := maskToJSON(t, in)

	if !contains(got, `"email":null`) {
		t.Errorf("null email should stay null, got %s", got)
	}
	// Operational/security fields and non-PII names are intentionally kept.
	for _, kept := range []string{"203.0.113.9", "Mozilla/5.0", "acme_admin", "Tower A"} {
		if !contains(got, kept) {
			t.Errorf("expected %q to be preserved in %s", kept, got)
		}
	}
}

func TestMaskAuditValue_NestedStructs(t *testing.T) {
	type inner struct {
		Email string `json:"email"`
		City  string `json:"city"`
	}
	type outer struct {
		Contact inner  `json:"contact"`
		Ref     string `json:"ref"`
	}
	got := maskToJSON(t, outer{Contact: inner{Email: "nested@example.com", City: "Dhaka"}, Ref: "R-1"})

	if contains(got, "nested@example.com") {
		t.Errorf("nested email leaked: %s", got)
	}
	for _, kept := range []string{"Dhaka", "R-1"} {
		if !contains(got, kept) {
			t.Errorf("expected %q preserved in %s", kept, got)
		}
	}
}

func TestMaskAuditValue_Nil(t *testing.T) {
	masked, err := maskAuditValue(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if masked != nil {
		t.Errorf("expected nil, got %v", masked)
	}
}

func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}
