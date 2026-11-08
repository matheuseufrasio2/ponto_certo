-- 000007 Down: Drop Employee Leaves, Revert Barber Schedule Blocks, and Drop Audit Logs

DROP TRIGGER IF EXISTS trg_employee_leaves_updated_at ON employee_leaves;
DROP TABLE IF EXISTS employee_leaves;

ALTER TABLE barber_schedule_blocks
    DROP COLUMN IF EXISTS created_by_user_id;

DROP TABLE IF EXISTS audit_logs;
