package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ysnarafat/tenantly/internal/models"
)

var (
	ErrPlanNotFound         = errors.New("plan not found")
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrOrganizationSlugUsed = errors.New("organization slug already exists")
	ErrInvitationEmailUsed  = errors.New("invitation email already has an account")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrSupportCaseNotFound  = errors.New("support case not found")
	ErrInvitationNotFound   = errors.New("invitation not found")
	ErrInvitationAccepted   = errors.New("invitation has already been accepted")
	ErrSupportAgentNotFound = errors.New("support agent not found")
)

type OperationsRepository struct {
	db *sqlx.DB
}

func NewOperationsRepository(db *sqlx.DB) *OperationsRepository {
	return &OperationsRepository{db: db}
}

func (r *OperationsRepository) CreateOrganizationOnboarding(
	req *models.CreateOrganizationOnboardingRequest,
	invitedByUserID int,
	token string,
	expiresAt time.Time,
) (*models.OrganizationOnboardingResponse, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin organization onboarding: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var existingUser bool
	if err := tx.Get(&existingUser, "SELECT EXISTS(SELECT 1 FROM users WHERE lower(email) = lower($1))", req.AdminEmail); err != nil {
		return nil, fmt.Errorf("check onboarding email: %w", err)
	}
	if existingUser {
		return nil, ErrInvitationEmailUsed
	}

	var maxUsers int
	if err := tx.Get(&maxUsers, "SELECT max_users FROM plans WHERE code = $1 AND is_active = true", req.PlanCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPlanNotFound
		}
		return nil, fmt.Errorf("get onboarding plan: %w", err)
	}

	org := models.Organization{
		Name:             req.Name,
		Slug:             req.Slug,
		SubscriptionTier: models.SubscriptionTier(req.PlanCode),
		MaxUsers:         maxUsers,
		Active:           true,
	}
	err = tx.QueryRowx(`
		INSERT INTO organizations (name, slug, subscription_tier, max_users, active)
		VALUES ($1, $2, $3, $4, true)
		RETURNING id, created_at, updated_at`,
		req.Name, req.Slug, req.PlanCode, maxUsers,
	).Scan(&org.ID, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		var postgresError *pq.Error
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return nil, ErrOrganizationSlugUsed
		}
		return nil, fmt.Errorf("create onboarding organization: %w", err)
	}

	invitation := models.OrganizationInvitationSummary{
		OrganizationID:      org.ID,
		OrganizationName:    org.Name,
		Email:               req.AdminEmail,
		Role:                "ORG_ADMIN",
		ExpiresAt:           expiresAt,
		EmailDeliveryStatus: "pending",
	}
	err = tx.QueryRowx(`
		INSERT INTO user_invitations (
			organization_id, email, role, invitation_token, expires_at, invited_by_user_id,
			email_delivery_status
		)
		VALUES ($1, $2, 'ORG_ADMIN', $3, $4, $5, 'pending')
		RETURNING id`,
		org.ID, req.AdminEmail, token, expiresAt, invitedByUserID,
	).Scan(&invitation.ID)
	if err != nil {
		return nil, fmt.Errorf("create onboarding invitation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit organization onboarding: %w", err)
	}

	return &models.OrganizationOnboardingResponse{Organization: org, Invitation: invitation}, nil
}

func (r *OperationsRepository) ListPendingOrganizationInvitations() ([]models.OrganizationInvitationSummary, error) {
	const query = `
		SELECT i.id, i.organization_id, o.name AS organization_name, i.email, i.role,
			i.expires_at, i.email_delivery_status, i.last_sent_at
		FROM user_invitations i
		JOIN organizations o ON o.id = i.organization_id
		WHERE i.accepted_at IS NULL
		ORDER BY i.created_at DESC, i.id DESC`
	invitations := make([]models.OrganizationInvitationSummary, 0)
	if err := r.db.Select(&invitations, query); err != nil {
		return nil, fmt.Errorf("list pending organization invitations: %w", err)
	}
	return invitations, nil
}

