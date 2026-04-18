# Lease Repository Testing Documentation

## Overview
This document describes the comprehensive test suite for the `GetLeasesDueForMonth(orgID int)` method in the Lease Repository.

## Test File Location
`src/backend/api/internal/repositories/lease_repository_test.go`

## Test Coverage Summary

### Total Test Cases: 17

The test suite covers the following scenarios:

### 1. Happy Path Tests

#### TestGetLeasesDueForMonth_Success
- **Purpose**: Validates successful retrieval of multiple leases due for the current month
- **Setup**: Creates 3 active leases for current month
- **Verifies**: 
  - Correct number of leases returned (3)
  - All leases belong to the correct organization
  - No errors occurred

#### TestGetLeasesDueForMonth_DataIntegrity
- **Purpose**: Validates that all expected data fields are correctly populated
- **Setup**: Creates a single lease with full related data
- **Verifies**: All 13 fields of LeaseDue model:
  - LeaseID, TenantID, TenantName
  - UnitID, UnitNumber, UnitType
  - BuildingID, BuildingName, BuildingCode
  - PropertyID, PropertyName
  - MonthlyRent, DaysOverdue, OrganizationID

### 2. Payment Status Tests

#### TestGetLeasesDueForMonth_WithPaidPayments
- **Purpose**: Ensures leases with "Paid" status payments are excluded
- **Setup**: Creates 2 leases - one paid, one unpaid
- **Verifies**: Only unpaid lease appears in results

#### TestGetLeasesDueForMonth_WithPartialPayments
- **Purpose**: Ensures leases with "Partial" status payments are excluded
- **Setup**: Creates a lease with partial payment
- **Verifies**: Lease with partial payment is excluded

#### TestGetLeasesDueForMonth_PendingPaymentStatus
- **Purpose**: Ensures leases with "Pending" status payments are included
- **Setup**: Creates a lease with pending payment
- **Verifies**: Lease with pending payment is included in due list

#### TestGetLeasesDueForMonth_PreviousMonthPayments
- **Purpose**: Validates that previous month payments don't affect current month
- **Setup**: Creates a lease with payment for previous month
- **Verifies**: Lease is still due for current month

#### TestGetLeasesDueForMonth_FuturePayment
- **Purpose**: Validates that future month payments don't affect current month
- **Setup**: Creates a lease with payment for next month
- **Verifies**: Lease is still due for current month

### 3. Lease Status Tests

#### TestGetLeasesDueForMonth_InactiveLeases
- **Purpose**: Ensures inactive leases are excluded
- **Setup**: Creates an inactive lease
- **Verifies**: Inactive lease is excluded from results

#### TestGetLeasesDueForMonth_LeaseNotStarted
- **Purpose**: Ensures leases that haven't started are excluded
- **Setup**: Creates a lease starting next month
- **Verifies**: Future lease is excluded

#### TestGetLeasesDueForMonth_LeaseEnded
- **Purpose**: Ensures expired leases are excluded
- **Setup**: Creates a lease that ended last month
- **Verifies**: Expired lease is excluded

### 4. Organization Scoping Tests

#### TestGetLeasesDueForMonth_OrganizationScoping
- **Purpose**: Validates proper organization data isolation
- **Setup**: Creates leases for 2 different organizations
- **Verifies**: 
  - Querying org1 returns only org1's lease
  - Querying org2 returns only org2's lease
  - No data leakage between organizations

#### TestGetLeasesDueForMonth_InvalidOrgID
- **Purpose**: Tests behavior with non-existent organization
- **Setup**: Queries with invalid org ID (999999)
- **Verifies**: Returns empty slice without error

### 5. Days Overdue Calculation Tests

#### TestGetLeasesDueForMonth_DaysOverdueCalculation
- **Purpose**: Validates days overdue calculation for month-start leases
- **Setup**: Creates lease starting on 1st of current month
- **Verifies**: Days overdue is reasonable and non-negative

