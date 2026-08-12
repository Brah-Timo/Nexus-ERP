-- =============================================================================
-- 0005_manufacturing.sql  — Idempotent manufacturing migration
-- =============================================================================

-- ── Indexes on work_centers ──────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_wc_company ON work_centers(company_id);
CREATE INDEX IF NOT EXISTS idx_wc_code ON work_centers(company_id, code);
CREATE INDEX IF NOT EXISTS idx_wc_active ON work_centers(company_id, is_active);

-- ── Indexes on bill_of_materials ─────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_bom_company ON bill_of_materials(company_id);
CREATE INDEX IF NOT EXISTS idx_bom_product ON bill_of_materials(product_id);
CREATE INDEX IF NOT EXISTS idx_bom_active ON bill_of_materials(company_id, is_active);
CREATE INDEX IF NOT EXISTS idx_bom_code ON bill_of_materials(company_id, code);

-- ── Indexes on bom_components ────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_bom_comp_bom ON bom_components(bom_id);
CREATE INDEX IF NOT EXISTS idx_bom_comp_item ON bom_components(component_id);

-- ── Indexes on bom_operations ────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_bom_op_bom ON bom_operations(bom_id);
CREATE INDEX IF NOT EXISTS idx_bom_op_wc ON bom_operations(work_center_id);

-- ── Indexes on manufacturing_orders ──────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_mo_company ON manufacturing_orders(company_id);
CREATE INDEX IF NOT EXISTS idx_mo_status ON manufacturing_orders(company_id, status);
CREATE INDEX IF NOT EXISTS idx_mo_product ON manufacturing_orders(product_id);
CREATE INDEX IF NOT EXISTS idx_mo_bom ON manufacturing_orders(bom_id);
CREATE INDEX IF NOT EXISTS idx_mo_planned_start ON manufacturing_orders(planned_start);
CREATE INDEX IF NOT EXISTS idx_mo_number ON manufacturing_orders(company_id, number);

-- ── Indexes on mo_component_lines ────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_mo_comp_mo ON mo_component_lines(mo_id);
CREATE INDEX IF NOT EXISTS idx_mo_comp_item ON mo_component_lines(component_id);

-- ── View: v_bom_detail ───────────────────────────────────────────────────────
DROP VIEW IF EXISTS v_bom_detail;
CREATE VIEW v_bom_detail AS
SELECT
    b.id,
    b.company_id,
    b.code,
    b.version,
    b.quantity,
    b.is_active,
    b.notes,
    b.created_at,
    b.updated_at,
    i.code AS product_code,
    i.name AS product_name,
    i.item_type AS product_type,
    u.code AS uom_code,
    u.name AS uom_name,
    (SELECT COUNT(*) FROM bom_components bc WHERE bc.bom_id = b.id) AS component_count,
    (SELECT COUNT(*) FROM bom_operations bo WHERE bo.bom_id = b.id) AS operation_count
FROM bill_of_materials b
JOIN items i ON b.product_id = i.id
LEFT JOIN units_of_measure u ON b.uom_id = u.id;

-- ── View: v_manufacturing_orders_detail ──────────────────────────────────────
DROP VIEW IF EXISTS v_manufacturing_orders_detail;
CREATE VIEW v_manufacturing_orders_detail AS
SELECT
    mo.id,
    mo.company_id,
    mo.number,
    mo.status,
    mo.planned_qty,
    mo.produced_qty,
    mo.planned_start,
    mo.planned_end,
    mo.actual_start,
    mo.actual_end,
    mo.material_cost,
    mo.labor_cost,
    mo.overhead_cost,
    mo.total_cost,
    mo.notes,
    mo.created_at,
    mo.updated_at,
    i.code AS product_code,
    i.name AS product_name,
    b.code AS bom_code,
    b.version AS bom_version,
    w.name AS warehouse_name,
    CASE
        WHEN mo.planned_qty > 0
        THEN ROUND((mo.produced_qty / mo.planned_qty) * 100, 1)
        ELSE 0
    END AS progress_pct
FROM manufacturing_orders mo
JOIN items i ON mo.product_id = i.id
JOIN bill_of_materials b ON mo.bom_id = b.id
LEFT JOIN warehouses w ON mo.warehouse_id = w.id;

-- ── Sequence function: next_mo_number ────────────────────────────────────────
DROP FUNCTION IF EXISTS next_mo_number(UUID);
CREATE FUNCTION next_mo_number(p_company_id UUID)
RETURNS TEXT AS $$
DECLARE
    v_year TEXT := TO_CHAR(NOW(), 'YY');
    v_next INT;
BEGIN
    SELECT COALESCE(MAX(
        NULLIF(REGEXP_REPLACE(number, '[^0-9]', '', 'g'), '')::INT
    ), 0) + 1
    INTO v_next
    FROM manufacturing_orders
    WHERE company_id = p_company_id
      AND number LIKE 'MO' || v_year || '%';

    RETURN 'MO' || v_year || LPAD(v_next::TEXT, 5, '0');
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- End of 0005_manufacturing.sql (NO schema_migrations entry)
-- =============================================================================