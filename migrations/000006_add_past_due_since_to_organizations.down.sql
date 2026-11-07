-- 000006 Down: Remove past_due_since from organizations

DROP INDEX IF EXISTS idx_organizations_past_due_since;

ALTER TABLE organizations
    DROP COLUMN IF EXISTS past_due_since;
