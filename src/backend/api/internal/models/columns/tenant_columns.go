package columns

// Tenant table and column constants
const (
	TenantTable = "tenants"

	// Tenant columns (non-PII operational data)
	TenantID        = "id"
	TenantName      = "name"
	TenantType      = "tenant_type"
	TenantAddress   = "address"
	TenantActive    = "active"
	TenantOrgID     = "organization_id"
	TenantCreatedAt = "created_at"
	TenantUpdatedAt = "updated_at"

	// tenant_pii columns (personally identifiable information)
	TenantPiiTable          = "tenant_pii"
	TenantPiiTenantID       = "tenant_id"
	TenantPiiPhoneNumber    = "phone_number"
	TenantPiiPhoneEncrypted = "phone_encrypted"
	TenantPiiPhoneHash      = "phone_hash"
	TenantPiiEmail          = "email"
	TenantPiiEmailEncrypted = "email_encrypted"
	TenantPiiEmailHash      = "email_hash"
	TenantPiiNIDNumber      = "nid_number"
	TenantPiiNIDEncrypted   = "nid_encrypted"
	TenantPiiNIDLastFour    = "nid_last_four"
	TenantPiiNIDHash        = "nid_hash"
	TenantPiiCreatedAt      = "created_at"
	TenantPiiUpdatedAt      = "updated_at"
)

// TenantAllColumns returns the non-PII tenant columns for standard SELECT
// queries against the tenants table.
func TenantAllColumns() string {
	return "t.id, t.name, t.tenant_type, t.address, t.active, t.organization_id, t.created_at, t.updated_at"
}

// TenantPiiAllColumns returns the tenant_pii columns for a join alongside the
// tenants table. PII columns are qualified with the tenant_pii table alias.
func TenantPiiAllColumns(alias string) string {
	return alias + ".phone_number, " +
		alias + ".phone_encrypted, " +
		alias + ".phone_hash, " +
		alias + ".email, " +
		alias + ".email_encrypted, " +
		alias + ".email_hash, " +
		alias + ".nid_number, " +
		alias + ".nid_encrypted, " +
		alias + ".nid_last_four, " +
		alias + ".nid_hash"
}
