-- 1. Alter Barbershop Units: Add chair count for capacity & occupancy tracking
ALTER TABLE barbershop_units
    ADD COLUMN IF NOT EXISTS chair_count INT NOT NULL DEFAULT 4 CHECK (chair_count > 0);

-- 2. Alter Barbers: Add operational status (AVAILABLE, BUSY, LUNCH, OFF_DUTY, VACATION)
ALTER TABLE barbers
    ADD COLUMN IF NOT EXISTS status VARCHAR(30) NOT NULL DEFAULT 'AVAILABLE'
    CHECK (status IN ('AVAILABLE', 'BUSY', 'LUNCH', 'OFF_DUTY', 'VACATION'));

CREATE INDEX IF NOT EXISTS idx_barbers_unit_status ON barbers (unit_id, status);

-- 3. Barber Status Shift Audit Logs
CREATE TABLE IF NOT EXISTS barber_status_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    barber_id UUID NOT NULL REFERENCES barbers(id) ON DELETE CASCADE,
    previous_status VARCHAR(30),
    new_status VARCHAR(30) NOT NULL,
    changed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_barber_status_logs_updated_at
BEFORE UPDATE ON barber_status_logs
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX IF NOT EXISTS idx_barber_status_logs_barber ON barber_status_logs (barber_id);