func (r *OperationsRepository) PrepareOrganizationInvitationResend(
	invitationID int,
	token string,
	expiresAt time.Time,
) (*models.OrganizationInvitationSummary, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin invitation resend: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var invitation models.OrganizationInvitationSummary
	err = tx.Get(&invitation, `
		SELECT i.id, i.organization_id, o.name AS organization_name, i.email, i.role,
			i.expires_at, i.accepted_at, i.email_delivery_status, i.last_sent_at
		FROM user_invitations i
		JOIN organizations o ON o.id = i.organization_id
		WHERE i.id = $1
		FOR UPDATE OF i`, invitationID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvitationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get invitation for resend: %w", err)
	}
	if invitation.AcceptedAt != nil {
		return nil, ErrInvitationAccepted
	}

	if _, err := tx.Exec(`
		UPDATE user_invitations SET invitation_token = $2, expires_at = $3,
			email_delivery_status = 'pending', updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1`, invitationID, token, expiresAt); err != nil {
		return nil, fmt.Errorf("rotate invitation token: %w", err)
	}
	invitation.ExpiresAt = expiresAt
	invitation.EmailDeliveryStatus = "pending"
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit invitation resend: %w", err)
	}
	return &invitation, nil
}

func (r *OperationsRepository) UpdateOrganizationInvitationDeliveryStatus(invitationID int, status string) error {
	result, err := r.db.Exec(`
		UPDATE user_invitations
		SET email_delivery_status = $2,
			last_sent_at = CASE WHEN $2 = 'sent' THEN NOW() AT TIME ZONE 'UTC' ELSE last_sent_at END,
			updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1 AND accepted_at IS NULL`, invitationID, status)
	if err != nil {
		return fmt.Errorf("update organization invitation delivery status: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check organization invitation update: %w", err)
	}
	if count == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

func (r *OperationsRepository) RevokeOrganizationInvitation(invitationID int) error {
	result, err := r.db.Exec(`
		DELETE FROM user_invitations WHERE id = $1 AND accepted_at IS NULL`, invitationID)
	if err != nil {
		return fmt.Errorf("revoke organization invitation: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check invitation revoke: %w", err)
	}
	if count == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

const planColumns = `
	id, code, name_en, name_bn, description_en, description_bn, price_minor,
	RTRIM(currency) AS currency, billing_period, max_users, max_properties,
	features_en, features_bn, is_public, is_active, sort_order, updated_at`

func (r *OperationsRepository) ListPlans(publicOnly bool) ([]models.Plan, error) {
	query := "SELECT " + planColumns + " FROM plans"
	if publicOnly {
		query += " WHERE is_public = true AND is_active = true"
	}
	query += " ORDER BY sort_order, id"

	plans := make([]models.Plan, 0)
	if err := r.db.Select(&plans, query); err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}
	return plans, nil
}

func (r *OperationsRepository) UpdatePlan(code string, req *models.UpdatePlanRequest, userID int) (*models.Plan, error) {
	query := `
		UPDATE plans
		SET name_en = $2, name_bn = $3, description_en = $4, description_bn = $5,
			price_minor = $6, currency = $7, billing_period = $8, max_users = $9,
			max_properties = $10, features_en = $11, features_bn = $12,
			is_public = $13, is_active = $14, sort_order = $15, updated_by = $16,
			updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE code = $1
		RETURNING ` + planColumns

	var plan models.Plan
	err := r.db.Get(&plan, query, code, req.NameEn, req.NameBn, req.DescriptionEn, req.DescriptionBn,
		req.PriceMinor, req.Currency, req.BillingPeriod, req.MaxUsers, req.MaxProperties,
		pq.Array(req.FeaturesEn), pq.Array(req.FeaturesBn), req.IsPublic, req.IsActive, req.SortOrder, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPlanNotFound
		}
		return nil, fmt.Errorf("update plan: %w", err)
	}
	return &plan, nil
}

