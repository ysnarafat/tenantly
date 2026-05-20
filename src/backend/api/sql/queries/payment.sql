-- Queries for payment repository
-- All payment-related queries for sqlc generation

-- name: CreatePayment :one
INSERT INTO payments (
    unit_id, tenant_id, building_id, property_id, organization_id,
    month, year, amount_due, amount_paid, status, due_date
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, 0, 'Due', $9
)
RETURNING id, unit_id, tenant_id, building_id, property_id, organization_id,
    month, year, amount_due, amount_paid, status,
    COALESCE(payment_method, '') AS payment_method,
    COALESCE(notes, '') AS notes,
    COALESCE(receipt_number, '') AS receipt_number,
    payment_date, due_date, created_at, updated_at;

-- name: GetPaymentByID :one
SELECT id, unit_id, tenant_id, building_id, property_id, organization_id,
    month, year, amount_due, amount_paid, status,
    COALESCE(payment_method, '') AS payment_method,
    COALESCE(notes, '') AS notes,
    COALESCE(receipt_number, '') AS receipt_number,
    payment_date, due_date, created_at, updated_at
FROM payments
WHERE id = $1;

-- name: GetPaymentByIDWithDetails :one
SELECT
    p.id, p.unit_id, p.tenant_id, p.building_id, p.property_id, p.organization_id,
    p.month, p.year, p.amount_due, p.amount_paid, p.status,
    COALESCE(p.payment_method, '') AS payment_method,
    COALESCE(p.notes, '') AS notes,
    COALESCE(p.receipt_number, '') AS receipt_number,
    p.payment_date, p.due_date, p.created_at, p.updated_at,
    COALESCE(pr.property_name, '') AS property_name,
    COALESCE(b.building_name, '') AS building_name,
    COALESCE(b.building_code, '') AS building_code,
    COALESCE(u.unit_number, '') AS unit_number,
    COALESCE(u.unit_type::text, '') AS unit_type,
    COALESCE(t.name, '') AS tenant_name
FROM payments p
LEFT JOIN units u ON p.unit_id = u.id
LEFT JOIN tenants t ON p.tenant_id = t.id
LEFT JOIN buildings b ON p.building_id = b.id
LEFT JOIN properties pr ON p.property_id = pr.id
WHERE p.id = $1;

-- name: CountBuildingPaymentsInPeriod :one
SELECT COUNT(*)
FROM payments p
LEFT JOIN units u ON p.unit_id = u.id
LEFT JOIN tenants t ON p.tenant_id = t.id
LEFT JOIN buildings b ON p.building_id = b.id
LEFT JOIN properties pr ON p.property_id = pr.id
WHERE p.building_id = $1
  AND p.created_at >= $2
  AND p.created_at <= $3;

-- name: GetBuildingPaymentsInPeriod :many
SELECT
    p.id, p.unit_id, p.tenant_id, p.building_id, p.property_id, p.organization_id,
    p.month, p.year, p.amount_due, p.amount_paid, p.status,
    COALESCE(p.payment_method, '') AS payment_method,
    COALESCE(p.notes, '') AS notes,
    COALESCE(p.receipt_number, '') AS receipt_number,
    p.payment_date, p.due_date, p.created_at, p.updated_at,
    COALESCE(pr.property_name, '') AS property_name,
    COALESCE(b.building_name, '') AS building_name,
    COALESCE(b.building_code, '') AS building_code,
    COALESCE(u.unit_number, '') AS unit_number,
    COALESCE(u.unit_type::text, '') AS unit_type,
    COALESCE(t.name, '') AS tenant_name
FROM payments p
LEFT JOIN units u ON p.unit_id = u.id
LEFT JOIN tenants t ON p.tenant_id = t.id
LEFT JOIN buildings b ON p.building_id = b.id
LEFT JOIN properties pr ON p.property_id = pr.id
WHERE p.building_id = $1
  AND p.created_at >= $2
  AND p.created_at <= $3
ORDER BY p.year DESC, p.month DESC
LIMIT $4 OFFSET $5;

-- name: GetBuildingPaymentStats :one
SELECT
    COUNT(*) AS total_records,
    COALESCE(SUM(amount_due), 0) AS total_due,
    COALESCE(SUM(amount_paid), 0) AS total_paid,
    COALESCE(SUM(CASE WHEN status='Overdue' THEN amount_due - amount_paid ELSE 0 END), 0) AS total_overdue,
    COUNT(CASE WHEN status='Paid' THEN 1 END) AS paid_count,
    COUNT(CASE WHEN status='Due' THEN 1 END) AS due_count,
    COUNT(CASE WHEN status='Partial' THEN 1 END) AS partial_count,
    COUNT(CASE WHEN status='Overdue' THEN 1 END) AS overdue_count
FROM payments
WHERE building_id = $1
  AND created_at >= $2
  AND created_at <= $3;

