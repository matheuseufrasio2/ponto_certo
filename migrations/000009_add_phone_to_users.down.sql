-- 000009: Drop phone column from users

ALTER TABLE users
    DROP COLUMN IF EXISTS phone;
