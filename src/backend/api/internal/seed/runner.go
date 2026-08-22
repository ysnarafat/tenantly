package seed

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	appcrypto "github.com/ysnarafat/tenantly/internal/crypto"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

// Runner seeds a database from JSON files under seeds/{env}/.
type Runner struct {
	db          *sqlx.DB
	orgRepo     *repositories.OrganizationRepository
	userRepo    *repositories.UserRepository
	userOrgRepo *repositories.UserOrganizationRoleRepository
	propRepo    *repositories.PropertyRepository
	buildRepo   *repositories.BuildingRepository
	unitRepo    *repositories.UnitRepository
	tenantRepo  *repositories.TenantRepository
	leaseRepo   *repositories.LeaseRepository
}

func New(db *sqlx.DB, nid *appcrypto.NIDProtector) *Runner {
	return &Runner{
		db:          db,
		orgRepo:     repositories.NewOrganizationRepository(db),
		userRepo:    repositories.NewUserRepository(db),
		userOrgRepo: repositories.NewUserOrganizationRoleRepository(db),
		propRepo:    repositories.NewPropertyRepository(db),
		buildRepo:   repositories.NewBuildingRepository(db),
		unitRepo:    repositories.NewUnitRepository(db),
		tenantRepo:  repositories.NewTenantRepository(db, nid),
		leaseRepo:   repositories.NewLeaseRepository(db),
	}
}

// Run seeds all entities from the given directory in dependency order.
// Each step is idempotent: existing records are reused, not duplicated.
func (r *Runner) Run(seedDir string) error {
	log.Printf("[seed] loading from %s", seedDir)

	orgsBySlug, err := r.seedOrgs(seedDir)
	if err != nil {
		return err
	}
	if err := r.seedUsers(seedDir, orgsBySlug); err != nil {
		return err
	}
	propsByCode, err := r.seedProperties(seedDir, orgsBySlug)
	if err != nil {
		return err
	}
	names, err := readJSON[tenantNames](filepath.Join(seedDir, "tenant_names.json"))
	if err != nil {
		return fmt.Errorf("tenant_names.json: %w", err)
	}
	return r.seedBuildings(seedDir, orgsBySlug, propsByCode, &names)
}

// ── JSON seed types ──────────────────────────────────────────────────────────

type seedOrg struct {
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	SubscriptionTier string `json:"subscription_tier"`
}

type seedUser struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	OrgSlug   string `json:"org_slug"`
}

type seedProperty struct {
	OrgSlug      string `json:"org_slug"`
	PropertyName string `json:"property_name"`
	PropertyCode string `json:"property_code"`
	Address      string `json:"address"`
	City         string `json:"city"`
	PostalCode   string `json:"postal_code"`
	PropertyType string `json:"property_type"`
}

type unitGenConfig struct {
	Floors        int     `json:"floors"`
	UnitsPerFloor int     `json:"units_per_floor"`
	UnitType      string  `json:"unit_type"`
	BaseRent      float64 `json:"base_rent"`
	RentPerFloor  float64 `json:"rent_per_floor"`
	RentPerUnit   float64 `json:"rent_per_unit"`
	AreaBaseSqft  int     `json:"area_base_sqft"`
	AreaPerUnit   int     `json:"area_per_unit"`
	LeaseStart    string  `json:"lease_start"`
	LeaseEnd      string  `json:"lease_end"`
	LeaseMonths   int     `json:"lease_months"`
	LeaseType     string  `json:"lease_type"`
}

type seedBuilding struct {
	OrgSlug          string         `json:"org_slug"`
	PropertyCode     string         `json:"property_code"`
	BuildingName     string         `json:"building_name"`
	BuildingCode     string         `json:"building_code"`
	BuildingType     string         `json:"building_type"`
	TotalFloors      int            `json:"total_floors"`
	HasElevator      bool           `json:"has_elevator"`
	ConstructionYear int            `json:"construction_year"`
	ActiveStatus     bool           `json:"active_status"`
	GenerateUnits    *unitGenConfig `json:"generate_units"`
}

