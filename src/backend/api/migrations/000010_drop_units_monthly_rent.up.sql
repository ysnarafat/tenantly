-- monthly_rent belongs to leases, not units; drop the column
ALTER TABLE units DROP COLUMN monthly_rent;
