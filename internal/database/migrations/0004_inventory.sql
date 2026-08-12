-- =============================================================================
-- 0004_inventory.sql
-- Idempotent inventory supplemental migration
-- Version: PostgreSQL 14 compatible
-- =============================================================================

-- ---------------------------------------------------------------------------
-- Ensure movement_type ENUM values are complete (defined in 0001)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'movement_type' AND e.enumlabel = 'adjustment'
    ) THEN
        ALTER TYPE movement_type ADD VALUE IF NOT EXISTS 'adjustment';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'movement_type' AND e.enumlabel = 'transfer'
    ) THEN
        ALTER TYPE movement_type ADD VALUE IF NOT EXISTS 'transfer';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'movement_type' AND e.enumlabel = 'purchase'
    ) THEN
        ALTER TYPE movement_type ADD VALUE IF NOT EXISTS 'purchase';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'movement_type' AND e.enumlabel = 'sale'
    ) THEN
        ALTER TYPE movement_type ADD VALUE IF NOT EXISTS 'sale';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'movement_type' AND e.enumlabel = 'return_in'
    ) THEN
        ALTER TYPE movement_type ADD VALUE IF NOT EXISTS 'return_in';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'movement_type' AND e.enumlabel = 'return_out'
    ) THEN
        ALTER TYPE movement_type ADD VALUE IF NOT EXISTS 'return_out';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'movement_type' AND e.enumlabel = 'production_in'
    ) THEN
        ALTER TYPE movement_type ADD VALUE IF NOT EXISTS 'production_in';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'movement_type' AND e.enumlabel = 'production_out'
    ) THEN
        ALTER TYPE movement_type ADD VALUE IF NOT EXISTS 'production_out';
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Additional indexes on items (only if table exists)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'items') THEN
        CREATE INDEX IF NOT EXISTS idx_items_category ON items(category_id);
        CREATE INDEX IF NOT EXISTS idx_items_uom ON items(uom_id);
        CREATE INDEX IF NOT EXISTS idx_items_type ON items(item_type);
        CREATE INDEX IF NOT EXISTS idx_items_barcode ON items(barcode) WHERE barcode IS NOT NULL;
        CREATE INDEX IF NOT EXISTS idx_items_active ON items(company_id) WHERE is_active = TRUE;
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Additional indexes on warehouses (only if table exists)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'warehouses') THEN
        CREATE INDEX IF NOT EXISTS idx_warehouses_company ON warehouses(company_id, is_active);
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Additional indexes on warehouse_locations (only if table exists)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'warehouse_locations') THEN
        CREATE INDEX IF NOT EXISTS idx_wloc_warehouse ON warehouse_locations(warehouse_id);
        CREATE INDEX IF NOT EXISTS idx_wloc_parent ON warehouse_locations(parent_id) WHERE parent_id IS NOT NULL;
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Additional indexes on stock_levels (only if table exists)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'stock_levels') THEN
        CREATE INDEX IF NOT EXISTS idx_sl_low_stock ON stock_levels(company_id, item_id)
            WHERE qty_available < 0;
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Additional indexes on stock_movements (only if table exists)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'stock_movements') THEN
        CREATE INDEX IF NOT EXISTS idx_sm_warehouse ON stock_movements(warehouse_id);
        CREATE INDEX IF NOT EXISTS idx_sm_type_date ON stock_movements(type, date DESC);
        CREATE INDEX IF NOT EXISTS idx_sm_reference ON stock_movements(reference) WHERE reference IS NOT NULL;
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Additional indexes on inventory_counts (only if table exists)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'inventory_counts') THEN
        CREATE INDEX IF NOT EXISTS idx_ic_company_date ON inventory_counts(company_id, date DESC);
        CREATE INDEX IF NOT EXISTS idx_ic_warehouse ON inventory_counts(warehouse_id);
        CREATE INDEX IF NOT EXISTS idx_ic_status ON inventory_counts(company_id, status);
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'inventory_count_lines') THEN
        CREATE INDEX IF NOT EXISTS idx_icl_count ON inventory_count_lines(count_id);
        CREATE INDEX IF NOT EXISTS idx_icl_item ON inventory_count_lines(item_id);
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- View: v_inventory_valuation (only if required tables exist)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables WHERE table_name = 'stock_levels'
    ) AND EXISTS (
        SELECT 1 FROM information_schema.tables WHERE table_name = 'items'
    ) AND EXISTS (
        SELECT 1 FROM information_schema.tables WHERE table_name = 'warehouses'
    ) THEN
        EXECUTE '
        CREATE OR REPLACE VIEW v_inventory_valuation AS
        SELECT
            sl.company_id,
            sl.item_id,
            i.code                          AS item_code,
            i.name                          AS item_name,
            i.item_type,
            c.name                          AS category_name,
            u.code                          AS uom_code,
            sl.warehouse_id,
            w.name                          AS warehouse_name,
            sl.location_id,
            sl.qty_on_hand,
            sl.qty_reserved,
            sl.qty_available,
            sl.cmup_cost,
            ROUND(sl.qty_on_hand * sl.cmup_cost, 2)     AS total_value,
            ROUND(sl.qty_available * sl.cmup_cost, 2)   AS available_value,
            i.min_stock_qty,
            i.reorder_qty,
            i.max_stock_qty,
            CASE
                WHEN sl.qty_available <= 0           THEN ''out_of_stock''
                WHEN sl.qty_available < i.min_stock_qty THEN ''low_stock''
                WHEN i.max_stock_qty > 0 AND sl.qty_available > i.max_stock_qty THEN ''over_stock''
                ELSE ''normal''
            END                                         AS stock_status,
            sl.updated_at
        FROM stock_levels sl
        JOIN items       i  ON i.id = sl.item_id
        JOIN warehouses  w  ON w.id = sl.warehouse_id
        LEFT JOIN item_categories c ON c.id = i.category_id
        LEFT JOIN units_of_measure u ON u.id = i.uom_id;
        ';
    ELSE
        RAISE NOTICE 'Required tables for v_inventory_valuation do not exist. Skipping view creation.';
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- View: v_stock_movements_detail (FIXED - users table columns)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables WHERE table_name = 'stock_movements'
    ) AND EXISTS (
        SELECT 1 FROM information_schema.tables WHERE table_name = 'items'
    ) AND EXISTS (
        SELECT 1 FROM information_schema.tables WHERE table_name = 'warehouses'
    ) THEN
        EXECUTE '
        CREATE OR REPLACE VIEW v_stock_movements_detail AS
        SELECT
            sm.id,
            sm.company_id,
            sm.number,
            sm.date,
            sm.type,
            sm.item_id,
            i.code                          AS item_code,
            i.name                          AS item_name,
            sm.warehouse_id,
            w.name                          AS warehouse_name,
            sm.to_warehouse_id,
            w2.name                         AS to_warehouse_name,
            sm.quantity,
            sm.unit_cost,
            sm.total_cost,
            sm.reference,
            sm.source_type,
            sm.source_id,
            sm.notes,
            sm.created_by,
            u.username                      AS created_by_name,
            sm.created_at
        FROM stock_movements sm
        JOIN items       i  ON i.id = sm.item_id
        JOIN warehouses  w  ON w.id = sm.warehouse_id
        LEFT JOIN warehouses  w2 ON w2.id = sm.to_warehouse_id
        LEFT JOIN users       u  ON u.id = sm.created_by;
        ';
    ELSE
        RAISE NOTICE 'Required tables for v_stock_movements_detail do not exist. Skipping view creation.';
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- View: v_inventory_count_detail (only if required tables exist)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables WHERE table_name = 'inventory_count_lines'
    ) AND EXISTS (
        SELECT 1 FROM information_schema.tables WHERE table_name = 'inventory_counts'
    ) AND EXISTS (
        SELECT 1 FROM information_schema.tables WHERE table_name = 'items'
    ) AND EXISTS (
        SELECT 1 FROM information_schema.tables WHERE table_name = 'warehouses'
    ) THEN
        EXECUTE '
        CREATE OR REPLACE VIEW v_inventory_count_detail AS
        SELECT
            icl.id,
            icl.count_id,
            ic.number                       AS count_number,
            ic.date,
            ic.status,
            ic.warehouse_id,
            wh.name                         AS warehouse_name,
            icl.item_id,
            i.code                          AS item_code,
            i.name                          AS item_name,
            u.code                          AS uom_code,
            icl.location_id,
            icl.book_qty,
            icl.counted_qty,
            icl.difference,
            icl.unit_cost,
            ROUND(icl.difference * icl.unit_cost, 2)   AS variance_value
        FROM inventory_count_lines  icl
        JOIN inventory_counts ic    ON ic.id = icl.count_id
        JOIN items            i     ON i.id  = icl.item_id
        JOIN warehouses       wh    ON wh.id = ic.warehouse_id
        LEFT JOIN units_of_measure u ON u.id = i.uom_id;
        ';
    ELSE
        RAISE NOTICE 'Required tables for v_inventory_count_detail do not exist. Skipping view creation.';
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Function: next_movement_number(company_id UUID) → VARCHAR
-- Generates sequential movement numbers: MOV-YYYY-000001
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION next_movement_number(p_company_id UUID)
RETURNS VARCHAR AS $$
DECLARE
    v_seq  BIGINT;
    v_year VARCHAR(4);
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'stock_movements') THEN
        v_year := TO_CHAR(NOW(), 'YYYY');
        SELECT COALESCE(MAX(CAST(SUBSTRING(number FROM 'MOV-\d{4}-(\d+)') AS BIGINT)), 0) + 1
        INTO v_seq
        FROM stock_movements
        WHERE company_id = p_company_id
          AND number LIKE 'MOV-' || v_year || '-%';

        RETURN 'MOV-' || v_year || '-' || LPAD(v_seq::TEXT, 6, '0');
    ELSE
        RETURN 'MOV-' || TO_CHAR(NOW(), 'YYYY') || '-000001';
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ---------------------------------------------------------------------------
-- Function: next_count_number(company_id UUID) → VARCHAR
-- Generates sequential inventory count numbers: IC-YYYY-000001
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION next_count_number(p_company_id UUID)
RETURNS VARCHAR AS $$
DECLARE
    v_seq  BIGINT;
    v_year VARCHAR(4);
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'inventory_counts') THEN
        v_year := TO_CHAR(NOW(), 'YYYY');
        SELECT COALESCE(MAX(CAST(SUBSTRING(number FROM 'IC-\d{4}-(\d+)') AS BIGINT)), 0) + 1
        INTO v_seq
        FROM inventory_counts
        WHERE company_id = p_company_id
          AND number LIKE 'IC-' || v_year || '-%';

        RETURN 'IC-' || v_year || '-' || LPAD(v_seq::TEXT, 6, '0');
    ELSE
        RETURN 'IC-' || TO_CHAR(NOW(), 'YYYY') || '-000001';
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ---------------------------------------------------------------------------
-- Done
-- ---------------------------------------------------------------------------