-- 1. Organizations (Multi-tenant Root / Franchises / Brands)
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(120) NOT NULL,
    trade_name VARCHAR(120),
    document VARCHAR(30), -- CNPJ or CPF
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_organizations_updated_at
BEFORE UPDATE ON organizations
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 2. Barbershop Units (Branches / Physical Locations)
CREATE TABLE IF NOT EXISTS barbershop_units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(120) NOT NULL,
    phone VARCHAR(20),
    address VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_barbershop_units_updated_at
BEFORE UPDATE ON barbershop_units
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX IF NOT EXISTS idx_barbershop_units_org ON barbershop_units (organization_id);

-- 3. Users (Authentication and System Actors with RBAC)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    unit_id UUID REFERENCES barbershop_units(id) ON DELETE SET NULL,
    name VARCHAR(120) NOT NULL,
    email VARCHAR(120) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(30) NOT NULL CHECK (role IN ('SUPER_ADMIN', 'ORG_OWNER', 'UNIT_MANAGER', 'RECEPTIONIST', 'BARBER')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX IF NOT EXISTS idx_users_org ON users (organization_id);
CREATE INDEX IF NOT EXISTS idx_users_unit ON users (unit_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- 4. Alter Barbers: link to Unit, User and add Commission Percentage
ALTER TABLE barbers 
    ADD COLUMN IF NOT EXISTS unit_id UUID REFERENCES barbershop_units(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS user_id UUID UNIQUE REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS commission_percentage INT NOT NULL DEFAULT 50 CHECK (commission_percentage >= 0 AND commission_percentage <= 100);

CREATE INDEX IF NOT EXISTS idx_barbers_unit ON barbers (unit_id);

-- 5. Alter Clients: link to Unit
ALTER TABLE clients 
    ADD COLUMN IF NOT EXISTS unit_id UUID REFERENCES barbershop_units(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_clients_unit ON clients (unit_id);

-- 6. Alter Services: link to Unit
ALTER TABLE services 
    ADD COLUMN IF NOT EXISTS unit_id UUID REFERENCES barbershop_units(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_services_unit ON services (unit_id);

-- 7. Alter Billings: link to Unit and update Payment Method constraint
ALTER TABLE billings 
    ADD COLUMN IF NOT EXISTS unit_id UUID REFERENCES barbershop_units(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_billings_unit_date ON billings (unit_id, date);

-- Expand payment method constraints to include CREDIT_CARD and DEBIT_CARD
ALTER TABLE billings DROP CONSTRAINT IF EXISTS billings_payment_method_check;
ALTER TABLE billings ADD CONSTRAINT billings_payment_method_check 
    CHECK (payment_method IN ('PIX', 'CREDIT_CARD', 'DEBIT_CARD', 'CASH', 'OTHER', 'CARD'));
