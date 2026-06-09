# PII Encryption at Rest — Implementation Plan

## Overview

Tenantly stores personally identifiable information (PII) for tenants in the `tenants` table. If the database backup or volume is ever compromised, this data is exposed in plaintext. This document defines which fields to encrypt, the chosen approach, and a step-by-step implementation plan.

---

## Fields to Encrypt

| Table      | Column         | Sensitivity | Reason                                  |
|------------|----------------|-------------|-----------------------------------------|
| `tenants`  | `nid_number`   | **High**    | Bangladesh National ID — government ID  |
| `tenants`  | `phone_number` | **High**    | Direct contact, SIM-card tied to NID    |
| `tenants`  | `email`        | Medium      | Personally identifiable contact info    |
| `tenants`  | `address`      | Low–Medium  | Physical location, encrypt opportunistically |

Fields **not** encrypted (low sensitivity / needed for querying):

- `name` — used for display and search; low risk alone
- `tenant_type` — enumeration, not PII
- All foreign keys, timestamps, status flags

---

## Approach: Application-Layer Encryption (AES-256-GCM)

Encrypt in the Go service layer **before** writing to PostgreSQL. The DB stores ciphertext in `TEXT` / `BYTEA` columns.

**Why not `pgcrypto`?**  
`pgcrypto` runs inside the DB process. The encryption key must be passed in every SQL query, which means it appears in query logs, `pg_stat_activity`, and connection metadata. Application-layer encryption keeps the key entirely outside the DB.

**Algorithm:** AES-256-GCM  
- Authenticated encryption (prevents tampering)  
- 32-byte key, 12-byte random nonce prepended to ciphertext  
- Standard library: `golang.org/x/crypto` + `crypto/aes` + `crypto/cipher`

**Key storage:** Environment variable `TENANT_PII_KEY` (32 bytes, base64-encoded). In production this should be sourced from a secret manager (e.g. AWS Secrets Manager, HashiCorp Vault, or at minimum a `.env` file excluded from version control).

---

## Lookup Strategy

Encrypting a field breaks `WHERE column = ?` queries. We handle this with a **blind index**:

- Add a `phone_number_hash` column alongside `phone_number`
- Value: `HMAC-SHA256(phone_number, LOOKUP_HMAC_KEY)` — deterministic, non-reversible
- Use `phone_number_hash` for equality lookups; decrypt `phone_number` only when displaying
- Same pattern applies to `nid_number` → `nid_number_hash`

Two separate keys: `TENANT_PII_KEY` (encryption) and `TENANT_LOOKUP_HMAC_KEY` (hashing). Compromise of one does not compromise the other.

---

## Implementation Steps

### Step 1 — Add a crypto utility package

**File:** `internal/crypto/pii.go`

```go
package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "fmt"
    "io"
)

// Encrypt encrypts plaintext with AES-256-GCM. Returns base64(nonce+ciphertext).
func Encrypt(plaintext string, key []byte) (string, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", fmt.Errorf("encrypt: new cipher: %w", err)
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", fmt.Errorf("encrypt: new gcm: %w", err)
    }
    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return "", fmt.Errorf("encrypt: nonce: %w", err)
    }
    sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt decrypts a value produced by Encrypt.
func Decrypt(ciphertext string, key []byte) (string, error) {
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", fmt.Errorf("decrypt: decode: %w", err)
    }
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", fmt.Errorf("decrypt: new cipher: %w", err)
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", fmt.Errorf("decrypt: new gcm: %w", err)
    }
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return "", fmt.Errorf("decrypt: ciphertext too short")
    }
    plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
    if err != nil {
        return "", fmt.Errorf("decrypt: open: %w", err)
    }
    return string(plaintext), nil
}

// BlindIndex returns HMAC-SHA256(value, hmacKey) as hex — used for lookups.
func BlindIndex(value string, hmacKey []byte) string {
    mac := hmac.New(sha256.New, hmacKey)
    mac.Write([]byte(value))
    return fmt.Sprintf("%x", mac.Sum(nil))
}
```

---

### Step 2 — Load keys in config

**File:** `internal/config/config.go`

Add to the `Config` struct:

```go
TenantPIIKey       []byte // 32 bytes, AES-256
TenantLookupHMACKey []byte // 32 bytes, HMAC key
```

Load from environment:

```go
piiKeyB64 := os.Getenv("TENANT_PII_KEY")
piiKey, err := base64.StdEncoding.DecodeString(piiKeyB64)
if err != nil || len(piiKey) != 32 {
    log.Fatal("TENANT_PII_KEY must be a base64-encoded 32-byte key")
}

hmacKeyB64 := os.Getenv("TENANT_LOOKUP_HMAC_KEY")
hmacKey, err := base64.StdEncoding.DecodeString(hmacKeyB64)
if err != nil || len(hmacKey) != 32 {
    log.Fatal("TENANT_LOOKUP_HMAC_KEY must be a base64-encoded 32-byte key")
}
```