#### TestGetLeasesDueForMonth_MidMonthLease
- **Purpose**: Validates days overdue calculation for mid-month leases
- **Setup**: Creates lease starting on 15th of current month
- **Verifies**: Days overdue calculated from lease start date

### 6. Sorting Tests

#### TestGetLeasesDueForMonth_Sorting
- **Purpose**: Validates result sorting logic
- **Setup**: Creates 3 leases with varying overdue days and tenant names
- **Verifies**: 
  - Sorted by days overdue DESC
  - Sorted by tenant name ASC (as secondary criteria)
  - Alice (10 days) > Charlie (10 days) > Bob (5 days)

### 7. Edge Case Tests

#### TestGetLeasesDueForMonth_EmptyResult
- **Purpose**: Tests empty result scenario
- **Setup**: Creates organization with no leases
- **Verifies**: 
  - Returns empty slice (not nil)
  - No errors occurred

## Test Utilities

The test suite includes comprehensive helper functions:

### setupLeaseRepository(t *testing.T)
- Creates test database with migrations
- Initializes LeaseRepository
- Returns repo, db, and cleanup function

### createTestOrganization(t, db)
- Creates test organization record
- Returns organization ID

### createTestPropertyWithOrg(t, db, orgID)
- Creates test property linked to organization
- Returns property ID

### createTestBuilding(t, db, propertyID)
- Creates test building linked to property
- Returns building ID

### createTestUnit(t, db, propertyID, buildingID)
- Creates test unit linked to building
- Returns unit ID

### createTestTenant(t, db, orgID)
- Creates test tenant linked to organization
- Returns tenant ID

### createTestLease(t, db, unitID, tenantID, orgID, rent, startDate, endDate, active)
- Creates test lease with all parameters
- Returns lease ID

### createTestPayment(t, db, unitID, tenantID, month, year, status, amount)
- Creates test payment with all parameters
- Returns payment ID

## Test Database Setup

Tests use the existing `testutil.SetupTestDB` which:
1. Creates a test database connection
2. Runs all database migrations
3. Provides a cleanup function to drop tables
4. Automatically skips tests if PostgreSQL is unavailable

## Running the Tests

### Run all GetLeasesDueForMonth tests:
```bash
cd src/backend/api
go test -v ./internal/repositories -run TestGetLeasesDueForMonth
```

### Run a specific test:
```bash
go test -v ./internal/repositories -run TestGetLeasesDueForMonth_Success
```

### Run with coverage:
```bash
go test -cover ./internal/repositories -run TestGetLeasesDueForMonth
```

## Test Dependencies

- `database/sql` - Database operations
- `testing` - Go testing framework
- `time` - Time manipulation for date calculations
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/ysnarafat/tenantly/internal/testutil` - Test utilities

## Key Assertions

The tests verify:

1. **Correctness**: Expected results match actual results
2. **Error Handling**: No unexpected errors occur
3. **Data Integrity**: All fields are correctly populated
4. **Business Logic**: 
   - Active, current leases are included
   - Paid/Partial payments exclude leases
   - Inactive/expired leases are excluded
   - Organization scoping works correctly
5. **Sorting**: Results are properly ordered
6. **Edge Cases**: Empty results, invalid IDs handled gracefully

## Test Maintenance

When modifying `GetLeasesDueForMonth`:
1. Review all test cases to understand expected behavior
2. Add new tests for new functionality
3. Update existing tests if behavior changes
4. Ensure all tests pass before committing
5. Consider adding tests for edge cases if modifications introduce new scenarios

## Coverage

The test suite provides comprehensive coverage of:
- ✅ Happy path scenarios
- ✅ Payment status filtering
- ✅ Lease status filtering
- ✅ Organization scoping
- ✅ Date-based filtering
- ✅ Days overdue calculation
- ✅ Result sorting
- ✅ Data integrity
- ✅ Edge cases
- ✅ Error handling

## Notes

- Tests are integration tests that use a real PostgreSQL database
- Each test runs in isolation with its own database setup/cleanup
- Tests can be skipped if PostgreSQL is not available
- All helper functions are reusable across different test scenarios