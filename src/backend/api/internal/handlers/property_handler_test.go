package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/database"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/repositories"
	"github.com/ysnarafat/tenantly/internal/services"
	"github.com/ysnarafat/tenantly/internal/testutil"

	_ "github.com/lib/pq"
)

// handlersTestDB is this package's own database. See testutil.SetupTestDBNamed
// for why it cannot share one with the repository tests.
const handlersTestDB = "tenantly_test_handlers"

// testUserPassword is the plaintext the integration suites log in with; their
// seeded users must carry its bcrypt digest.
const testUserPassword = "password"

func setupPropertyHandlerTestDB(t *testing.T) (*sqlx.DB, func()) {
	// Migrate rather than hand-roll the schema: the tables these handlers
	// write to are org-scoped, and a bespoke CREATE TABLE drifts from the
	// migrations the moment a column is added.
	return testutil.SetupTestDBNamed(t, handlersTestDB)
}

func setupPropertyHandler(t *testing.T) (*PropertyHandler, *gin.Engine, func()) {
	db, cleanup := setupPropertyHandlerTestDB(t)

	orgID := testutil.CreateTestOrganization(t, db)
	userID := testutil.CreateTestUser(t, db)

	propertyRepo := repositories.NewPropertyRepository(db)
	auditService := database.NewAuditService(db)
	propertyService := services.NewPropertyService(propertyRepo, auditService)
	handler := NewPropertyHandler(propertyService)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Stand in for AuthRequired + RequireOrgContext, which every real
	// property route sits behind.
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("org_id", orgID)
		c.Next()
	})

	return handler, router, cleanup
}

func TestPropertyHandler_CreateProperty(t *testing.T) {
	handler, router, cleanup := setupPropertyHandler(t)
	defer cleanup()

	router.POST("/properties", handler.CreateProperty)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid property creation",
			requestBody: models.CreatePropertyRequest{
				PropertyName: "Test Property",
				PropertyCode: "TEST001",
				Address:      "123 Test Street",
				City:         "Test City",
				PostalCode:   "12345",
				PropertyType: models.PropertyTypeCommercial,
				Metadata:     models.PropertyMetadata{"test": "value"},
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "Invalid request body",
			requestBody: map[string]interface{}{
				"property_name": "", // Empty name should fail validation
				"property_code": "TEST002",
				"address":       "456 Test Street",
				"property_type": "Commercial",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			// Rejected by the binding's `oneof` tag before the handler runs,
			// so this is a 400, not the 500 this case originally expected.
			name: "Invalid property type",
			requestBody: models.CreatePropertyRequest{
				PropertyName: "Test Property 2",
				PropertyCode: "TEST003",
				Address:      "789 Test Street",
				PropertyType: "InvalidType",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "Duplicate property name",
			requestBody: models.CreatePropertyRequest{
				PropertyName: "Test Property", // Same as first test
				PropertyCode: "TEST004",
				Address:      "101 Test Street",
				PropertyType: models.PropertyTypeResidential,
			},
			expectedStatus: http.StatusConflict,
			expectError:    true,
		},
		{
			name: "Address exceeds max length",
			requestBody: models.CreatePropertyRequest{
				PropertyName: "Test Property 5",
				PropertyCode: "TEST005",
				Address:      strings.Repeat("a", 501),
				PropertyType: models.PropertyTypeResidential,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/properties", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.expectError {
				if _, exists := response["error"]; !exists {
					t.Errorf("Expected error in response but got none")
				}
			} else {
				// Create/Update return the property itself, not a wrapper.
				if _, exists := response["id"]; !exists {
					t.Errorf("Expected property in response but got none")
				}
			}
		})
	}
}

func TestPropertyHandler_GetProperties(t *testing.T) {
	handler, router, cleanup := setupPropertyHandler(t)
	defer cleanup()

	router.GET("/properties", handler.GetProperties)
	router.POST("/properties", handler.CreateProperty)

	// Create test properties
	properties := []models.CreatePropertyRequest{
		{
			PropertyName: "Commercial Property",
			PropertyCode: "COM001",
			Address:      "123 Commercial St",
			PropertyType: models.PropertyTypeCommercial,
		},
		{
			PropertyName: "Residential Property",
			PropertyCode: "RES001",
			Address:      "456 Residential Ave",
			PropertyType: models.PropertyTypeResidential,
		},
	}

	for _, prop := range properties {
		jsonBody, _ := json.Marshal(prop)
		req, _ := http.NewRequest("POST", "/properties", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Get all properties",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "Filter by property type",
			queryParams:    "?property_type=Commercial",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Search properties",
			queryParams:    "?search=Commercial",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Pagination",
			queryParams:    "?page=1&page_size=1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/properties"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			properties, exists := response["properties"].([]interface{})
			if !exists {
				t.Errorf("Expected properties in response")
				return
			}

			if len(properties) != tt.expectedCount {
				t.Errorf("Expected %d properties, got %d", tt.expectedCount, len(properties))
			}

			// Check pagination info
			if pagination, exists := response["pagination"]; exists {
				paginationMap := pagination.(map[string]interface{})
				if totalItems, exists := paginationMap["total_items"]; exists {
					if totalItems.(float64) < float64(tt.expectedCount) {
						t.Errorf("Expected total_items >= %d, got %v", tt.expectedCount, totalItems)
					}
				}
			}
		})
	}
}

