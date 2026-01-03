package columns

// Property table and column names
const (
	PropertyTable          = "properties"
	PropertyID             = "id"
	PropertyName           = "property_name"
	PropertyCode           = "property_code"
	PropertyAddress        = "address"
	PropertyCity           = "city"
	PropertyPostalCode     = "postal_code"
	PropertyType           = "property_type"
	PropertyTotalBuildings = "total_buildings"
	PropertyMetadata       = "metadata"
	PropertyActive         = "active"
	PropertyCreatedAt      = "created_at"
	PropertyUpdatedAt      = "updated_at"
)

// PropertyAllColumns returns a comma-separated list of all property columns
// for use in SELECT statements
func PropertyAllColumns() string {
	return PropertyID + ", " +
		PropertyName + ", " +
		PropertyCode + ", " +
		PropertyAddress + ", " +
		PropertyCity + ", " +
		PropertyPostalCode + ", " +
		PropertyType + ", " +
		PropertyTotalBuildings + ", " +
		PropertyMetadata + ", " +
		PropertyActive + ", " +
		PropertyCreatedAt + ", " +
		PropertyUpdatedAt
}

// PropertySelectWithAlias returns all property columns with a table alias
// Example: PropertySelectWithAlias("p") returns "p.id, p.property_name, ..."
func PropertySelectWithAlias(alias string) string {
	if alias == "" {
		return PropertyAllColumns()
	}
	return alias + "." + PropertyID + ", " +
		alias + "." + PropertyName + ", " +
		alias + "." + PropertyCode + ", " +
		alias + "." + PropertyAddress + ", " +
		alias + "." + PropertyCity + ", " +
		alias + "." + PropertyPostalCode + ", " +
		alias + "." + PropertyType + ", " +
		alias + "." + PropertyTotalBuildings + ", " +
		alias + "." + PropertyMetadata + ", " +
		alias + "." + PropertyActive + ", " +
		alias + "." + PropertyCreatedAt + ", " +
		alias + "." + PropertyUpdatedAt
}
