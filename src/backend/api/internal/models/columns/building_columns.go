package columns

// Building table and column names
const (
	BuildingTable            = "buildings"
	BuildingID               = "id"
	BuildingPropertyID       = "property_id"
	BuildingName             = "building_name"
	BuildingCode             = "building_code"
	BuildingType             = "building_type"
	BuildingTotalFloors      = "total_floors"
	BuildingHasElevator      = "has_elevator"
	BuildingConstructionYear = "construction_year"
	BuildingMetadata         = "metadata"
	BuildingActiveStatus     = "active_status"
	BuildingCreatedAt        = "created_at"
	BuildingUpdatedAt        = "updated_at"
)

// BuildingAllColumns returns a comma-separated list of all building columns
// for use in SELECT statements
func BuildingAllColumns() string {
	return BuildingID + ", " +
		BuildingPropertyID + ", " +
		BuildingName + ", " +
		BuildingCode + ", " +
		BuildingType + ", " +
		BuildingTotalFloors + ", " +
		BuildingHasElevator + ", " +
		BuildingConstructionYear + ", " +
		BuildingMetadata + ", " +
		BuildingActiveStatus + ", " +
		BuildingCreatedAt + ", " +
		BuildingUpdatedAt
}

// BuildingSelectWithAlias returns all building columns with a table alias
// Example: BuildingSelectWithAlias("b") returns "b.id, b.property_id, ..."
func BuildingSelectWithAlias(alias string) string {
	if alias == "" {
		return BuildingAllColumns()
	}
	return alias + "." + BuildingID + ", " +
		alias + "." + BuildingPropertyID + ", " +
		alias + "." + BuildingName + ", " +
		alias + "." + BuildingCode + ", " +
		alias + "." + BuildingType + ", " +
		alias + "." + BuildingTotalFloors + ", " +
		alias + "." + BuildingHasElevator + ", " +
		alias + "." + BuildingConstructionYear + ", " +
		alias + "." + BuildingMetadata + ", " +
		alias + "." + BuildingActiveStatus + ", " +
		alias + "." + BuildingCreatedAt + ", " +
		alias + "." + BuildingUpdatedAt
}
