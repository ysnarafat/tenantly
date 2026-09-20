-- Per-user TOTP MFA enrollment, isolated from the users table. The secret is
-- stored encrypted at rest (AES-256-GCM, same protector as NID); `enabled` flips
-- true once the user confirms enrollment by verifying a code.
CREATE TABLE IF NOT EXISTS user_mfa (
    user_id    INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    secret     TEXT NOT NULL,
    enabled    BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);
