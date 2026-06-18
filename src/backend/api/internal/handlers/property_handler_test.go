package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/database"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/repositories"
	"github.com/ysnarafat/tenantly/internal/services"

	_ "github.com/lib/pq"
)

func setupPropertyHandlerTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/tenantly_test?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: PostgreSQL not available")
	}

	if err := db.Ping(); err != nil {
		db.Close()
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
	}

	// Create test tables
	createPropertyHandlerTestTables(t, db)

	cleanup := func() {
		dropPropertyHandlerTestTables(t, db)
		db.Close()
	}

	return db, cleanup
}

func createPropertyHandlerTestTables(t *testing.T, db *sql.DB) {
	// Create properties table for testing
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS properties (
			id SERIAL PRIMARY KEY,
			property_name VARCHAR(200) NOT NULL,
			property_code VARCHAR(50) UNIQUE NOT NULL,
			address TEXT NOT NULL,
			city VARCHAR(100),
			postal_code VARCHAR(20),
			property_type VARCHAR(50) NOT NULL CHECK (property_type IN ('Residential', 'Commercial', 'Mixed')),
			total_buildings INTEGER DEFAULT 1,
			metadata JSONB,
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
			updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create properties table: %v", err)
	}

	// Create audit_log table for testing
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_log (
			id SERIAL PRIMARY KEY,
			user_id INTEGER,
			action VARCHAR(50) NOT NULL,
			table_name VARCHAR(50) NOT NULL,
			record_id INTEGER,
			old_values JSONB,
			new_values JSONB,
			created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create audit_log table: %v", err)
	}
}

func dropPropertyHandlerTestTables(t *testing.T, db *sql.DB) {
	tables := []string{"properties", "audit_log"}
	for _, table := range tables {
		_, err := db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE")
		if err != nil {
			t.Logf("Warning: Failed to drop table %s: %v", table, err)
		}
	}
}

func setupPropertyHandler(t *testing.T) (*PropertyHandler, *gin.Engine, func()) {
	db, cleanup := setupPropertyHandlerTestDB(t)

	propertyRepo := repositories.NewPropertyRepository(db)
	auditService := database.NewAuditService(db)
	propertyService := services.NewPropertyService(propertyRepo, auditService)
	handler := NewPropertyHandler(propertyService)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware to set user_id in context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
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
			name: "Invalid property type",
			requestBody: models.CreatePropertyRequest{
				PropertyName: "Test Property 2",
				PropertyCode: "TEST003",
				Address:      "789 Test Street",
				PropertyType: "InvalidType",
			},
			expectedStatus: http.StatusInternalServerError,
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
				if _, exists := response["property"]; !exists {
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
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	property := createResponse["property"].(map[string]interface{})
	propertyID := int(property["id"].(float64))

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
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	property := createResponse["property"].(map[string]interface{})
	propertyID := int(property["id"].(float64))

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
				if _, exists := response["property"]; !exists {
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
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	property := createResponse["property"].(map[string]interface{})
	propertyID := int(property["id"].(float64))

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
