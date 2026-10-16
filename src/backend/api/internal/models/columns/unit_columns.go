package columns

// Unit table and column names
const (
	UnitTable                  = "units"
	UnitID                     = "id"
	UnitBuildingID             = "building_id"
	UnitPropertyID             = "property_id"
	UnitOrganizationID         = "organization_id"
	UnitNumber                 = "unit_number"
	UnitName                   = "unit_name"
	UnitFloor                  = "floor"
	UnitSection                = "section"
	UnitType                   = "unit_type"
	UnitMetadata               = "metadata"
	UnitActive                 = "active"
	UnitDefaultLeaseType       = "default_lease_type"
	UnitDefaultMonthlyRent     = "default_monthly_rent"
	UnitDefaultSecurityDeposit = "default_security_deposit"
	UnitDefaultDurationMonths  = "default_duration_months"
	UnitCreatedAt              = "created_at"
	UnitUpdatedAt              = "updated_at"
)

// unitColumnOrder is the canonical column ordering shared by UnitAllColumns and
// UnitSelectWithAlias. Scan targets in the repository must match this order.
var unitColumnOrder = []string{
	UnitID,
	UnitBuildingID,
	UnitPropertyID,
	UnitOrganizationID,
	UnitNumber,
	UnitName,
	UnitFloor,
	UnitSection,
	UnitType,
	UnitMetadata,
	UnitActive,
	UnitDefaultLeaseType,
	UnitDefaultMonthlyRent,
	UnitDefaultSecurityDeposit,
	UnitDefaultDurationMonths,
	UnitCreatedAt,
	UnitUpdatedAt,
}

// UnitAllColumns returns a comma-separated list of all unit columns
// for use in SELECT statements
func UnitAllColumns() string {
	return UnitSelectWithAlias("")
}

// UnitSelectWithAlias returns all unit columns with a table alias
// Example: UnitSelectWithAlias("u") returns "u.id, u.building_id, ..."
func UnitSelectWithAlias(alias string) string {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}

	out := ""
	for i, col := range unitColumnOrder {
		if i > 0 {
			out += ", "
		}
		out += prefix + col
	}
	return out
}
