package models

import "time"

// ReceiptTokenValidity is how long a receipt download link stays valid once
// issued — long-lived (about a year) since a tenant may reasonably want to
// re-download the same receipt months later as proof of payment, unlike a
// one-time invitation/reset link.
const ReceiptTokenValidity = 365 * 24 * time.Hour

// ReceiptAccessToken grants unauthenticated access to a single payment's
// receipt PDF — the mechanism that lets a "payment recorded" SMS include a
// working download link, since tenants have no login of their own.
type ReceiptAccessToken struct {
	ID        int       `json:"id" db:"id"`
	Token     string    `json:"token" db:"token"`
	PaymentID int       `json:"payment_id" db:"payment_id"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
