-- Trigger helper for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 1. Barbers
CREATE TABLE IF NOT EXISTS barbers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(80) NOT NULL,
    phone VARCHAR(20),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_barbers_updated_at
BEFORE UPDATE ON barbers
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 2. Clients
CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(120) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(120),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_clients_updated_at
BEFORE UPDATE ON clients
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 3. Services
CREATE TABLE IF NOT EXISTS services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(120) NOT NULL,
    duration_minutes INT NOT NULL DEFAULT 30,
    price_in_cents BIGINT NOT NULL CHECK (price_in_cents >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_services_updated_at
BEFORE UPDATE ON services
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 4. Billings
CREATE TABLE IF NOT EXISTS billings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL DEFAULT CURRENT_DATE,
    barber_id UUID NOT NULL REFERENCES barbers(id) ON DELETE RESTRICT,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE RESTRICT,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
    amount_in_cents BIGINT NOT NULL CHECK (amount_in_cents >= 0),
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('CARD', 'CASH', 'PIX', 'OTHER')),
    status VARCHAR(20) NOT NULL DEFAULT 'PAID' CHECK (status IN ('PAID', 'CANCELLED')),
    notes VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_billings_updated_at
BEFORE UPDATE ON billings
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- High-performance indexes for financial and operational queries
CREATE INDEX IF NOT EXISTS idx_billings_date_status ON billings (date, status);
CREATE INDEX IF NOT EXISTS idx_billings_barber_date ON billings (barber_id, date);
CREATE INDEX IF NOT EXISTS idx_billings_client ON billings (client_id);
