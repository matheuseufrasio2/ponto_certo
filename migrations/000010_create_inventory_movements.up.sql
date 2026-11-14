-- 000010: Create Product Inventory Movements Table (Stock Control & Audit)
CREATE TABLE IF NOT EXISTS product_inventory_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    unit_id UUID NOT NULL REFERENCES barbershop_units(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    movement_type VARCHAR(20) NOT NULL CHECK (movement_type IN ('ENTRY', 'EXIT', 'ADJUSTMENT')),
    quantity INT NOT NULL,
    previous_stock INT NOT NULL,
    new_stock INT NOT NULL,
    reason VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_inventory_movements_product_created ON product_inventory_movements (product_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_inventory_movements_unit_created ON product_inventory_movements (unit_id, created_at DESC);
