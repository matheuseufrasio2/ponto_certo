-- 000008: Add phone and logo_url to organizations for corporate branding

ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS phone VARCHAR(20),
    ADD COLUMN IF NOT EXISTS logo_url VARCHAR(500);

