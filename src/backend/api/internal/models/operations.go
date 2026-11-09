package models

import (
	"time"

	"github.com/lib/pq"
)

type Plan struct {
	ID            int            `json:"id" db:"id"`
	Code          string         `json:"code" db:"code"`
	NameEn        string         `json:"name_en" db:"name_en"`
	NameBn        string         `json:"name_bn" db:"name_bn"`
	DescriptionEn string         `json:"description_en" db:"description_en"`
	DescriptionBn string         `json:"description_bn" db:"description_bn"`
	PriceMinor    *int64         `json:"price_minor" db:"price_minor"`
	Currency      string         `json:"currency" db:"currency"`
	BillingPeriod string         `json:"billing_period" db:"billing_period"`
	MaxUsers      int            `json:"max_users" db:"max_users"`
	MaxProperties *int           `json:"max_properties" db:"max_properties"`
	FeaturesEn    pq.StringArray `json:"features_en" db:"features_en"`
	FeaturesBn    pq.StringArray `json:"features_bn" db:"features_bn"`
	IsPublic      bool           `json:"is_public" db:"is_public"`
	IsActive      bool           `json:"is_active" db:"is_active"`
	SortOrder     int            `json:"sort_order" db:"sort_order"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}

type UpdatePlanRequest struct {
	NameEn        string   `json:"name_en" binding:"required,min=1,max=100"`
	NameBn        string   `json:"name_bn" binding:"required,min=1,max=100"`
	DescriptionEn string   `json:"description_en" binding:"max=2000"`
	DescriptionBn string   `json:"description_bn" binding:"max=2000"`
	PriceMinor    *int64   `json:"price_minor"`
	Currency      string   `json:"currency" binding:"required,len=3"`
	BillingPeriod string   `json:"billing_period" binding:"required,oneof=monthly yearly"`
	MaxUsers      int      `json:"max_users" binding:"required,min=1"`
	MaxProperties *int     `json:"max_properties"`
	FeaturesEn    []string `json:"features_en"`
	FeaturesBn    []string `json:"features_bn"`
	IsPublic      bool     `json:"is_public"`
	IsActive      bool     `json:"is_active"`
	SortOrder     int      `json:"sort_order"`
}

type OrganizationCustomer struct {
	Organization
	ContactName          string     `json:"contact_name" db:"contact_name"`
	ContactEmail         string     `json:"contact_email" db:"contact_email"`
	SubscriptionStatus   string     `json:"subscription_status" db:"subscription_status"`
	SubscriptionPlanCode string     `json:"subscription_plan_code" db:"subscription_plan_code"`
	SubscriptionRenewsAt *time.Time `json:"subscription_renews_at" db:"subscription_renews_at"`
}

type OrganizationSubscription struct {
	ID             int        `json:"id" db:"id"`
	OrganizationID int        `json:"organization_id" db:"organization_id"`
	PlanID         int        `json:"plan_id" db:"plan_id"`
	PlanCode       string     `json:"plan_code" db:"plan_code"`
	PlanName       string     `json:"plan_name" db:"plan_name"`
	Status         string     `json:"status" db:"status"`
	PriceMinor     *int64     `json:"price_minor" db:"price_minor_snapshot"`
	Currency       string     `json:"currency" db:"currency_snapshot"`
	BillingPeriod  string     `json:"billing_period" db:"billing_period_snapshot"`
	StartedAt      time.Time  `json:"started_at" db:"started_at"`
	RenewsAt       *time.Time `json:"renews_at" db:"renews_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type UpdateSubscriptionRequest struct {
	PlanCode  string     `json:"plan_code" binding:"required,oneof=basic professional enterprise"`
	Status    string     `json:"status" binding:"required,oneof=trialing active past_due canceled"`
	StartedAt *time.Time `json:"started_at"`
	RenewsAt  *time.Time `json:"renews_at"`
}

