-- 000005: Add SaaS platform governance attributes to organizations

ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS plan VARCHAR(30) NOT NULL DEFAULT 'BASIC',
    ADD COLUMN IF NOT EXISTS monthly_fee_in_cents BIGINT NOT NULL DEFAULT 19900 CHECK (monthly_fee_in_cents >= 0),
    ADD COLUMN IF NOT EXISTS subscription_status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_organizations_subscription_status'
    ) THEN
        ALTER TABLE organizations
            ADD CONSTRAINT chk_organizations_subscription_status 
            CHECK (subscription_status IN ('ACTIVE', 'PAST_DUE', 'TRIAL', 'CANCELLED'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_organizations_subscription_status ON organizations (subscription_status);
CREATE INDEX IF NOT EXISTS idx_organizations_is_active ON organizations (is_active);