type tenantNames struct {
	FirstNames    []string `json:"first_names"`
	LastNames     []string `json:"last_names"`
	PhonePrefixes []string `json:"phone_prefixes"`
}

// ── Seeder steps ─────────────────────────────────────────────────────────────

func (r *Runner) seedOrgs(dir string) (map[string]*models.Organization, error) {
	orgs, err := readJSON[[]seedOrg](filepath.Join(dir, "organizations.json"))
	if err != nil {
		return nil, fmt.Errorf("organizations.json: %w", err)
	}

	result := make(map[string]*models.Organization, len(orgs))
	for _, o := range orgs {
		existing, err := r.orgRepo.GetBySlug(o.Slug)
		if err != nil && !isNotFound(err) {
			return nil, fmt.Errorf("lookup org %q: %w", o.Slug, err)
		}
		if existing != nil {
			log.Printf("[seed] org %q already exists (id=%d)", o.Slug, existing.ID)
			result[o.Slug] = existing
			continue
		}

		org := &models.Organization{
			Name:             o.Name,
			Slug:             o.Slug,
			SubscriptionTier: models.SubscriptionTier(o.SubscriptionTier),
			Active:           true,
		}
		if err := r.orgRepo.Create(org); err != nil {
			return nil, fmt.Errorf("create org %q: %w", o.Slug, err)
		}
		log.Printf("[seed] created org %q (id=%d)", o.Slug, org.ID)
		result[o.Slug] = org
	}
	return result, nil
}