Generate keys (one-time, store in secret manager):

```bash
openssl rand -base64 32   # TENANT_PII_KEY
openssl rand -base64 32   # TENANT_LOOKUP_HMAC_KEY
```

---

### Step 3 — Database migration

**File:** `migrations/000011_tenant_pii_encryption.up.sql`

```sql
ALTER TABLE tenants
    ADD COLUMN nid_number_hash   TEXT,
    ADD COLUMN phone_number_hash TEXT;

-- Existing plaintext data must be migrated by the Go migration tool (Step 4).
-- Hash columns are nullable until migration is complete, then set NOT NULL.

CREATE INDEX idx_tenants_nid_hash   ON tenants (nid_number_hash);
CREATE INDEX idx_tenants_phone_hash ON tenants (phone_number_hash);
```

**File:** `migrations/000011_tenant_pii_encryption.down.sql`

```sql
ALTER TABLE tenants
    DROP COLUMN IF EXISTS nid_number_hash,
    DROP COLUMN IF EXISTS phone_number_hash;
```

---

### Step 4 — One-time data migration tool

**File:** `cmd/migrate-pii/main.go`

A standalone CLI tool that:
1. Reads all tenant rows with plaintext PII
2. Encrypts each value with the new key
3. Computes blind indexes
4. Writes encrypted values + hashes back
5. Verifies decryption round-trips before committing

Run once in a maintenance window with a DB transaction — rollback on any error.

> **Note:** Run this tool **after** deploying Step 5 (service changes) so both old and new code paths work during the transition window. The service can detect if a value is already encrypted (base64 prefix check or a separate `pii_encrypted` boolean column) during the cutover period.

---

### Step 5 — Update TenantRepository

**File:** `internal/repositories/tenant_repository.go`

Inject `piiKey []byte` and `hmacKey []byte` into the repository struct.

On **write** (create/update):
```go
encNID, err := crypto.Encrypt(req.NIDNumber, r.piiKey)
// store encNID in nid_number column
nidHash := crypto.BlindIndex(req.NIDNumber, r.hmacKey)
// store nidHash in nid_number_hash column
```

On **read** (scan row into struct):
```go
t.NIDNumber, err = crypto.Decrypt(t.NIDNumber, r.piiKey)
```

Lookup by NID or phone (e.g. duplicate check):
```go
hash := crypto.BlindIndex(nid, r.hmacKey)
// WHERE nid_number_hash = $1
```

---

### Step 6 — Update `.env.example`

```dotenv
# PII Encryption (AES-256-GCM) — generate with: openssl rand -base64 32
TENANT_PII_KEY=
TENANT_LOOKUP_HMAC_KEY=
```

---

### Step 7 — Add tests

**File:** `internal/crypto/pii_test.go`

- Round-trip: `Decrypt(Encrypt(x)) == x`
- Different nonces produce different ciphertext for same input
- `BlindIndex` is deterministic across calls
- Wrong key returns error

**File:** `internal/repositories/tenant_repository_test.go`

- Extend existing table-driven tests to assert DB stores ciphertext (not plaintext)
- Assert decrypted value matches original after read

---

## Cutover Sequence (Production)

1. Deploy migration `000011` (adds hash columns, existing plaintext untouched)
2. Deploy updated service binary (writes encrypted + hash for new records; reads detect plaintext vs ciphertext)
3. Run `cmd/migrate-pii` in a maintenance window to backfill all existing rows
4. Deploy final binary that assumes all rows are encrypted (remove plaintext detection shim)
5. Optionally set `NOT NULL` constraint on hash columns

---

## What This Does NOT Cover

| Risk | Mitigation needed |
|------|--------------------|
| Key in memory at runtime | Acceptable — process memory is not the threat model here |
| Key rotation | Implement a `key_version` column + re-encryption job when needed |
| Audit log PII | `audit_logs.old_values` / `new_values` may contain plaintext — encrypt or redact separately |
| Transport layer | Already handled by TLS; orthogonal to this plan |
| Application-level access control | Already handled by JWT + role middleware |

---

## Files to Create / Modify

| Action | Path |
|--------|------|
| Create | `internal/crypto/pii.go` |
| Create | `internal/crypto/pii_test.go` |
| Create | `migrations/000011_tenant_pii_encryption.up.sql` |
| Create | `migrations/000011_tenant_pii_encryption.down.sql` |
| Create | `cmd/migrate-pii/main.go` |
| Modify | `internal/config/config.go` |
| Modify | `internal/repositories/tenant_repository.go` |
| Modify | `.env.example` |
