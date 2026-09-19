package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// ctxWithQuery builds a gin context whose request carries the given raw query string.
func ctxWithQuery(rawQuery string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/?"+rawQuery, nil)
	return c
}

func TestParseDateRange_ExplicitDates(t *testing.T) {
	c := ctxWithQuery("start_date=2025-03-01&end_date=2025-03-31")

	start, end, err := parseDateRange(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := start.Format("2006-01-02 15:04:05"); got != "2025-03-01 00:00:00" {
		t.Errorf("start = %q, want 2025-03-01 00:00:00", got)
	}
	// endDate is pushed to the last second of the day so the range is inclusive.
	if got := end.Format("2006-01-02 15:04:05"); got != "2025-03-31 23:59:59" {
		t.Errorf("end = %q, want 2025-03-31 23:59:59", got)
	}
}

func TestParseDateRange_DefaultsToLastMonth(t *testing.T) {
	now := time.Now()
	c := ctxWithQuery("")

	start, end, err := parseDateRange(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := start.Format("2006-01-02"), now.AddDate(0, -1, 0).Format("2006-01-02"); got != want {
		t.Errorf("default start = %q, want %q (one month ago)", got, want)
	}
	if got, want := end.Format("2006-01-02"), now.Format("2006-01-02"); got != want {
		t.Errorf("default end date = %q, want %q (today)", got, want)
	}
	if h, m, s := end.Hour(), end.Minute(), end.Second(); h != 23 || m != 59 || s != 59 {
		t.Errorf("default end time = %02d:%02d:%02d, want 23:59:59", h, m, s)
	}
	if !end.After(start) {
		t.Errorf("end (%v) should be after start (%v)", end, start)
	}
}

func TestParseDateRange_InvalidDates(t *testing.T) {
	cases := []struct {
		name    string
		query   string
		wantSub string
	}{
		{"bad start", "start_date=not-a-date", "start_date"},
		{"bad end", "start_date=2025-01-01&end_date=2025-13-40", "end_date"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := parseDateRange(ctxWithQuery(tc.query))
			if err == nil {
				t.Fatalf("expected an error for query %q", tc.query)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("error %q should mention %q", err.Error(), tc.wantSub)
			}
		})
	}
}

func TestCanCreateRole(t *testing.T) {
	cases := []struct {
		caller string
		target string
		want   bool
	}{
		// SUPER_ADMIN can create anyone.
		{"SUPER_ADMIN", "ORG_ADMIN", true},
		{"SUPER_ADMIN", "Admin", true},
		{"SUPER_ADMIN", "Accountant", true},
		// ORG_ADMIN can create org-scoped roles but not admins above it.
		{"ORG_ADMIN", "Admin", true},
		{"ORG_ADMIN", "PropertyManager", true},
		{"ORG_ADMIN", "Accountant", true},
		{"ORG_ADMIN", "ORG_ADMIN", false},
		{"ORG_ADMIN", "SUPER_ADMIN", false},
		// Admin can create only the two lowest roles.
		{"Admin", "PropertyManager", true},
		{"Admin", "Accountant", true},
		{"Admin", "Admin", false},
		{"Admin", "ORG_ADMIN", false},
		// Non-privileged roles cannot create anyone.
		{"PropertyManager", "Accountant", false},
		{"Accountant", "Accountant", false},
		{"", "Accountant", false},
	}
	for _, tc := range cases {
		if got := canCreateRole(tc.caller, tc.target); got != tc.want {
			t.Errorf("canCreateRole(%q, %q) = %v, want %v", tc.caller, tc.target, got, tc.want)
		}
	}
}

func TestCallerOrgID(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		orgID := 42
		c.Set("organization_id", &orgID)

		got, ok := callerOrgID(c)
		if !ok || got != 42 {
			t.Errorf("callerOrgID = (%d, %v), want (42, true)", got, ok)
		}
	})

	t.Run("absent", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		if got, ok := callerOrgID(c); ok || got != 0 {
			t.Errorf("callerOrgID = (%d, %v), want (0, false)", got, ok)
		}
	})

	t.Run("nil pointer (e.g. SUPER_ADMIN)", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		var nilPtr *int
		c.Set("organization_id", nilPtr)
		if got, ok := callerOrgID(c); ok || got != 0 {
			t.Errorf("callerOrgID = (%d, %v), want (0, false)", got, ok)
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("organization_id", "not-an-int-ptr")
		if got, ok := callerOrgID(c); ok || got != 0 {
			t.Errorf("callerOrgID = (%d, %v), want (0, false)", got, ok)
		}
	})
}
