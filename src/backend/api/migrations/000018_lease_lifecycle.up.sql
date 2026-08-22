-- Leases are meant to be immutable historical records of a specific tenancy
-- term. Renewing a lease or ending one early should be traceable: why it
-- stopped being active, and — for a renewal — which lease replaced it.
ALTER TABLE leases ADD COLUMN end_reason VARCHAR(20)
    CHECK (end_reason IN ('Expired', 'Terminated', 'Renewed'));
ALTER TABLE leases ADD COLUMN renewed_from_lease_id INTEGER REFERENCES leases(id);

CREATE INDEX idx_leases_renewed_from_lease_id ON leases(renewed_from_lease_id)
    WHERE renewed_from_lease_id IS NOT NULL;

COMMENT ON COLUMN leases.end_reason IS 'Why this lease stopped being active: naturally Expired, ended early via Terminated, or replaced via Renewed';
COMMENT ON COLUMN leases.renewed_from_lease_id IS 'If this lease was created by renewing an earlier one, the lease it replaced';
