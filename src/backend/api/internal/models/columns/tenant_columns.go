package columns

// Tenant table and column constants
const (
	TenantTable = "tenants"

	// Tenant columns
	TenantID           = "id"
	TenantName         = "name"
	TenantType         = "tenant_type"
	TenantPhoneNumber  = "phone_number"
	TenantEmail        = "email"
	TenantNIDNumber    = "nid_number"
	TenantNIDEncrypted = "nid_encrypted"
	TenantNIDLastFour  = "nid_last_four"
	TenantNIDHash      = "nid_hash"
	TenantAddress      = "address"
	TenantActive       = "active"
	TenantOrgID        = "organization_id"
	TenantCreatedAt    = "created_at"
	TenantUpdatedAt    = "updated_at"
)

// TenantAllColumns returns the tenant columns for standard SELECT queries. The
// full NID is never selected here; only the non-sensitive last-four projection.
func TenantAllColumns() string {
	return "id, name, tenant_type, phone_number, email, nid_last_four, address, active, organization_id, created_at, updated_at"
}
