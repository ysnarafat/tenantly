DROP INDEX IF EXISTS idx_support_case_messages_case;
DROP TABLE IF EXISTS support_case_messages;

DROP INDEX IF EXISTS idx_support_case_notes_case;
DROP TABLE IF EXISTS support_case_notes;

DROP INDEX IF EXISTS idx_support_cases_organization;
DROP INDEX IF EXISTS idx_support_cases_status;
DROP TABLE IF EXISTS support_cases;

DROP TRIGGER IF EXISTS organizations_subscription_tier_update ON organizations;
DROP TRIGGER IF EXISTS organizations_subscription_insert ON organizations;
DROP FUNCTION IF EXISTS sync_organization_subscription();

DROP INDEX IF EXISTS idx_subscriptions_status;
DROP TABLE IF EXISTS organization_subscriptions;

DROP INDEX IF EXISTS idx_plans_public;
DROP TABLE IF EXISTS plans;