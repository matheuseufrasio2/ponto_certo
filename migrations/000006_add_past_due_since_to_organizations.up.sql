-- 000006: Add past_due_since timestamp for SaaS billing dunning and governance

ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS past_due_since TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_organizations_past_due_since ON organizations (past_due_since);
