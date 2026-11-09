DROP TABLE IF EXISTS support_case_notes;
DROP TABLE IF EXISTS support_cases;
DROP TRIGGER IF EXISTS organizations_subscription_tier_update ON organizations;
DROP TRIGGER IF EXISTS organizations_subscription_insert ON organizations;
DROP FUNCTION IF EXISTS sync_organization_subscription();
DROP TABLE IF EXISTS organization_subscriptions;
DROP TABLE IF EXISTS plans;
