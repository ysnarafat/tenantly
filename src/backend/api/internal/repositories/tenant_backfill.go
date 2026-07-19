package repositories

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	appcrypto "github.com/ysnarafat/tenantly/internal/crypto"
)

// BackfillTenantNID migrates any legacy plaintext NID values into the protected
// columns (encrypted ciphertext, deterministic hash, last-four) and then nulls
// out the plaintext so no cleartext NID remains at rest. It is idempotent:
// rows already encrypted (nid_encrypted set) or without a plaintext value are
// skipped, so it is safe to run on every startup. The plaintext nid_number
// column itself is retained for one release and should be dropped in a later
// migration once every environment has completed this back-fill.
func BackfillTenantNID(db *sqlx.DB, nid *appcrypto.NIDProtector) error {
	rows, err := db.Query(`
		SELECT id, nid_number
		FROM tenants
		WHERE nid_number IS NOT NULL AND nid_number != '' AND nid_encrypted IS NULL`)
	if err != nil {
		return fmt.Errorf("failed to query tenants for NID back-fill: %w", err)
	}
	defer rows.Close()

	type pending struct {
		id        int
		plaintext string
	}
	var todo []pending
	for rows.Next() {
		var p pending
		if err := rows.Scan(&p.id, &p.plaintext); err != nil {
			return fmt.Errorf("failed to scan tenant for NID back-fill: %w", err)
		}
		todo = append(todo, p)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating tenants for NID back-fill: %w", err)
	}

	if len(todo) == 0 {
		return nil
	}

	for _, p := range todo {
		encrypted, err := nid.Encrypt(p.plaintext)
		if err != nil {
			return fmt.Errorf("failed to encrypt NID for tenant %d: %w", p.id, err)
		}
		_, err = db.Exec(`
			UPDATE tenants
			SET nid_encrypted = $1, nid_last_four = $2, nid_hash = $3, nid_number = NULL
			WHERE id = $4`,
			encrypted, appcrypto.LastFour(p.plaintext), nid.Hash(p.plaintext), p.id)
		if err != nil {
			return fmt.Errorf("failed to back-fill NID for tenant %d: %w", p.id, err)
		}
	}

	log.Printf("NID back-fill: encrypted %d legacy plaintext tenant NID(s)", len(todo))
	return nil
}
