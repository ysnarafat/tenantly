package services

import (
	"net/url"
	"testing"
)

func TestInvitationURLUsesActivationRouteAndEncodesToken(t *testing.T) {
	sender := NewSMTPInvitationEmailSender("smtp.example.com", 587, "user", "password", "noreply@example.com", "https://app.example.com/base/")
	link, err := sender.invitationURL("token+/with spaces")
	if err != nil {
		t.Fatalf("invitationURL returned an error: %v", err)
	}

	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse invitation URL: %v", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "app.example.com" || parsed.Path != "/base/accept-invitation" {
		t.Fatalf("unexpected invitation URL: %s", link)
	}
	if parsed.Query().Get("token") != "token+/with spaces" {
		t.Fatalf("invitation token was not preserved: %s", link)
	}
}

func TestInvitationURLRejectsUnsafeBaseURL(t *testing.T) {
	sender := NewSMTPInvitationEmailSender("smtp.example.com", 587, "", "", "noreply@example.com", "javascript:alert(1)")
	if _, err := sender.invitationURL("token"); err == nil {
		t.Fatal("expected invalid invitation base URL to be rejected")
	}
}
