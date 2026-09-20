package repositories

import (
	"crypto/sha256"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	appcrypto "github.com/ysnarafat/tenantly/internal/crypto"
	"github.com/ysnarafat/tenantly/internal/models"
)

func newSqlxMock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	return sqlx.NewDb(sqlDB, "postgres"), mock
}

func newTestNIDProtector(t *testing.T) *appcrypto.NIDProtector {
	t.Helper()
	key := sha256.Sum256([]byte("tenant-repo-test-key"))
	prot, err := appcrypto.NewNIDProtector(key[:], []byte("tenant-repo-test-pepper"))
	assert.NoError(t, err)
	return prot
}

func TestTenantRepository_Create(t *testing.T) {
	db, mock := newSqlxMock(t)

	repo := NewTenantRepository(db, newTestNIDProtector(t))

	t.Run("success", func(t *testing.T) {
		req := &models.CreateTenantRequest{
			Name:        "John Doe",
			TenantType:  models.TenantTypeIndividual,
			PhoneNumber: "1234567890",
			Email:       "john@example.com",
			NIDNumber:   "NID123",
			Address:     "123 Main St",
		}

		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO tenants")).
			WithArgs(
				// NID is stored as ciphertext, last-four, and hash; the ciphertext
				// carries a random nonce, so match the NID-derived args loosely.
				req.Name, req.TenantType, req.PhoneNumber, req.Email,
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), req.Address, 0,
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
	db, mock := newSqlxMock(t)

	repo := NewTenantRepository(db, newTestNIDProtector(t))

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
	db, mock := newSqlxMock(t)

	repo := NewTenantRepository(db, newTestNIDProtector(t))

	t.Run("exists", func(t *testing.T) {
		nid := "NID123"
		// Lookup is by deterministic hash, not the plaintext NID.
		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
			WithArgs(sqlmock.AnyArg(), 0).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		exists, err := repo.CheckNIDExists(nid, 0)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("does not exist", func(t *testing.T) {
		nid := "NID999"
		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
			WithArgs(sqlmock.AnyArg(), 0).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		exists, err := repo.CheckNIDExists(nid, 0)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}
