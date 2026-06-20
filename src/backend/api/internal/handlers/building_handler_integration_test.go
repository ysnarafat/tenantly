package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ysnarafat/tenantly/internal/models"
)

func TestBuildingHandler_EndpointRouting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBuildingService)
	handler := NewBuildingHandler(mockService)

	router := gin.New()

	// Register building routes
	buildings := router.Group("/api/v1/buildings")
	{
		buildings.GET("", handler.GetBuildings)
		buildings.POST("", handler.CreateBuilding)
		buildings.GET("/search", handler.SearchBuildings)
		buildings.GET("/:id", handler.GetBuilding)
		buildings.PUT("/:id", handler.UpdateBuilding)
		buildings.DELETE("/:id", handler.DeleteBuilding)
		buildings.GET("/:id/analytics", handler.GetBuildingAnalytics)
	}

	// Register property-building routes with mock middleware
	properties := router.Group("/api/v1/properties")
	{
		// Mock middleware that sets property in context
		properties.Use(func(c *gin.Context) {
			// Mock property for testing
			property := &models.Property{
				ID:           1,
				PropertyName: "Test Property",
				PropertyCode: "TEST001",
				PropertyType: models.PropertyTypeCommercial,
				Address:      "Test Address",
				Active:       true,
			}
			c.Set("property", property)
			c.Set("property_id", 1)
			c.Next()
		})
		properties.GET("/:id/buildings", handler.GetPropertyBuildings)
		properties.POST("/:id/buildings/bulk", handler.BulkCreateBuildings)
	}

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		setupMock      func()
	}{
		{
			name:           "GET /api/v1/buildings",
			method:         "GET",
			path:           "/api/v1/buildings",
			expectedStatus: http.StatusOK,
			setupMock: func() {
				mockService.On("SearchBuildings", mock.AnythingOfType("*models.BuildingSearchFilters")).Return([]*models.Building{}, nil)
			},
		},
		{
			name:           "GET /api/v1/buildings/search",
			method:         "GET",
			path:           "/api/v1/buildings/search",
			expectedStatus: http.StatusOK,
			setupMock: func() {
				mockService.On("SearchBuildings", mock.AnythingOfType("*models.BuildingSearchFilters")).Return([]*models.Building{}, nil)
			},
		},
		{
			name:           "GET /api/v1/buildings/1",
			method:         "GET",
			path:           "/api/v1/buildings/1",
			expectedStatus: http.StatusOK,
			setupMock: func() {
				building := &models.Building{ID: 1, BuildingName: "Test Building"}
				mockService.On("GetBuilding", 1, 0).Return(building, nil)
			},
		},
		{
			name:           "GET /api/v1/buildings/1/analytics",
			method:         "GET",
			path:           "/api/v1/buildings/1/analytics",
			expectedStatus: http.StatusOK,
			setupMock: func() {
				analytics := &models.BuildingAnalytics{BuildingID: 1}
				mockService.On("GetBuildingAnalytics", 1, 0).Return(analytics, nil)
			},
		},
		{
			name:           "GET /api/v1/properties/1/buildings",
			method:         "GET",
			path:           "/api/v1/properties/1/buildings",
			expectedStatus: http.StatusOK,
			setupMock: func() {
				// Mock the new method that's called by the enhanced handler
				response := &models.BuildingListResponse{
					Buildings: []*models.Building{},
					Pagination: &models.PaginationInfo{
						CurrentPage: 1,
						PageSize:    20,
						TotalItems:  0,
						TotalPages:  0,
						HasNext:     false,
						HasPrev:     false,
					},
				}
				mockService.On("GetPropertyBuildingsWithPagination", 1, mock.AnythingOfType("*models.BuildingSearchFilters"), "created_at", "desc", false).Return(response, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			// Setup mock expectations
			tt.setupMock()

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code, "Expected status %d, got %d for %s %s", tt.expectedStatus, w.Code, tt.method, tt.path)

			// Verify mock expectations were met
			mockService.AssertExpectations(t)
		})
	}
}

func TestBuildingHandler_CreateBuildingEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBuildingService)
	handler := NewBuildingHandler(mockService)

	router := gin.New()
	router.POST("/api/v1/buildings", handler.CreateBuilding)

	// Mock request
	req := &models.CreateBuildingRequest{
		PropertyID:   1,
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  5,
		HasElevator:  true,
	}

	// Mock response
	expectedBuilding := &models.Building{
		ID:           1,
		PropertyID:   1,
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  5,
		HasElevator:  true,
		ActiveStatus: true,
	}

	mockService.On("CreateBuilding", req).Return(expectedBuilding, nil)

	// Create HTTP request
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/buildings", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Building created successfully", response["message"])
	assert.NotNil(t, response["building"])

	mockService.AssertExpectations(t)
}