func (r *OperationsRepository) ListOrganizations() ([]models.OrganizationCustomer, error) {
	const query = `
		SELECT o.id, o.name, o.slug, o.subscription_tier, o.max_users, o.active,
			o.created_at, o.updated_at,
			COALESCE(NULLIF(TRIM(CONCAT_WS(' ', u.first_name, u.last_name)), ''), u.username, '') AS contact_name,
			COALESCE(u.email, '') AS contact_email,
			COALESCE(s.status, 'none') AS subscription_status,
			COALESCE(p.code, o.subscription_tier) AS subscription_plan_code,
			s.renews_at AS subscription_renews_at
		FROM organizations o
		LEFT JOIN LATERAL (
			SELECT u.first_name, u.last_name, u.username, u.email
			FROM user_organization_roles r
			JOIN users u ON u.id = r.user_id
			WHERE r.organization_id = o.id AND r.role IN ('ORG_ADMIN', 'Admin')
			ORDER BY CASE r.role WHEN 'ORG_ADMIN' THEN 0 ELSE 1 END, r.id
			LIMIT 1
		) u ON true
		LEFT JOIN organization_subscriptions s ON s.organization_id = o.id
		LEFT JOIN plans p ON p.id = s.plan_id
		ORDER BY o.created_at DESC, o.id DESC`

	orgs := make([]models.OrganizationCustomer, 0)
	if err := r.db.Select(&orgs, query); err != nil {
		return nil, fmt.Errorf("list customer organizations: %w", err)
	}
	return orgs, nil
}

const subscriptionSelect = `
	SELECT s.id, s.organization_id, s.plan_id, p.code AS plan_code,
		s.plan_name_snapshot AS plan_name, s.status,
		s.price_minor_snapshot, RTRIM(s.currency_snapshot) AS currency,
		s.billing_period_snapshot, s.started_at, s.renews_at, s.updated_at
	FROM organization_subscriptions s
	JOIN plans p ON p.id = s.plan_id`

func (r *OperationsRepository) GetSubscription(organizationID int) (*models.OrganizationSubscription, error) {
	var subscription models.OrganizationSubscription
	err := r.db.Get(&subscription, subscriptionSelect+" WHERE s.organization_id = $1", organizationID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSubscriptionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get organization subscription: %w", err)
	}
	return &subscription, nil
}

