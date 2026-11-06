-- 1. Unit Operating Hours (Store Schedule Mon-Sun)
CREATE TABLE IF NOT EXISTS unit_operating_hours (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_id UUID NOT NULL REFERENCES barbershop_units(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Sunday, 1=Monday, ..., 6=Saturday
    open_time TIME NOT NULL,
    close_time TIME NOT NULL,
    is_closed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_unit_operating_day UNIQUE (unit_id, day_of_week),
    CONSTRAINT chk_unit_operating_hours CHECK (open_time < close_time OR is_closed = TRUE)
);

CREATE TRIGGER trg_unit_operating_hours_updated_at
BEFORE UPDATE ON unit_operating_hours
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX IF NOT EXISTS idx_unit_operating_hours_unit ON unit_operating_hours (unit_id);

-- 2. Expand Services with Category, Description, and IsActive flag
ALTER TABLE services
    ADD COLUMN IF NOT EXISTS category VARCHAR(50) NOT NULL DEFAULT 'CABELO',
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS idx_services_unit_active ON services (unit_id, is_active);
CREATE INDEX IF NOT EXISTS idx_services_category ON services (category);

-- 3. Products Catalog & Inventory
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_id UUID NOT NULL REFERENCES barbershop_units(id) ON DELETE CASCADE,
    name VARCHAR(120) NOT NULL,
    category VARCHAR(50) NOT NULL,
    description TEXT,
    price_in_cents BIGINT NOT NULL CHECK (price_in_cents >= 0),
    cost_in_cents BIGINT NOT NULL DEFAULT 0 CHECK (cost_in_cents >= 0),
    stock_quantity INT NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_products_updated_at
BEFORE UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX IF NOT EXISTS idx_products_unit_active ON products (unit_id, is_active);
CREATE INDEX IF NOT EXISTS idx_products_category ON products (category);

-- 4. Barber Working Hours (Weekly shifts & lunch intervals)
CREATE TABLE IF NOT EXISTS barber_working_hours (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    barber_id UUID NOT NULL REFERENCES barbers(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    break_start_time TIME,
    break_end_time TIME,
    is_off BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_barber_working_day UNIQUE (barber_id, day_of_week),
    CONSTRAINT chk_barber_working_hours CHECK (start_time < end_time OR is_off = TRUE)
);

CREATE TRIGGER trg_barber_working_hours_updated_at
BEFORE UPDATE ON barber_working_hours
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX IF NOT EXISTS idx_barber_working_hours_barber ON barber_working_hours (barber_id);

-- 5. Barber Schedule Blocks (Vacations, time off, medical appointments)
CREATE TABLE IF NOT EXISTS barber_schedule_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    barber_id UUID NOT NULL REFERENCES barbers(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    reason VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_barber_schedule_blocks_range CHECK (start_time < end_time)
);

CREATE TRIGGER trg_barber_schedule_blocks_updated_at
BEFORE UPDATE ON barber_schedule_blocks
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX IF NOT EXISTS idx_barber_schedule_blocks_range 
    ON barber_schedule_blocks (barber_id, start_time, end_time);

-- 6. Barber Services Capability (Many-to-Many Skill Mapping)
CREATE TABLE IF NOT EXISTS barber_services (
    barber_id UUID NOT NULL REFERENCES barbers(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (barber_id, service_id)
);

CREATE INDEX IF NOT EXISTS idx_barber_services_service ON barber_services (service_id);

-- 7. Appointments
CREATE TABLE IF NOT EXISTS appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_id UUID NOT NULL REFERENCES barbershop_units(id) ON DELETE CASCADE,
    barber_id UUID NOT NULL REFERENCES barbers(id) ON DELETE RESTRICT,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE RESTRICT,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'SCHEDULED' 
        CHECK (status IN ('SCHEDULED', 'COMPLETED', 'CANCELLED', 'NO_SHOW')),
    payment_status VARCHAR(20) NOT NULL DEFAULT 'PENDING' 
        CHECK (payment_status IN ('PENDING', 'PAID')),
    price_in_cents BIGINT NOT NULL CHECK (price_in_cents >= 0),
    notes VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_appointments_range CHECK (start_time < end_time)
);

CREATE TRIGGER trg_appointments_updated_at
BEFORE UPDATE ON appointments
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX IF NOT EXISTS idx_appointments_barber_status_range 
    ON appointments (barber_id, status, start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_appointments_unit_date 
    ON appointments (unit_id, start_time);
CREATE INDEX IF NOT EXISTS idx_appointments_client 
    ON appointments (client_id);

-- 8. Alter Billings: allow product sales and appointment checkouts
ALTER TABLE billings ALTER COLUMN service_id DROP NOT NULL;

ALTER TABLE billings
    ADD COLUMN IF NOT EXISTS appointment_id UUID REFERENCES appointments(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS product_id UUID REFERENCES products(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS quantity INT NOT NULL DEFAULT 1 CHECK (quantity > 0);

CREATE INDEX IF NOT EXISTS idx_billings_appointment ON billings (appointment_id);
CREATE INDEX IF NOT EXISTS idx_billings_product ON billings (product_id);
