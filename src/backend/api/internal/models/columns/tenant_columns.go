package columns

// Tenant table and column constants
const (
	TenantTable = "tenants"

	// Tenant columns
	TenantID          = "id"
	TenantName        = "name"
	TenantType        = "tenant_type"
	TenantPhoneNumber = "phone_number"
	TenantEmail       = "email"
	TenantNIDNumber   = "nid_number"
	TenantAddress     = "address"
	TenantActive      = "active"
	TenantOrgID       = "organization_id"
	TenantCreatedAt   = "created_at"
	TenantUpdatedAt   = "updated_at"
)

// TenantAllColumns returns all tenant columns for SELECT queries
func TenantAllColumns() string {
	return "id, name, tenant_type, phone_number, email, nid_number, address, active, organization_id, created_at, updated_at"
}
