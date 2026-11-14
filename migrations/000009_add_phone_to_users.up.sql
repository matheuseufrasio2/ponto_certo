-- 000009: Add phone column to users for unit manager contact management

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS phone VARCHAR(20);