type SupportCase struct {
	ID             int                  `json:"id" db:"id"`
	OrganizationID int                  `json:"organization_id" db:"organization_id"`
	Organization   string               `json:"organization" db:"organization"`
	Subject        string               `json:"subject" db:"subject"`
	Description    string               `json:"description" db:"description"`
	Status         string               `json:"status" db:"status"`
	Priority       string               `json:"priority" db:"priority"`
	CreatedBy      int                  `json:"created_by" db:"created_by"`
	AssignedTo     *int                 `json:"assigned_to" db:"assigned_to"`
	CreatedAt      time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at" db:"updated_at"`
	Notes          []SupportCaseNote    `json:"notes,omitempty"`
	Messages       []SupportCaseMessage `json:"messages,omitempty"`
}

type SupportCaseNote struct {
	ID        int       `json:"id" db:"id"`
	CaseID    int       `json:"case_id" db:"case_id"`
	AuthorID  int       `json:"author_id" db:"author_id"`
	Author    string    `json:"author" db:"author"`
	Body      string    `json:"body" db:"body"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type SupportCaseMessage struct {
	ID        int       `json:"id" db:"id"`
	CaseID    int       `json:"case_id" db:"case_id"`
	AuthorID  int       `json:"-" db:"author_id"`
	Author    string    `json:"author" db:"author"`
	Body      string    `json:"body" db:"body"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type TenantSupportCase struct {
	ID          int                  `json:"id" db:"id"`
	Subject     string               `json:"subject" db:"subject"`
	Description string               `json:"description" db:"description"`
	Status      string               `json:"status" db:"status"`
	Priority    string               `json:"priority" db:"priority"`
	CreatedAt   time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at" db:"updated_at"`
	Messages    []SupportCaseMessage `json:"messages,omitempty"`
}

type SupportAgent struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type CreateOrganizationOnboardingRequest struct {
	Name       string `json:"name" binding:"required,min=2,max=255"`
	Slug       string `json:"slug" binding:"required,min=2,max=100"`
	PlanCode   string `json:"plan_code" binding:"required,oneof=basic professional enterprise"`
	AdminEmail string `json:"admin_email" binding:"required,email,max=100"`
}

type OrganizationInvitationSummary struct {
	ID                  int        `json:"id" db:"id"`
	OrganizationID      int        `json:"organization_id" db:"organization_id"`
	OrganizationName    string     `json:"organization_name" db:"organization_name"`
	Email               string     `json:"email" db:"email"`
	Role                string     `json:"role" db:"role"`
	ExpiresAt           time.Time  `json:"expires_at" db:"expires_at"`
	AcceptedAt          *time.Time `json:"-" db:"accepted_at"`
	EmailDeliveryStatus string     `json:"email_delivery_status" db:"email_delivery_status"`
	LastSentAt          *time.Time `json:"last_sent_at" db:"last_sent_at"`
}

type OrganizationOnboardingResponse struct {
	Organization Organization                  `json:"organization"`
	Invitation   OrganizationInvitationSummary `json:"invitation"`
}

type CreateSupportCaseRequest struct {
	OrganizationID int    `json:"organization_id" binding:"required,min=1"`
	Subject        string `json:"subject" binding:"required,min=1,max=200"`
	Description    string `json:"description" binding:"required,min=1,max=10000"`
	Priority       string `json:"priority" binding:"required,oneof=low normal high urgent"`
}

type CreateTenantSupportCaseRequest struct {
	Subject     string `json:"subject" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"required,min=1,max=10000"`
}

type UpdateSupportCaseRequest struct {
	Status string `json:"status" binding:"required,oneof=open in_progress waiting resolved closed"`
}

type UpdateSupportCaseAssignmentRequest struct {
	AssignedTo *int `json:"assigned_to"`
}

type CreateSupportCaseNoteRequest struct {
	Body string `json:"body" binding:"required,min=1,max=10000"`
}

type CreateSupportCaseMessageRequest struct {
	Body string `json:"body" binding:"required,min=1,max=10000"`
}
