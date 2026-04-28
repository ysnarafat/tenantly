package columns

// Unit table and column names
const (
	UnitTable           = "units"
	UnitID              = "id"
	UnitBuildingID      = "building_id"
	UnitPropertyID      = "property_id"
	UnitOrganizationID  = "organization_id"
	UnitNumber          = "unit_number"
	UnitName            = "unit_name"
	UnitFloor           = "floor"
	UnitSection         = "section"
	UnitType            = "unit_type"
	UnitMonthlyRent     = "monthly_rent"
	UnitMetadata        = "metadata"
	UnitActive          = "active"
	UnitCreatedAt       = "created_at"
	UnitUpdatedAt       = "updated_at"
)

// UnitAllColumns returns a comma-separated list of all unit columns
// for use in SELECT statements
func UnitAllColumns() string {
	return UnitID + ", " +
		UnitBuildingID + ", " +
		UnitPropertyID + ", " +
		UnitOrganizationID + ", " +
		UnitNumber + ", " +
		UnitName + ", " +
		UnitFloor + ", " +
		UnitSection + ", " +
		UnitType + ", " +
		UnitMonthlyRent + ", " +
		UnitMetadata + ", " +
		UnitActive + ", " +
		UnitCreatedAt + ", " +
		UnitUpdatedAt
}

// UnitSelectWithAlias returns all unit columns with a table alias
// Example: UnitSelectWithAlias("u") returns "u.id, u.building_id, ..."
func UnitSelectWithAlias(alias string) string {
	if alias == "" {
		return UnitAllColumns()
	}
	return alias + "." + UnitID + ", " +
		alias + "." + UnitBuildingID + ", " +
		alias + "." + UnitPropertyID + ", " +
		alias + "." + UnitOrganizationID + ", " +
		alias + "." + UnitNumber + ", " +
		alias + "." + UnitName + ", " +
		alias + "." + UnitFloor + ", " +
		alias + "." + UnitSection + ", " +
		alias + "." + UnitType + ", " +
		alias + "." + UnitMonthlyRent + ", " +
		alias + "." + UnitMetadata + ", " +
		alias + "." + UnitActive + ", " +
		alias + "." + UnitCreatedAt + ", " +
		alias + "." + UnitUpdatedAt
}
