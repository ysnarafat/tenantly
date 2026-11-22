package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/repositories"
	"github.com/ysnarafat/tenantly/internal/services"
)

func newSupportCaseTestHandler(t *testing.T) (*OperationsHandler, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock database: %v", err)
	}
	repo := repositories.NewOperationsRepository(sqlx.NewDb(db, "postgres"))
	handler := NewOperationsHandler(services.NewOperationsService(repo, nil))
	return handler, mock, func() { _ = db.Close() }
}

func supportCaseTestRouter(handler *OperationsHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 7)
		c.Set("org_id", 10)
		c.Next()
	})
	router.POST("/support/cases", handler.CreateTenantSupportCase)
	router.GET("/support/cases", handler.ListTenantSupportCases)
	router.GET("/support/cases/:case_id", handler.GetTenantSupportCase)
	router.POST("/support/cases/:case_id/messages", handler.CreateTenantSupportCaseMessage)
	return router
}

func TestTenantSupportCaseListUsesAuthenticatedOrganization(t *testing.T) {
	handler, mock, closeDB := newSupportCaseTestHandler(t)
	defer closeDB()

	mock.ExpectQuery("(?s)SELECT c.id.*FROM support_cases c.*WHERE c.organization_id = \\$1").
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "subject", "description", "status", "priority", "created_at", "updated_at",
		}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/support/cases", nil)
	supportCaseTestRouter(handler).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateTenantSupportCaseUsesAuthenticatedOrganization(t *testing.T) {
	handler, mock, closeDB := newSupportCaseTestHandler(t)
	defer closeDB()

	now := time.Now().UTC()
	mock.ExpectQuery("INSERT INTO support_cases").
		WithArgs(10, "Tenant issue", "Please help", 7).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "subject", "description", "status", "priority", "created_at", "updated_at",
		}).AddRow(15, "Tenant issue", "Please help", "open", "normal", now, now))

	router := supportCaseTestRouter(handler)
	request := httptest.NewRequest(http.MethodPost, "/support/cases", strings.NewReader(
		`{"organization_id":99,"priority":"urgent","status":"closed","assigned_to":3,"subject":"Tenant issue","description":"Please help"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["priority"] != "normal" {
		t.Errorf("expected staff-default priority, got %#v", response["priority"])
	}
	if response["status"] != "open" {
		t.Errorf("expected default open status, got %#v", response["status"])
	}
	if _, exists := response["assigned_to"]; exists {
		t.Error("tenant response exposed staff assignment")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTenantCannotReadSupportCaseFromAnotherOrganization(t *testing.T) {
	handler, mock, closeDB := newSupportCaseTestHandler(t)
	defer closeDB()

	mock.ExpectQuery("(?s)SELECT c.id.*WHERE c.id = \\$1 AND c.organization_id = \\$2").
		WithArgs(15, 10).
		WillReturnError(sql.ErrNoRows)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/support/cases/15", nil)
	supportCaseTestRouter(handler).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected cross-organization case to be hidden with 404, got %d", recorder.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTenantCannotReplyToSupportCaseFromAnotherOrganization(t *testing.T) {
	handler, mock, closeDB := newSupportCaseTestHandler(t)
	defer closeDB()

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO support_case_messages").
		WithArgs(15, 7, "Reply", 10).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/support/cases/15/messages", strings.NewReader(`{"body":"Reply"}`))
	request.Header.Set("Content-Type", "application/json")
	supportCaseTestRouter(handler).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected cross-organization reply to be hidden with 404, got %d", recorder.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTenantCaseDetailIncludesMessagesButNotPrivateNotes(t *testing.T) {
	handler, mock, closeDB := newSupportCaseTestHandler(t)
	defer closeDB()

	now := time.Now().UTC()
	mock.ExpectQuery("(?s)SELECT c.id.*WHERE c.id = \\$1 AND c.organization_id = \\$2").
		WithArgs(15, 10).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "subject", "description", "status", "priority", "created_at", "updated_at",
		}).AddRow(15, "Tenant issue", "Original details", "open", "normal", now, now))
	mock.ExpectQuery("(?s)SELECT m.id.*WHERE m.case_id = \\$1 AND c.organization_id = \\$2").
		WithArgs(15, 10).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "case_id", "author_id", "author", "body", "created_at",
		}).AddRow(31, 15, 7, "Tenant Admin", "Additional details", now))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/support/cases/15", nil)
	supportCaseTestRouter(handler).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, exists := response["notes"]; exists {
		t.Fatal("tenant response exposed private notes")
	}
	messages, ok := response["messages"].([]any)
	if !ok || len(messages) != 1 {
		t.Fatalf("expected one customer-visible message, got %#v", response["messages"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
