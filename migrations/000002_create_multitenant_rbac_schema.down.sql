-- Revert Billings payment method constraint
ALTER TABLE billings DROP CONSTRAINT IF EXISTS billings_payment_method_check;
ALTER TABLE billings ADD CONSTRAINT billings_payment_method_check 
    CHECK (payment_method IN ('CARD', 'CASH', 'PIX', 'OTHER'));

-- Remove unit_id from Billings
DROP INDEX IF EXISTS idx_billings_unit_date;
ALTER TABLE billings DROP COLUMN IF EXISTS unit_id;

-- Remove unit_id from Services
DROP INDEX IF EXISTS idx_services_unit;
ALTER TABLE services DROP COLUMN IF EXISTS unit_id;

-- Remove unit_id from Clients
DROP INDEX IF EXISTS idx_clients_unit;
ALTER TABLE clients DROP COLUMN IF EXISTS unit_id;

-- Remove unit_id, user_id and commission_percentage from Barbers
DROP INDEX IF EXISTS idx_barbers_unit;
ALTER TABLE barbers DROP COLUMN IF EXISTS commission_percentage;
ALTER TABLE barbers DROP COLUMN IF EXISTS user_id;
ALTER TABLE barbers DROP COLUMN IF EXISTS unit_id;

-- Drop Users table
DROP TABLE IF EXISTS users;

-- Drop Barbershop Units table
DROP TABLE IF EXISTS barbershop_units;

-- Drop Organizations table
DROP TABLE IF EXISTS organizations;
