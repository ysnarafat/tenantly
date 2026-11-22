-- Operations portal: plans, subscriptions, support cases

-- Billing plans
CREATE TABLE plans (
    id SERIAL PRIMARY KEY,
    code VARCHAR(30) NOT NULL UNIQUE CHECK (code IN ('basic', 'professional', 'enterprise')),
    name_en VARCHAR(100) NOT NULL,
    name_bn VARCHAR(100) NOT NULL DEFAULT '',
    description_en TEXT NOT NULL DEFAULT '',
    description_bn TEXT NOT NULL DEFAULT '',
    price_minor BIGINT CHECK (price_minor IS NULL OR price_minor >= 0),
    currency CHAR(3) NOT NULL DEFAULT 'BDT',
    billing_period VARCHAR(20) NOT NULL DEFAULT 'monthly'
        CHECK (billing_period IN ('monthly', 'yearly')),
    max_users INTEGER NOT NULL CHECK (max_users > 0),
    max_properties INTEGER CHECK (max_properties IS NULL OR max_properties > 0),
    features_en TEXT[] NOT NULL DEFAULT '{}',
    features_bn TEXT[] NOT NULL DEFAULT '{}',
    is_public BOOLEAN NOT NULL DEFAULT true,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INTEGER NOT NULL DEFAULT 0,
    updated_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);

INSERT INTO plans (
    code, name_en, name_bn, description_en, description_bn, price_minor,
    max_users, max_properties, features_en, features_bn, sort_order
)
VALUES
    ('basic', 'Basic', 'বেসিক', 'For independent property owners getting started.', 'স্বতন্ত্র সম্পত্তির মালিকদের জন্য শুরু করার সহজ পরিকল্পনা।', NULL, 50, NULL,
     ARRAY['Property and tenant management', 'Lease tracking', 'Rent payment records'],
     ARRAY['সম্পত্তি ও ভাড়াটিয়া ব্যবস্থাপনা', 'লিজ ট্র্যাকিং', 'ভাড়া পরিশোধের রেকর্ড'], 1),
    ('professional', 'Professional', 'প্রফেশনাল', 'For growing property management teams.', 'বর্ধনশীল সম্পত্তি ব্যবস্থাপনা দলের জন্য।', NULL, 50, NULL,
     ARRAY['Everything in Basic', 'Team access and role management', 'Reports and analysis'],
     ARRAY['বেসিকের সব সুবিধা', 'টিম অ্যাক্সেস ও ভূমিকা ব্যবস্থাপনা', 'রিপোর্ট ও বিশ্লেষণ'], 2),
    ('enterprise', 'Enterprise', 'এন্টারপ্রাইজ', 'For larger portfolios and multi-organization teams.', 'বৃহৎ পোর্টফোলিও এবং একাধিক প্রতিষ্ঠানের দলের জন্য।', NULL, 50, NULL,
     ARRAY['Everything in Professional', 'Portfolio-level oversight', 'Priority support'],
     ARRAY['প্রফেশনালের সব সুবিধা', 'পোর্টফোলিও পর্যবেক্ষণ', 'অগ্রাধিকারভিত্তিক সহায়তা'], 3);

-- Organization subscriptions
CREATE TABLE organization_subscriptions (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    plan_id INTEGER NOT NULL REFERENCES plans(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('trialing', 'active', 'past_due', 'canceled')),
    plan_name_snapshot VARCHAR(100) NOT NULL,
    price_minor_snapshot BIGINT CHECK (price_minor_snapshot IS NULL OR price_minor_snapshot >= 0),
    currency_snapshot CHAR(3) NOT NULL DEFAULT 'BDT',
    billing_period_snapshot VARCHAR(20) NOT NULL
        CHECK (billing_period_snapshot IN ('monthly', 'yearly')),
    started_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    renews_at TIMESTAMP,
    updated_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);

INSERT INTO organization_subscriptions (
    organization_id, plan_id, status, plan_name_snapshot, price_minor_snapshot,
    currency_snapshot, billing_period_snapshot
)
SELECT o.id, p.id, CASE WHEN o.active THEN 'active' ELSE 'canceled' END,
       p.name_en, p.price_minor, p.currency, p.billing_period
FROM organizations o
JOIN plans p ON p.code = o.subscription_tier;

-- Sync organization subscription trigger function
CREATE FUNCTION sync_organization_subscription() RETURNS trigger AS $$
DECLARE
    selected_plan plans%ROWTYPE;
BEGIN
    SELECT * INTO selected_plan FROM plans WHERE code = NEW.subscription_tier;
    INSERT INTO organization_subscriptions (
        organization_id, plan_id, status, plan_name_snapshot, price_minor_snapshot,
        currency_snapshot, billing_period_snapshot
    ) VALUES (
        NEW.id, selected_plan.id, CASE WHEN NEW.active THEN 'active' ELSE 'canceled' END,
        selected_plan.name_en, selected_plan.price_minor, selected_plan.currency,
        selected_plan.billing_period
    )
    ON CONFLICT (organization_id) DO UPDATE SET
        plan_id = EXCLUDED.plan_id,
        plan_name_snapshot = EXCLUDED.plan_name_snapshot,
        price_minor_snapshot = EXCLUDED.price_minor_snapshot,
        currency_snapshot = EXCLUDED.currency_snapshot,
        billing_period_snapshot = EXCLUDED.billing_period_snapshot,
        updated_at = NOW() AT TIME ZONE 'UTC';
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER organizations_subscription_insert
    AFTER INSERT ON organizations
    FOR EACH ROW EXECUTE FUNCTION sync_organization_subscription();

CREATE TRIGGER organizations_subscription_tier_update
    AFTER UPDATE OF subscription_tier ON organizations
    FOR EACH ROW
    WHEN (OLD.subscription_tier IS DISTINCT FROM NEW.subscription_tier)
    EXECUTE FUNCTION sync_organization_subscription();

-- Support cases
CREATE TABLE support_cases (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    subject VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'in_progress', 'waiting', 'resolved', 'closed')),
    priority VARCHAR(20) NOT NULL DEFAULT 'normal'
        CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    created_by INTEGER NOT NULL REFERENCES users(id),
    assigned_to INTEGER REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);

CREATE TABLE support_case_notes (
    id SERIAL PRIMARY KEY,
    case_id INTEGER NOT NULL REFERENCES support_cases(id) ON DELETE CASCADE,
    author_id INTEGER NOT NULL REFERENCES users(id),
    body TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);

-- Support case messages
CREATE TABLE support_case_messages (
    id SERIAL PRIMARY KEY,
    case_id INTEGER NOT NULL REFERENCES support_cases(id) ON DELETE CASCADE,
    author_id INTEGER NOT NULL REFERENCES users(id),
    body TEXT NOT NULL CHECK (length(btrim(body)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);

CREATE INDEX idx_plans_public ON plans(is_public, is_active, sort_order);
CREATE INDEX idx_subscriptions_status ON organization_subscriptions(status);
CREATE INDEX idx_support_cases_status ON support_cases(status, updated_at DESC);
CREATE INDEX idx_support_cases_organization ON support_cases(organization_id, created_at DESC);
CREATE INDEX idx_support_case_notes_case ON support_case_notes(case_id, created_at);
CREATE INDEX idx_support_case_messages_case ON support_case_messages(case_id, created_at, id);