func (r *Runner) seedUsers(dir string, orgsBySlug map[string]*models.Organization) error {
	users, err := readJSON[[]seedUser](filepath.Join(dir, "users.json"))
	if err != nil {
		return fmt.Errorf("users.json: %w", err)
	}

	for _, u := range users {
		existing, err := r.userRepo.GetByUsername(u.Username)
		if err != nil && !isNotFound(err) {
			return fmt.Errorf("lookup user %q: %w", u.Username, err)
		}
		if existing != nil {
			log.Printf("[seed] user %q already exists (id=%d)", u.Username, existing.ID)
			if u.OrgSlug != "" {
				org := orgsBySlug[u.OrgSlug]
				if err := r.userOrgRepo.Upsert(existing.ID, org.ID, u.Role); err != nil {
					return fmt.Errorf("upsert org role for %q: %w", u.Username, err)
				}
			}
			continue
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
		if err != nil {
			return fmt.Errorf("hash password for %q: %w", u.Username, err)
		}

		var orgID *int
		if u.OrgSlug != "" {
			id := orgsBySlug[u.OrgSlug].ID
			orgID = &id
		}

		user := &models.User{
			Username:       u.Username,
			Email:          u.Email,
			PasswordHash:   string(hash),
			Role:           u.Role,
			Active:         true,
			FirstName:      u.FirstName,
			LastName:       u.LastName,
			OrganizationID: orgID,
			Status:         "active",
		}
		if err := r.userRepo.Create(user); err != nil {
			return fmt.Errorf("create user %q: %w", u.Username, err)
		}
		log.Printf("[seed] created user %q (id=%d)", u.Username, user.ID)

		if u.OrgSlug != "" {
			org := orgsBySlug[u.OrgSlug]
			if err := r.userOrgRepo.Upsert(user.ID, org.ID, u.Role); err != nil {
				return fmt.Errorf("upsert org role for %q: %w", u.Username, err)
			}
		}
	}
	return nil
}

func (r *Runner) seedProperties(dir string, orgsBySlug map[string]*models.Organization) (map[string]*models.Property, error) {
	props, err := readJSON[[]seedProperty](filepath.Join(dir, "properties.json"))
	if err != nil {
		return nil, fmt.Errorf("properties.json: %w", err)
	}

	result := make(map[string]*models.Property, len(props))
	for _, p := range props {
		org := orgsBySlug[p.OrgSlug]
		existing, err := r.findPropertyByCode(org.ID, p.PropertyCode)
		if err != nil {
			return nil, fmt.Errorf("lookup property %q: %w", p.PropertyCode, err)
		}
		if existing != nil {
			log.Printf("[seed] property %q already exists (id=%d)", p.PropertyCode, existing.ID)
			result[p.PropertyCode] = existing
			continue
		}

		req := &models.CreatePropertyRequest{
			PropertyName:   p.PropertyName,
			PropertyCode:   p.PropertyCode,
			Address:        p.Address,
			City:           p.City,
			PostalCode:     p.PostalCode,
			PropertyType:   models.PropertyType(p.PropertyType),
			OrganizationID: org.ID,
		}
		created, err := r.propRepo.Create(req)
		if err != nil {
			return nil, fmt.Errorf("create property %q: %w", p.PropertyCode, err)
		}
		log.Printf("[seed] created property %q (id=%d)", p.PropertyCode, created.ID)
		result[p.PropertyCode] = created
	}
	return result, nil
}

func (r *Runner) seedBuildings(
	dir string,
	orgsBySlug map[string]*models.Organization,
	propsByCode map[string]*models.Property,
	names *tenantNames,
) error {
	buildings, err := readJSON[[]seedBuilding](filepath.Join(dir, "buildings.json"))
	if err != nil {
		return fmt.Errorf("buildings.json: %w", err)
	}

	tenantIdx := 0
	for _, b := range buildings {
		org := orgsBySlug[b.OrgSlug]
		prop := propsByCode[b.PropertyCode]

		existing, err := r.buildRepo.GetByPropertyAndCode(prop.ID, b.BuildingCode)
		if err != nil && !isNotFound(err) {
			return fmt.Errorf("lookup building %q: %w", b.BuildingCode, err)
		}
		if existing != nil {
			log.Printf("[seed] building %q already exists (id=%d), skipping units", b.BuildingCode, existing.ID)
			if b.GenerateUnits != nil {
				tenantIdx += b.GenerateUnits.Floors * b.GenerateUnits.UnitsPerFloor
			}
			continue
		}

		year := b.ConstructionYear
		bldg := &models.Building{
			PropertyID:       prop.ID,
			OrganizationID:   org.ID,
			BuildingName:     b.BuildingName,
			BuildingCode:     b.BuildingCode,
			BuildingType:     models.BuildingType(b.BuildingType),
			TotalFloors:      b.TotalFloors,
			HasElevator:      b.HasElevator,
			ConstructionYear: &year,
			ActiveStatus:     b.ActiveStatus,
		}
		if err := r.buildRepo.Create(bldg); err != nil {
			return fmt.Errorf("create building %q: %w", b.BuildingCode, err)
		}
		log.Printf("[seed] created building %q (id=%d)", b.BuildingCode, bldg.ID)

		if b.GenerateUnits == nil {
			continue
		}
		cfg := b.GenerateUnits
		for floor := 1; floor <= cfg.Floors; floor++ {
			for unitNum := 1; unitNum <= cfg.UnitsPerFloor; unitNum++ {
				if err := r.seedUnit(bldg, prop, org, cfg, floor, unitNum, tenantIdx, names); err != nil {
					return err
				}
				tenantIdx++
			}
		}
		log.Printf("[seed] building %q: %d units seeded", b.BuildingCode, cfg.Floors*cfg.UnitsPerFloor)
	}
	return nil
}

func (r *Runner) seedUnit(
	bldg *models.Building,
	prop *models.Property,
	org *models.Organization,
	cfg *unitGenConfig,
	floor, unitNum, tenantIdx int,
	names *tenantNames,
) error {
	unitNumber := fmt.Sprintf("%d%02d", floor, unitNum)
	unitName := fmt.Sprintf("Flat %d-%02d", floor, unitNum)
	rent := cfg.BaseRent + float64(floor)*cfg.RentPerFloor + float64(unitNum)*cfg.RentPerUnit
	areaSqft := cfg.AreaBaseSqft + unitNum*cfg.AreaPerUnit

	unitReq := &models.CreateUnitRequest{
		BuildingID: bldg.ID,
		PropertyID: prop.ID,
		UnitNumber: unitNumber,
		UnitName:   unitName,
		Floor:      floor,
		UnitType:   models.UnitType(cfg.UnitType),
		Metadata: models.UnitMetadata{
			"bedrooms":  2,
			"bathrooms": 1,
			"area_sqft": areaSqft,
		},
	}
	unit, err := r.unitRepo.Create(unitReq, org.ID)
	if err != nil {
		return fmt.Errorf("create unit %s/%s: %w", bldg.BuildingCode, unitNumber, err)
	}

	firstName := names.FirstNames[tenantIdx%len(names.FirstNames)]
	lastName := names.LastNames[(tenantIdx/len(names.FirstNames))%len(names.LastNames)]
	fullName := firstName + " " + lastName
	phone := fmt.Sprintf("%s%07d", names.PhonePrefixes[tenantIdx%len(names.PhonePrefixes)], tenantIdx+1000000)
	email := fmt.Sprintf("tenant.%d@example.com", tenantIdx+1)
	nid := fmt.Sprintf("%010d", 1000000000+tenantIdx)

	tenantReq := &models.CreateTenantRequest{
		Name:           fullName,
		TenantType:     models.TenantTypeIndividual,
		PhoneNumber:    phone,
		Email:          email,
		NIDNumber:      nid,
		Address:        fmt.Sprintf("Flat %s, %s, Mirpur-10, Dhaka", unitNumber, bldg.BuildingName),
		OrganizationID: org.ID,
	}
	tenant, err := r.tenantRepo.Create(tenantReq)
	if err != nil {
		return fmt.Errorf("create tenant %s: %w", email, err)
	}

	endDate := cfg.LeaseEnd
	leaseReq := &models.CreateLeaseRequest{
		UnitID:          unit.ID,
		TenantID:        tenant.ID,
		LeaseType:       models.LeaseType(cfg.LeaseType),
		StartDate:       cfg.LeaseStart,
		EndDate:         &endDate,
		DurationMonths:  cfg.LeaseMonths,
		MonthlyRent:     rent,
		SecurityDeposit: rent * 2,
		OrganizationID:  org.ID,
	}
	lease, err := r.leaseRepo.Create(leaseReq)
	if err != nil {
		return fmt.Errorf("create lease for unit %s: %w", unitNumber, err)
	}

	log.Printf("[seed]   unit %s | %s | rent=%.0f | lease=%d", unitNumber, fullName, rent, lease.ID)
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// findPropertyByCode returns the property with the given code in the org, or nil if not found.
func (r *Runner) findPropertyByCode(orgID int, code string) (*models.Property, error) {
	props, _, err := r.propRepo.List(map[string]any{"organization_id": orgID}, 1000, 0)
	if err != nil {
		return nil, err
	}
	for _, p := range props {
		if p.PropertyCode == code {
			return p, nil
		}
	}
	return nil, nil
}

// isNotFound reports whether err represents a "record not found" condition.
// Repositories are inconsistent: some return raw sql.ErrNoRows, others return
// a plain fmt.Errorf("... not found") without wrapping. This covers both.
func isNotFound(err error) bool {
	return err != nil && (errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found"))
}

// readJSON reads and unmarshals a JSON file into T.
func readJSON[T any](path string) (T, error) {
	var zero T
	data, err := os.ReadFile(path)
	if err != nil {
		return zero, err
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return zero, err
	}
	return v, nil
}
