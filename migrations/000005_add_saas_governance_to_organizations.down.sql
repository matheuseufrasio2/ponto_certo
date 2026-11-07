-- 000005: Revert SaaS platform governance attributes from organizations

DROP INDEX IF EXISTS idx_organizations_is_active;
DROP INDEX IF EXISTS idx_organizations_subscription_status;

ALTER TABLE organizations
    DROP CONSTRAINT IF EXISTS chk_organizations_subscription_status;

ALTER TABLE organizations
    DROP COLUMN IF EXISTS subscription_status,
    DROP COLUMN IF EXISTS monthly_fee_in_cents,
    DROP COLUMN IF EXISTS plan;