-- name: GetPropertyPaymentStats :one
SELECT
    COUNT(*) AS total_records,
    COALESCE(SUM(amount_due), 0) AS total_due,
    COALESCE(SUM(amount_paid), 0) AS total_paid,
    COALESCE(SUM(CASE WHEN status='Overdue' THEN amount_due - amount_paid ELSE 0 END), 0) AS total_overdue,
    COUNT(CASE WHEN status='Paid' THEN 1 END) AS paid_count,
    COUNT(CASE WHEN status='Due' THEN 1 END) AS due_count,
    COUNT(CASE WHEN status='Partial' THEN 1 END) AS partial_count,
    COUNT(CASE WHEN status='Overdue' THEN 1 END) AS overdue_count
FROM payments
WHERE property_id = $1
  AND created_at >= $2
  AND created_at <= $3;

-- name: GetSystemPaymentStats :one
SELECT
    COUNT(*) AS total_records,
    COALESCE(SUM(amount_due), 0) AS total_due,
    COALESCE(SUM(amount_paid), 0) AS total_paid,
    COALESCE(SUM(CASE WHEN status='Overdue' THEN amount_due - amount_paid ELSE 0 END), 0) AS total_overdue,
    COUNT(CASE WHEN status='Paid' THEN 1 END) AS paid_count,
    COUNT(CASE WHEN status='Due' THEN 1 END) AS due_count,
    COUNT(CASE WHEN status='Partial' THEN 1 END) AS partial_count,
    COUNT(CASE WHEN status='Overdue' THEN 1 END) AS overdue_count
FROM payments
WHERE created_at >= $1
  AND created_at <= $2;

-- name: GetDashboardSummary :one
SELECT
    COALESCE(SUM(amount_due), 0) AS total_due,
    COALESCE(SUM(amount_paid), 0) AS total_paid,
    COALESCE(SUM(CASE WHEN status != 'Paid' THEN amount_due - amount_paid ELSE 0 END), 0) AS total_pending,
    COALESCE(SUM(CASE WHEN status='Overdue' THEN amount_due - amount_paid ELSE 0 END), 0) AS total_overdue,
    (SELECT COUNT(DISTINCT id) FROM properties) AS property_count,
    (SELECT COUNT(DISTINCT id) FROM buildings) AS building_count,
    (SELECT COUNT(DISTINCT id) FROM units WHERE active = true) AS unit_count,
    (SELECT COUNT(DISTINCT id) FROM tenants WHERE active = true) AS tenant_count
FROM payments;

-- name: GetBuildingLevelSummary :one
SELECT
    COUNT(DISTINCT building_id) AS total_buildings,
    COALESCE(SUM(amount_due), 0) AS total_due,
    COALESCE(SUM(amount_paid), 0) AS total_paid
FROM payments;

-- name: GetBuildingPaymentAnalytics :one
SELECT
    COALESCE(SUM(amount_paid), 0) AS total_revenue,
    CASE WHEN COALESCE(SUM(amount_due), 0) > 0
        THEN (COALESCE(SUM(amount_paid), 0) / COALESCE(SUM(amount_due), 0)) * 100
        ELSE 0 END AS collection_rate,
    COALESCE(AVG(
        CASE WHEN payment_date IS NOT NULL AND due_date IS NOT NULL
            THEN EXTRACT(EPOCH FROM (payment_date - due_date)) / 86400
            ELSE NULL END
    ), 0) AS avg_payment_days,
    COUNT(CASE WHEN status='Overdue' THEN 1 END) AS overdue_count
FROM payments
WHERE building_id = $1 AND created_at >= $2 AND created_at <= $3;

-- name: GetBuildingPaymentTrend :many
SELECT year, month,
    COALESCE(SUM(amount_due), 0) AS due,
    COALESCE(SUM(amount_paid), 0) AS paid
FROM payments
WHERE building_id = $1 AND created_at >= $2 AND created_at <= $3
GROUP BY year, month
ORDER BY year, month;

-- name: SearchLeases :many
SELECT
    l.id AS lease_id,
    t.id AS tenant_id,
    t.name AS tenant_name,
    COALESCE(t.phone_number, '') AS tenant_phone,
    p.id AS property_id,
    p.property_name,
    b.id AS building_id,
    b.building_name,
    b.building_code,
    u.id AS unit_id,
    u.unit_number,
    COALESCE(u.unit_type::text, '') AS unit_type,
    l.start_date AS lease_start_date,
    l.end_date AS lease_end_date,
    l.monthly_rent,
    l.active
FROM leases l
LEFT JOIN tenants t ON l.tenant_id = t.id
LEFT JOIN units u ON l.unit_id = u.id
LEFT JOIN buildings b ON u.building_id = b.id
LEFT JOIN properties p ON u.property_id = p.id
WHERE l.organization_id = $1
    AND l.active = true
    AND (
        LOWER(t.name) LIKE $2
        OR LOWER(p.property_name) LIKE $2
        OR LOWER(b.building_name) LIKE $2
        OR LOWER(b.building_code) LIKE $2
        OR LOWER(u.unit_number) LIKE $2
        OR LOWER(t.phone_number) LIKE $2
        OR CAST(l.id AS TEXT) LIKE $2
    )
ORDER BY t.name, l.start_date DESC
LIMIT 50;
