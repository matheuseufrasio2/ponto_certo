# Migrations (migrations/)

This directory stores versioned SQL schema evolution scripts (DDL).

## Recommended Standard (with `golang-migrate` or `goose`):
Each migration uses a paired timestamped or sequential file:
- `000001_create_barbers_table.up.sql` (applies the schema change)
- `000001_create_barbers_table.down.sql` (reverts the schema change)

### Core Barbershop Tables:
- `barbers`: id (uuid), name, phone, is_active, created_at
- `customers`: id (uuid), name, phone, email, created_at
- `services`: id (uuid), name, duration_minutes, price_in_cents, created_at
- `appointments`: id (uuid), barber_id, customer_id, service_id, start_time, end_time, status, notes, created_at, updated_at
