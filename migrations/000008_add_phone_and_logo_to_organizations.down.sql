-- 000008: Drop phone and logo_url from organizations

ALTER TABLE organizations
    DROP COLUMN IF EXISTS logo_url,
    DROP COLUMN IF EXISTS phone;

