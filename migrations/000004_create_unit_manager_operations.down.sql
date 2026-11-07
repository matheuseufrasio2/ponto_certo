DROP TRIGGER IF EXISTS trg_barber_status_logs_updated_at ON barber_status_logs;
DROP TABLE IF EXISTS barber_status_logs;

DROP INDEX IF EXISTS idx_barbers_unit_status;
ALTER TABLE barbers DROP COLUMN IF EXISTS status;

ALTER TABLE barbershop_units DROP COLUMN IF EXISTS chair_count;
