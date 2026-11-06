-- 8. Revert Billings alterations
DROP INDEX IF EXISTS idx_billings_product;
DROP INDEX IF EXISTS idx_billings_appointment;
ALTER TABLE billings DROP COLUMN IF EXISTS quantity;
ALTER TABLE billings DROP COLUMN IF EXISTS product_id;
ALTER TABLE billings DROP COLUMN IF EXISTS appointment_id;

-- 7. Drop Appointments
DROP TRIGGER IF EXISTS trg_appointments_updated_at ON appointments;
DROP TABLE IF EXISTS appointments;

-- 6. Drop Barber Services
DROP TABLE IF EXISTS barber_services;

-- 5. Drop Barber Schedule Blocks
DROP TRIGGER IF EXISTS trg_barber_schedule_blocks_updated_at ON barber_schedule_blocks;
DROP TABLE IF EXISTS barber_schedule_blocks;

-- 4. Drop Barber Working Hours
DROP TRIGGER IF EXISTS trg_barber_working_hours_updated_at ON barber_working_hours;
DROP TABLE IF EXISTS barber_working_hours;

-- 3. Drop Products
DROP TRIGGER IF EXISTS trg_products_updated_at ON products;
DROP TABLE IF EXISTS products;

-- 2. Revert Services additions
DROP INDEX IF EXISTS idx_services_category;
DROP INDEX IF EXISTS idx_services_unit_active;
ALTER TABLE services DROP COLUMN IF EXISTS is_active;
ALTER TABLE services DROP COLUMN IF EXISTS description;
ALTER TABLE services DROP COLUMN IF EXISTS category;

-- 1. Drop Unit Operating Hours
DROP TRIGGER IF EXISTS trg_unit_operating_hours_updated_at ON unit_operating_hours;
DROP TABLE IF EXISTS unit_operating_hours;
