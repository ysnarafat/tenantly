-- Backs the application-level "unit already has an active lease" check
-- (LeaseService.CreateLease / HasActiveLeaseOnUnit) with a real DB
-- constraint, closing the check-then-insert race where two concurrent
-- requests could both pass the check and both insert an active lease for
-- the same unit.
CREATE UNIQUE INDEX idx_leases_unit_active_unique ON leases(unit_id) WHERE active = true;