func TestPropertyHandler_GetProperty(t *testing.T) {
	handler, router, cleanup := setupPropertyHandler(t)
	defer cleanup()

	router.POST("/properties", handler.CreateProperty)
	router.GET("/properties/:id", handler.GetProperty)

	// Create a test property
	createReq := models.CreatePropertyRequest{
		PropertyName: "Test Property",
		PropertyCode: "TEST001",
		Address:      "123 Test Street",
		PropertyType: models.PropertyTypeCommercial,
	}

	jsonBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/properties", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &createResponse)
	// CreateProperty returns the property itself, not a wrapper.
	propertyID := int(createResponse["id"].(float64))

	tests := []struct {
		name           string
		propertyID     string
		queryParams    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid property ID",
			propertyID:     strconv.Itoa(propertyID),
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Valid property ID with stats",
			propertyID:     strconv.Itoa(propertyID),
			queryParams:    "?include_stats=true",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid property ID",
			propertyID:     "99999",
			queryParams:    "",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "Non-numeric property ID",
			propertyID:     "invalid",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/properties/"+tt.propertyID+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.expectError {
				if _, exists := response["error"]; !exists {
					t.Errorf("Expected error in response but got none")
				}
			} else {
				if _, exists := response["property"]; !exists {
					t.Errorf("Expected property in response but got none")
				}
			}
		})
	}
}

func TestPropertyHandler_UpdateProperty(t *testing.T) {
	handler, router, cleanup := setupPropertyHandler(t)
	defer cleanup()

	router.POST("/properties", handler.CreateProperty)
	router.PUT("/properties/:id", handler.UpdateProperty)

	// Create a test property
	createReq := models.CreatePropertyRequest{
		PropertyName: "Test Property",
		PropertyCode: "TEST001",
		Address:      "123 Test Street",
		PropertyType: models.PropertyTypeCommercial,
	}

	jsonBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/properties", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &createResponse)
	// CreateProperty returns the property itself, not a wrapper.
	propertyID := int(createResponse["id"].(float64))

	newName := "Updated Property Name"
	newAddress := "Updated Address"

	tests := []struct {
		name           string
		propertyID     string
		requestBody    interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name:       "Valid update",
			propertyID: strconv.Itoa(propertyID),
			requestBody: models.UpdatePropertyRequest{
				PropertyName: &newName,
				Address:      &newAddress,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:       "Invalid property ID",
			propertyID: "99999",
			requestBody: models.UpdatePropertyRequest{
				PropertyName: &newName,
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:       "Non-numeric property ID",
			propertyID: "invalid",
			requestBody: models.UpdatePropertyRequest{
				PropertyName: &newName,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Invalid request body",
			propertyID:     strconv.Itoa(propertyID),
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:       "Address exceeds max length",
			propertyID: strconv.Itoa(propertyID),
			requestBody: models.UpdatePropertyRequest{
				Address: stringPtr(strings.Repeat("a", 501)),
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PUT", "/properties/"+tt.propertyID, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.expectError {
				if _, exists := response["error"]; !exists {
					t.Errorf("Expected error in response but got none")
				}
			} else {
				// Create/Update return the property itself, not a wrapper.
				if _, exists := response["id"]; !exists {
					t.Errorf("Expected property in response but got none")
				}
			}
		})
	}
}

func TestPropertyHandler_DeleteProperty(t *testing.T) {
	handler, router, cleanup := setupPropertyHandler(t)
	defer cleanup()

	router.POST("/properties", handler.CreateProperty)
	router.DELETE("/properties/:id", handler.DeleteProperty)

	// Create a test property
	createReq := models.CreatePropertyRequest{
		PropertyName: "Test Property",
		PropertyCode: "TEST001",
		Address:      "123 Test Street",
		PropertyType: models.PropertyTypeCommercial,
	}

	jsonBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/properties", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &createResponse)
	// CreateProperty returns the property itself, not a wrapper.
	propertyID := int(createResponse["id"].(float64))

	tests := []struct {
		name           string
		propertyID     string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid deletion",
			propertyID:     strconv.Itoa(propertyID),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid property ID",
			propertyID:     "99999",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "Non-numeric property ID",
			propertyID:     "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", "/properties/"+tt.propertyID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.expectError {
				if _, exists := response["error"]; !exists {
					t.Errorf("Expected error in response but got none")
				}
			} else {
				if message, exists := response["message"]; !exists || message != "Property deleted successfully" {
					t.Errorf("Expected success message in response")
				}
			}
		})
	}
}

func TestPropertyHandler_SearchProperties(t *testing.T) {
	handler, router, cleanup := setupPropertyHandler(t)
	defer cleanup()

	router.POST("/properties", handler.CreateProperty)
	router.GET("/properties/search", handler.SearchProperties)

	// Create test properties
	properties := []models.CreatePropertyRequest{
		{
			PropertyName: "Downtown Commercial Center",
			PropertyCode: "DCC001",
			Address:      "123 Downtown St",
			PropertyType: models.PropertyTypeCommercial,
		},
		{
			PropertyName: "Uptown Residential Complex",
			PropertyCode: "URC001",
			Address:      "456 Uptown Ave",
			PropertyType: models.PropertyTypeResidential,
		},
	}

	for _, prop := range properties {
		jsonBody, _ := json.Marshal(prop)
		req, _ := http.NewRequest("POST", "/properties", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Search by name",
			queryParams:    "?q=Downtown",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Search by address",
			queryParams:    "?q=Uptown",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Search with filter",
			queryParams:    "?q=Commercial&property_type=Commercial",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "No results",
			queryParams:    "?q=NonExistent",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/properties/search"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			properties, exists := response["properties"].([]interface{})
			if !exists {
				t.Errorf("Expected properties in response")
				return
			}

			if len(properties) != tt.expectedCount {
				t.Errorf("Expected %d properties, got %d", tt.expectedCount, len(properties))
			}
		})
	}
}
