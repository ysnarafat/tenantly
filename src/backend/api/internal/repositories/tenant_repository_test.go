package repositories

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/ysnarafat/tenantly/internal/models"
)

func TestTenantRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewTenantRepository(db)

	t.Run("success", func(t *testing.T) {
		req := &models.CreateTenantRequest{
			Name:        "John Doe",
			TenantType:  models.TenantTypeIndividual,
			PhoneNumber: "1234567890",
			Email:       "john@example.com",
			NIDNumber:   "NID123",
			Address:     "123 Main St",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`
			INSERT INTO tenants (
				name, tenant_type, phone_number, email, nid_number, address, active
			) VALUES ($1, $2, $3, $4, $5, $6, true)
			RETURNING id, created_at, updated_at`)).
			WithArgs(
				req.Name, req.TenantType, req.PhoneNumber, req.Email, req.NIDNumber, req.Address,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(1, time.Now(), time.Now()))

		tenant, err := repo.Create(req)
		assert.NoError(t, err)
		assert.NotNil(t, tenant)
		assert.Equal(t, 1, tenant.ID)
		assert.Equal(t, req.Name, tenant.Name)
	})

	t.Run("database error", func(t *testing.T) {
		req := &models.CreateTenantRequest{Name: "John Doe"}

		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO tenants")).
			WillReturnError(assert.AnError)

		tenant, err := repo.Create(req)
		assert.Error(t, err)
		assert.Nil(t, tenant)
	})
}

func TestTenantRepository_CheckEmailExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewTenantRepository(db)

	t.Run("exists", func(t *testing.T) {
		email := "john@example.com"
		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
			WithArgs(email, 0).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		exists, err := repo.CheckEmailExists(email, 0)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("does not exist", func(t *testing.T) {
		email := "new@example.com"
		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
			WithArgs(email, 0).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		exists, err := repo.CheckEmailExists(email, 0)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestTenantRepository_CheckNIDExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewTenantRepository(db)

	t.Run("exists", func(t *testing.T) {
		nid := "NID123"
		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
			WithArgs(nid, 0).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		exists, err := repo.CheckNIDExists(nid, 0)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("does not exist", func(t *testing.T) {
		nid := "NID999"
		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
			WithArgs(nid, 0).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		exists, err := repo.CheckNIDExists(nid, 0)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}