func (r *OperationsRepository) UpdateSubscription(organizationID int, req *models.UpdateSubscriptionRequest, userID int) (*models.OrganizationSubscription, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin subscription update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var planID int
	var planName, currency, billingPeriod string
	var priceMinor *int64
	var maxUsers int
	err = tx.QueryRowx(`
		SELECT id, name_en, price_minor, RTRIM(currency), billing_period, max_users
		FROM plans WHERE code = $1 AND is_active = true`, req.PlanCode).
		Scan(&planID, &planName, &priceMinor, &currency, &billingPeriod, &maxUsers)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPlanNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get subscription plan: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE organizations
		SET subscription_tier = $2, max_users = $3, updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1`, organizationID, req.PlanCode, maxUsers)
	if err != nil {
		return nil, fmt.Errorf("update organization tier: %w", err)
	}
	var organizationExists bool
	if err := tx.Get(&organizationExists, "SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1)", organizationID); err != nil {
		return nil, fmt.Errorf("check subscription organization: %w", err)
	}
	if !organizationExists {
		return nil, ErrOrganizationNotFound
	}

	_, err = tx.Exec(`
		INSERT INTO organization_subscriptions (
			organization_id, plan_id, status, plan_name_snapshot, price_minor_snapshot,
			currency_snapshot, billing_period_snapshot, started_at, renews_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE($8, NOW() AT TIME ZONE 'UTC'), $9, $10)
		ON CONFLICT (organization_id) DO UPDATE SET
			plan_id = EXCLUDED.plan_id,
			status = EXCLUDED.status,
			plan_name_snapshot = CASE
				WHEN organization_subscriptions.plan_id = EXCLUDED.plan_id
				THEN organization_subscriptions.plan_name_snapshot ELSE EXCLUDED.plan_name_snapshot END,
			price_minor_snapshot = CASE
				WHEN organization_subscriptions.plan_id = EXCLUDED.plan_id
				THEN organization_subscriptions.price_minor_snapshot ELSE EXCLUDED.price_minor_snapshot END,
			currency_snapshot = CASE
				WHEN organization_subscriptions.plan_id = EXCLUDED.plan_id
				THEN organization_subscriptions.currency_snapshot ELSE EXCLUDED.currency_snapshot END,
			billing_period_snapshot = CASE
				WHEN organization_subscriptions.plan_id = EXCLUDED.plan_id
				THEN organization_subscriptions.billing_period_snapshot ELSE EXCLUDED.billing_period_snapshot END,
			started_at = CASE WHEN $8::timestamp IS NULL
				THEN organization_subscriptions.started_at ELSE EXCLUDED.started_at END,
			renews_at = EXCLUDED.renews_at,
			updated_by = EXCLUDED.updated_by,
			updated_at = NOW() AT TIME ZONE 'UTC'`,
		organizationID, planID, req.Status, planName, priceMinor, currency, billingPeriod,
		req.StartedAt, req.RenewsAt, userID)
	if err != nil {
		return nil, fmt.Errorf("save organization subscription: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit subscription update: %w", err)
	}
	return r.GetSubscription(organizationID)
}

func (r *OperationsRepository) ListSupportCases() ([]models.SupportCase, error) {
	const query = `
		SELECT c.id, c.organization_id, o.name AS organization, c.subject, c.description,
			c.status, c.priority, c.created_by, c.assigned_to, c.created_at, c.updated_at
		FROM support_cases c
		JOIN organizations o ON o.id = c.organization_id
		ORDER BY CASE c.status
			WHEN 'open' THEN 0 WHEN 'in_progress' THEN 1 WHEN 'waiting' THEN 2
			WHEN 'resolved' THEN 3 ELSE 4 END, c.updated_at DESC`
	cases := make([]models.SupportCase, 0)
	if err := r.db.Select(&cases, query); err != nil {
		return nil, fmt.Errorf("list support cases: %w", err)
	}
	return cases, nil
}

func (r *OperationsRepository) CreateSupportCase(req *models.CreateSupportCaseRequest, userID int) (*models.SupportCase, error) {
	const query = `
		INSERT INTO support_cases (organization_id, subject, description, priority, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, organization_id, subject, description, status, priority, created_by,
			assigned_to, created_at, updated_at`
	var supportCase models.SupportCase
	err := r.db.Get(&supportCase, query, req.OrganizationID, strings.TrimSpace(req.Subject),
		strings.TrimSpace(req.Description), req.Priority, userID)
	if err != nil {
		return nil, fmt.Errorf("create support case: %w", err)
	}
	var organization string
	if err := r.db.Get(&organization, "SELECT name FROM organizations WHERE id = $1", req.OrganizationID); err != nil {
		return nil, fmt.Errorf("get support case organization: %w", err)
	}
	supportCase.Organization = organization
	supportCase.Notes = []models.SupportCaseNote{}
	return &supportCase, nil
}

func (r *OperationsRepository) ListTenantSupportCases(organizationID int) ([]models.TenantSupportCase, error) {
	const query = `
		SELECT c.id, c.subject, c.description, c.status, c.priority, c.created_at, c.updated_at
		FROM support_cases c
		WHERE c.organization_id = $1
		ORDER BY c.updated_at DESC, c.id DESC`
	cases := make([]models.TenantSupportCase, 0)
	if err := r.db.Select(&cases, query, organizationID); err != nil {
		return nil, fmt.Errorf("list organization support cases: %w", err)
	}
	return cases, nil
}

func (r *OperationsRepository) GetTenantSupportCase(caseID, organizationID int) (*models.TenantSupportCase, error) {
	const query = `
		SELECT c.id, c.subject, c.description, c.status, c.priority, c.created_at, c.updated_at
		FROM support_cases c
		WHERE c.id = $1 AND c.organization_id = $2`
	var supportCase models.TenantSupportCase
	err := r.db.Get(&supportCase, query, caseID, organizationID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSupportCaseNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get organization support case: %w", err)
	}
	supportCase.Messages, err = r.listSupportCaseMessages(caseID, &organizationID)
	if err != nil {
		return nil, err
	}
	return &supportCase, nil
}

func (r *OperationsRepository) CreateTenantSupportCase(
	organizationID, userID int,
	req *models.CreateTenantSupportCaseRequest,
) (*models.TenantSupportCase, error) {
	const query = `
		INSERT INTO support_cases (organization_id, subject, description, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, subject, description, status, priority, created_at, updated_at`
	var supportCase models.TenantSupportCase
	err := r.db.Get(&supportCase, query, organizationID, strings.TrimSpace(req.Subject),
		strings.TrimSpace(req.Description), userID)
	if err != nil {
		return nil, fmt.Errorf("create organization support case: %w", err)
	}
	return &supportCase, nil
}

func (r *OperationsRepository) ListSupportCaseMessages(caseID int) ([]models.SupportCaseMessage, error) {
	var exists bool
	if err := r.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM support_cases WHERE id = $1)", caseID); err != nil {
		return nil, fmt.Errorf("check support case for messages: %w", err)
	}
	if !exists {
		return nil, ErrSupportCaseNotFound
	}
	return r.listSupportCaseMessages(caseID, nil)
}

func (r *OperationsRepository) AddSupportCaseMessage(
	caseID, authorID int,
	body string,
) (*models.SupportCaseMessage, error) {
	return r.addSupportCaseMessage(caseID, nil, authorID, body)
}

func (r *OperationsRepository) AddTenantSupportCaseMessage(
	caseID, organizationID, authorID int,
	body string,
) (*models.SupportCaseMessage, error) {
	return r.addSupportCaseMessage(caseID, &organizationID, authorID, body)
}

func (r *OperationsRepository) ListSupportAgents() ([]models.SupportAgent, error) {
	const query = `
		SELECT id, COALESCE(NULLIF(TRIM(CONCAT_WS(' ', first_name, last_name)), ''), username) AS name
		FROM users
		WHERE role = 'SUPER_ADMIN' AND active = true
		ORDER BY name, id`
	agents := make([]models.SupportAgent, 0)
	if err := r.db.Select(&agents, query); err != nil {
		return nil, fmt.Errorf("list support agents: %w", err)
	}
	return agents, nil
}

func (r *OperationsRepository) UpdateSupportCaseAssignment(caseID int, assignedTo *int) error {
	result, err := r.db.Exec(`
		UPDATE support_cases
		SET assigned_to = $2, updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
			AND ($2::integer IS NULL OR EXISTS (
				SELECT 1 FROM users WHERE id = $2 AND role = 'SUPER_ADMIN' AND active = true
			))`, caseID, assignedTo)
	if err != nil {
		return fmt.Errorf("assign support case: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check support case assignment: %w", err)
	}
	if count > 0 {
		return nil
	}
	var exists bool
	if err := r.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM support_cases WHERE id = $1)", caseID); err != nil {
		return fmt.Errorf("check support case after assignment: %w", err)
	}
	if !exists {
		return ErrSupportCaseNotFound
	}
	return ErrSupportAgentNotFound
}

func (r *OperationsRepository) listSupportCaseMessages(
	caseID int,
	organizationID *int,
) ([]models.SupportCaseMessage, error) {
	query := `
		SELECT m.id, m.case_id, m.author_id,
			COALESCE(NULLIF(TRIM(CONCAT_WS(' ', u.first_name, u.last_name)), ''), u.username) AS author,
			m.body, m.created_at
		FROM support_case_messages m
		JOIN support_cases c ON c.id = m.case_id
		JOIN users u ON u.id = m.author_id
		WHERE m.case_id = $1`
	args := []any{caseID}
	if organizationID != nil {
		query += " AND c.organization_id = $2"
		args = append(args, *organizationID)
	}
	query += " ORDER BY m.created_at, m.id"

	messages := make([]models.SupportCaseMessage, 0)
	if err := r.db.Select(&messages, query, args...); err != nil {
		return nil, fmt.Errorf("list support case messages: %w", err)
	}
	return messages, nil
}

func (r *OperationsRepository) addSupportCaseMessage(
	caseID int,
	organizationID *int,
	authorID int,
	body string,
) (*models.SupportCaseMessage, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin support message: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		INSERT INTO support_case_messages (case_id, author_id, body)
		SELECT c.id, $2, $3
		FROM support_cases c
		WHERE c.id = $1`
	args := []any{caseID, authorID, strings.TrimSpace(body)}
	if organizationID != nil {
		query += " AND c.organization_id = $4"
		args = append(args, *organizationID)
	}
	query += " RETURNING id, case_id, author_id, body, created_at"

	var message models.SupportCaseMessage
	err = tx.Get(&message, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSupportCaseNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("insert support case message: %w", err)
	}
	if err := tx.Get(&message.Author, `
		SELECT COALESCE(NULLIF(TRIM(CONCAT_WS(' ', first_name, last_name)), ''), username)
		FROM users WHERE id = $1`, authorID); err != nil {
		return nil, fmt.Errorf("get support message author: %w", err)
	}

	updateQuery := `UPDATE support_cases SET updated_at = NOW() AT TIME ZONE 'UTC' WHERE id = $1`
	updateArgs := []any{caseID}
	if organizationID != nil {
		updateQuery += " AND organization_id = $2"
		updateArgs = append(updateArgs, *organizationID)
	}
	if _, err := tx.Exec(updateQuery, updateArgs...); err != nil {
		return nil, fmt.Errorf("update support case timestamp: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit support case message: %w", err)
	}
	return &message, nil
}

func (r *OperationsRepository) UpdateSupportCaseStatus(caseID int, status string) error {
	result, err := r.db.Exec(`
		UPDATE support_cases SET status = $2, updated_at = NOW() AT TIME ZONE 'UTC' WHERE id = $1`,
		caseID, status)
	if err != nil {
		return fmt.Errorf("update support case: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check support case update: %w", err)
	}
	if count == 0 {
		return ErrSupportCaseNotFound
	}
	return nil
}

func (r *OperationsRepository) GetSupportCaseNotes(caseID int) ([]models.SupportCaseNote, error) {
	var exists bool
	if err := r.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM support_cases WHERE id = $1)", caseID); err != nil {
		return nil, fmt.Errorf("check support case: %w", err)
	}
	if !exists {
		return nil, ErrSupportCaseNotFound
	}
	const query = `
		SELECT n.id, n.case_id, n.author_id,
			COALESCE(NULLIF(TRIM(CONCAT_WS(' ', u.first_name, u.last_name)), ''), u.username) AS author,
			n.body, n.created_at
		FROM support_case_notes n
		JOIN users u ON u.id = n.author_id
		WHERE n.case_id = $1
		ORDER BY n.created_at, n.id`
	notes := make([]models.SupportCaseNote, 0)
	if err := r.db.Select(&notes, query, caseID); err != nil {
		return nil, fmt.Errorf("get support case notes: %w", err)
	}
	return notes, nil
}

func (r *OperationsRepository) AddSupportCaseNote(caseID, authorID int, body string) (*models.SupportCaseNote, error) {
	const query = `
		INSERT INTO support_case_notes (case_id, author_id, body)
		SELECT $1, $2, $3 WHERE EXISTS (SELECT 1 FROM support_cases WHERE id = $1)
		RETURNING id, case_id, author_id, body, created_at`
	var note models.SupportCaseNote
	err := r.db.Get(&note, query, caseID, authorID, strings.TrimSpace(body))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSupportCaseNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("add support case note: %w", err)
	}
	if err := r.db.Get(&note.Author, `
		SELECT COALESCE(NULLIF(TRIM(CONCAT_WS(' ', first_name, last_name)), ''), username)
		FROM users WHERE id = $1`, authorID); err != nil {
		return nil, fmt.Errorf("get support note author: %w", err)
	}
	return &note, nil
}

func (r *OperationsRepository) HasOrganization(organizationID int) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1)", organizationID)
	if err != nil {
		return false, fmt.Errorf("check organization: %w", err)
	}
	return exists, nil
}
