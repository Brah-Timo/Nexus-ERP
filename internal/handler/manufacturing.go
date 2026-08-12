package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nexus-erp/internal/middleware"
)

// ─── Manufacturing Handler ────────────────────────────────────────────────────

type ManufacturingHandler struct{ db *pgxpool.Pool }

// ─── Domain structs ───────────────────────────────────────────────────────────

type WorkCenter struct {
	ID          string    `json:"id"`
	CompanyID   string    `json:"company_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Capacity    float64   `json:"capacity"`
	CostPerHour float64   `json:"cost_per_hour"`
	AccountID   *string   `json:"account_id,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type BOMComponent struct {
	ID            string   `json:"id"`
	BOMID         string   `json:"bom_id"`
	ComponentID   string   `json:"component_id"`
	ComponentCode string   `json:"component_code,omitempty"`
	ComponentName string   `json:"component_name,omitempty"`
	Quantity      float64  `json:"quantity"`
	UOMID         *string  `json:"uom_id,omitempty"`
	UOMCode       string   `json:"uom_code,omitempty"`
	ScrapPct      float64  `json:"scrap_pct"`
	SortOrder     int      `json:"sort_order"`
}

type BOMOperation struct {
	ID             string  `json:"id"`
	BOMID          string  `json:"bom_id"`
	WorkCenterID   string  `json:"work_center_id"`
	WorkCenterName string  `json:"work_center_name,omitempty"`
	Name           string  `json:"name"`
	DurationHours  float64 `json:"duration_hours"`
	SortOrder      int     `json:"sort_order"`
}

type BOM struct {
	ID             string         `json:"id"`
	CompanyID      string         `json:"company_id"`
	Code           string         `json:"code"`
	ProductID      string         `json:"product_id"`
	ProductCode    string         `json:"product_code,omitempty"`
	ProductName    string         `json:"product_name,omitempty"`
	Version        string         `json:"version"`
	Quantity       float64        `json:"quantity"`
	UOMID          *string        `json:"uom_id,omitempty"`
	UOMCode        string         `json:"uom_code,omitempty"`
	IsActive       bool           `json:"is_active"`
	Notes          *string        `json:"notes,omitempty"`
	ComponentCount int            `json:"component_count"`
	OperationCount int            `json:"operation_count"`
	Components     []BOMComponent `json:"components,omitempty"`
	Operations     []BOMOperation `json:"operations,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type MOComponentLine struct {
	ID            string  `json:"id"`
	MOID          string  `json:"mo_id"`
	ComponentID   string  `json:"component_id"`
	ComponentCode string  `json:"component_code,omitempty"`
	ComponentName string  `json:"component_name,omitempty"`
	RequiredQty   float64 `json:"required_qty"`
	ConsumedQty   float64 `json:"consumed_qty"`
	UnitCost      float64 `json:"unit_cost"`
}

type ManufacturingOrder struct {
	ID              string            `json:"id"`
	CompanyID       string            `json:"company_id"`
	Number          string            `json:"number"`
	BOMID           string            `json:"bom_id"`
	BOMCode         string            `json:"bom_code,omitempty"`
	BOMVersion      string            `json:"bom_version,omitempty"`
	ProductID       string            `json:"product_id"`
	ProductCode     string            `json:"product_code,omitempty"`
	ProductName     string            `json:"product_name,omitempty"`
	WarehouseID     *string           `json:"warehouse_id,omitempty"`
	WarehouseName   string            `json:"warehouse_name,omitempty"`
	PlannedQty      float64           `json:"planned_qty"`
	ProducedQty     float64           `json:"produced_qty"`
	Status          string            `json:"status"`
	PlannedStart    *string           `json:"planned_start,omitempty"`
	PlannedEnd      *string           `json:"planned_end,omitempty"`
	ActualStart     *time.Time        `json:"actual_start,omitempty"`
	ActualEnd       *time.Time        `json:"actual_end,omitempty"`
	MaterialCost    float64           `json:"material_cost"`
	LaborCost       float64           `json:"labor_cost"`
	OverheadCost    float64           `json:"overhead_cost"`
	TotalCost       float64           `json:"total_cost"`
	ProgressPct     float64           `json:"progress_pct"`
	Notes           *string           `json:"notes,omitempty"`
	ComponentLines  []MOComponentLine `json:"component_lines,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// ─── Work Centers ─────────────────────────────────────────────────────────────

func (h *ManufacturingHandler) ListWorkCenters(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	activeOnly := c.Query("active") == "true"

	query := `
		SELECT wc.id, wc.company_id, wc.code, wc.name, wc.capacity,
		       wc.cost_per_hour, wc.account_id::text, wc.is_active, wc.created_at
		FROM work_centers wc
		WHERE wc.company_id = $1`
	if activeOnly {
		query += ` AND wc.is_active = true`
	}
	query += ` ORDER BY wc.name`

	rows, err := h.db.Query(ctx, query, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []WorkCenter
	for rows.Next() {
		var wc WorkCenter
		if err := rows.Scan(
			&wc.ID, &wc.CompanyID, &wc.Code, &wc.Name,
			&wc.Capacity, &wc.CostPerHour, &wc.AccountID,
			&wc.IsActive, &wc.CreatedAt,
		); err != nil {
			continue
		}
		list = append(list, wc)
	}
	if list == nil {
		list = []WorkCenter{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *ManufacturingHandler) CreateWorkCenter(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var wc WorkCenter
	if err := c.ShouldBindJSON(&wc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wc.ID = uuid.NewString()
	wc.CompanyID = companyID
	wc.IsActive = true
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		INSERT INTO work_centers (id, company_id, code, name, capacity, cost_per_hour, account_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		wc.ID, wc.CompanyID, wc.Code, wc.Name,
		wc.Capacity, wc.CostPerHour, wc.AccountID, wc.IsActive,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, wc)
}

func (h *ManufacturingHandler) UpdateWorkCenter(c *gin.Context) {
	id := c.Param("id")
	var wc WorkCenter
	if err := c.ShouldBindJSON(&wc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE work_centers
		SET code=$1, name=$2, capacity=$3, cost_per_hour=$4, account_id=$5, is_active=$6
		WHERE id=$7`,
		wc.Code, wc.Name, wc.Capacity, wc.CostPerHour,
		wc.AccountID, wc.IsActive, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	wc.ID = id
	c.JSON(http.StatusOK, wc)
}

func (h *ManufacturingHandler) DeactivateWorkCenter(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `UPDATE work_centers SET is_active=false WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Work center deactivated"})
}

// ─── Bill of Materials ────────────────────────────────────────────────────────

func (h *ManufacturingHandler) ListBOMs(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	productID := c.Query("product_id")
	activeOnly := c.Query("active") == "true"
	search := c.Query("search")

	query := `
		SELECT
			b.id, b.company_id, b.code, b.product_id,
			i.code, i.name,
			b.version, b.quantity,
			b.uom_id::text, COALESCE(u.code,''),
			b.is_active, b.notes,
			(SELECT COUNT(*) FROM bom_components bc WHERE bc.bom_id = b.id),
			(SELECT COUNT(*) FROM bom_operations  bo WHERE bo.bom_id = b.id),
			b.created_at, b.updated_at
		FROM bill_of_materials b
		JOIN items             i ON b.product_id = i.id
		LEFT JOIN units_of_measure u ON b.uom_id = u.id
		WHERE b.company_id = $1`

	args := []interface{}{companyID}
	idx := 2
	if productID != "" {
		query += fmt.Sprintf(` AND b.product_id = $%d`, idx)
		args = append(args, productID)
		idx++
	}
	if activeOnly {
		query += ` AND b.is_active = true`
	}
	if search != "" {
		query += fmt.Sprintf(` AND (b.code ILIKE $%d OR i.name ILIKE $%d)`, idx, idx)
		args = append(args, "%"+search+"%")
		idx++
	}
	query += ` ORDER BY b.code`

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []BOM
	for rows.Next() {
		var b BOM
		if err := rows.Scan(
			&b.ID, &b.CompanyID, &b.Code, &b.ProductID,
			&b.ProductCode, &b.ProductName,
			&b.Version, &b.Quantity,
			&b.UOMID, &b.UOMCode,
			&b.IsActive, &b.Notes,
			&b.ComponentCount, &b.OperationCount,
			&b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			continue
		}
		list = append(list, b)
	}
	if list == nil {
		list = []BOM{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *ManufacturingHandler) GetBOM(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var b BOM
	err := h.db.QueryRow(ctx, `
		SELECT
			b.id, b.company_id, b.code, b.product_id,
			i.code, i.name,
			b.version, b.quantity,
			b.uom_id::text, COALESCE(u.code,''),
			b.is_active, b.notes,
			0, 0,
			b.created_at, b.updated_at
		FROM bill_of_materials b
		JOIN items             i ON b.product_id = i.id
		LEFT JOIN units_of_measure u ON b.uom_id = u.id
		WHERE b.id = $1`, id,
	).Scan(
		&b.ID, &b.CompanyID, &b.Code, &b.ProductID,
		&b.ProductCode, &b.ProductName,
		&b.Version, &b.Quantity,
		&b.UOMID, &b.UOMCode,
		&b.IsActive, &b.Notes,
		&b.ComponentCount, &b.OperationCount,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "BOM not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load components
	compRows, err := h.db.Query(ctx, `
		SELECT bc.id, bc.bom_id, bc.component_id,
		       i.code, i.name,
		       bc.quantity, bc.uom_id::text, COALESCE(u.code,''),
		       bc.scrap_pct, bc.sort_order
		FROM bom_components bc
		JOIN items i ON bc.component_id = i.id
		LEFT JOIN units_of_measure u ON bc.uom_id = u.id
		WHERE bc.bom_id = $1
		ORDER BY bc.sort_order, i.name`, id,
	)
	if err == nil {
		defer compRows.Close()
		for compRows.Next() {
			var comp BOMComponent
			_ = compRows.Scan(
				&comp.ID, &comp.BOMID, &comp.ComponentID,
				&comp.ComponentCode, &comp.ComponentName,
				&comp.Quantity, &comp.UOMID, &comp.UOMCode,
				&comp.ScrapPct, &comp.SortOrder,
			)
			b.Components = append(b.Components, comp)
		}
	}
	if b.Components == nil {
		b.Components = []BOMComponent{}
	}
	b.ComponentCount = len(b.Components)

	// Load operations
	opRows, err := h.db.Query(ctx, `
		SELECT bo.id, bo.bom_id, bo.work_center_id, wc.name,
		       bo.name, bo.duration_hours, bo.sort_order
		FROM bom_operations bo
		JOIN work_centers   wc ON bo.work_center_id = wc.id
		WHERE bo.bom_id = $1
		ORDER BY bo.sort_order`, id,
	)
	if err == nil {
		defer opRows.Close()
		for opRows.Next() {
			var op BOMOperation
			_ = opRows.Scan(
				&op.ID, &op.BOMID, &op.WorkCenterID, &op.WorkCenterName,
				&op.Name, &op.DurationHours, &op.SortOrder,
			)
			b.Operations = append(b.Operations, op)
		}
	}
	if b.Operations == nil {
		b.Operations = []BOMOperation{}
	}
	b.OperationCount = len(b.Operations)

	c.JSON(http.StatusOK, b)
}

func (h *ManufacturingHandler) CreateBOM(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req struct {
		BOM
		Components []BOMComponent `json:"components"`
		Operations []BOMOperation `json:"operations"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.BOM.ID = uuid.NewString()
	req.BOM.CompanyID = companyID
	req.BOM.IsActive = true

	ctx := context.Background()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO bill_of_materials
		    (id, company_id, code, product_id, version, quantity, uom_id, is_active, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		req.BOM.ID, req.BOM.CompanyID, req.BOM.Code, req.BOM.ProductID,
		req.BOM.Version, req.BOM.Quantity, req.BOM.UOMID, req.BOM.IsActive, req.BOM.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Insert components
	for i, comp := range req.Components {
		comp.ID = uuid.NewString()
		comp.BOMID = req.BOM.ID
		comp.SortOrder = i
		_, err = tx.Exec(ctx, `
			INSERT INTO bom_components (id, bom_id, component_id, quantity, uom_id, scrap_pct, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			comp.ID, comp.BOMID, comp.ComponentID, comp.Quantity,
			comp.UOMID, comp.ScrapPct, comp.SortOrder,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Insert operations
	for i, op := range req.Operations {
		op.ID = uuid.NewString()
		op.BOMID = req.BOM.ID
		op.SortOrder = i
		_, err = tx.Exec(ctx, `
			INSERT INTO bom_operations (id, bom_id, work_center_id, name, duration_hours, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			op.ID, op.BOMID, op.WorkCenterID, op.Name, op.DurationHours, op.SortOrder,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req.BOM)
}

func (h *ManufacturingHandler) UpdateBOM(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		BOM
		Components []BOMComponent `json:"components"`
		Operations []BOMOperation `json:"operations"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE bill_of_materials
		SET code=$1, product_id=$2, version=$3, quantity=$4, uom_id=$5, is_active=$6, notes=$7, updated_at=NOW()
		WHERE id=$8`,
		req.BOM.Code, req.BOM.ProductID, req.BOM.Version, req.BOM.Quantity,
		req.BOM.UOMID, req.BOM.IsActive, req.BOM.Notes, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Replace components
	_, _ = tx.Exec(ctx, `DELETE FROM bom_components WHERE bom_id=$1`, id)
	for i, comp := range req.Components {
		comp.ID = uuid.NewString()
		comp.BOMID = id
		comp.SortOrder = i
		_, err = tx.Exec(ctx, `
			INSERT INTO bom_components (id, bom_id, component_id, quantity, uom_id, scrap_pct, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			comp.ID, comp.BOMID, comp.ComponentID, comp.Quantity,
			comp.UOMID, comp.ScrapPct, comp.SortOrder,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Replace operations
	_, _ = tx.Exec(ctx, `DELETE FROM bom_operations WHERE bom_id=$1`, id)
	for i, op := range req.Operations {
		op.ID = uuid.NewString()
		op.BOMID = id
		op.SortOrder = i
		_, err = tx.Exec(ctx, `
			INSERT INTO bom_operations (id, bom_id, work_center_id, name, duration_hours, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			op.ID, op.BOMID, op.WorkCenterID, op.Name, op.DurationHours, op.SortOrder,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req.BOM.ID = id
	c.JSON(http.StatusOK, req.BOM)
}

func (h *ManufacturingHandler) DeactivateBOM(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `UPDATE bill_of_materials SET is_active=false, updated_at=NOW() WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "BOM deactivated"})
}

// ─── Manufacturing Orders ─────────────────────────────────────────────────────

func (h *ManufacturingHandler) ListOrders(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	status := c.Query("status")
	productID := c.Query("product_id")
	search := c.Query("search")
	limit := 100

	query := `
		SELECT
			mo.id, mo.company_id, mo.number,
			mo.bom_id, b.code, b.version,
			mo.product_id, i.code, i.name,
			mo.warehouse_id::text, COALESCE(w.name,''),
			mo.planned_qty, mo.produced_qty, mo.status,
			mo.planned_start::text, mo.planned_end::text,
			mo.actual_start, mo.actual_end,
			mo.material_cost, mo.labor_cost, mo.overhead_cost, mo.total_cost,
			CASE WHEN mo.planned_qty > 0
			     THEN ROUND((mo.produced_qty / mo.planned_qty) * 100, 1)
			     ELSE 0 END,
			mo.notes,
			mo.created_at, mo.updated_at
		FROM manufacturing_orders mo
		JOIN items             i ON mo.product_id = i.id
		JOIN bill_of_materials b ON mo.bom_id = b.id
		LEFT JOIN warehouses   w ON mo.warehouse_id = w.id
		WHERE mo.company_id = $1`

	args := []interface{}{companyID}
	idx := 2
	if status != "" {
		query += fmt.Sprintf(` AND mo.status = $%d`, idx)
		args = append(args, status)
		idx++
	}
	if productID != "" {
		query += fmt.Sprintf(` AND mo.product_id = $%d`, idx)
		args = append(args, productID)
		idx++
	}
	if search != "" {
		query += fmt.Sprintf(` AND (mo.number ILIKE $%d OR i.name ILIKE $%d)`, idx, idx)
		args = append(args, "%"+search+"%")
		idx++
	}
	query += fmt.Sprintf(` ORDER BY mo.created_at DESC LIMIT %d`, limit)

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []ManufacturingOrder
	for rows.Next() {
		var mo ManufacturingOrder
		if err := rows.Scan(
			&mo.ID, &mo.CompanyID, &mo.Number,
			&mo.BOMID, &mo.BOMCode, &mo.BOMVersion,
			&mo.ProductID, &mo.ProductCode, &mo.ProductName,
			&mo.WarehouseID, &mo.WarehouseName,
			&mo.PlannedQty, &mo.ProducedQty, &mo.Status,
			&mo.PlannedStart, &mo.PlannedEnd,
			&mo.ActualStart, &mo.ActualEnd,
			&mo.MaterialCost, &mo.LaborCost, &mo.OverheadCost, &mo.TotalCost,
			&mo.ProgressPct,
			&mo.Notes,
			&mo.CreatedAt, &mo.UpdatedAt,
		); err != nil {
			continue
		}
		list = append(list, mo)
	}
	if list == nil {
		list = []ManufacturingOrder{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *ManufacturingHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var mo ManufacturingOrder
	err := h.db.QueryRow(ctx, `
		SELECT
			mo.id, mo.company_id, mo.number,
			mo.bom_id, b.code, b.version,
			mo.product_id, i.code, i.name,
			mo.warehouse_id::text, COALESCE(w.name,''),
			mo.planned_qty, mo.produced_qty, mo.status,
			mo.planned_start::text, mo.planned_end::text,
			mo.actual_start, mo.actual_end,
			mo.material_cost, mo.labor_cost, mo.overhead_cost, mo.total_cost,
			CASE WHEN mo.planned_qty > 0
			     THEN ROUND((mo.produced_qty / mo.planned_qty) * 100, 1)
			     ELSE 0 END,
			mo.notes,
			mo.created_at, mo.updated_at
		FROM manufacturing_orders mo
		JOIN items             i ON mo.product_id = i.id
		JOIN bill_of_materials b ON mo.bom_id = b.id
		LEFT JOIN warehouses   w ON mo.warehouse_id = w.id
		WHERE mo.id = $1`, id,
	).Scan(
		&mo.ID, &mo.CompanyID, &mo.Number,
		&mo.BOMID, &mo.BOMCode, &mo.BOMVersion,
		&mo.ProductID, &mo.ProductCode, &mo.ProductName,
		&mo.WarehouseID, &mo.WarehouseName,
		&mo.PlannedQty, &mo.ProducedQty, &mo.Status,
		&mo.PlannedStart, &mo.PlannedEnd,
		&mo.ActualStart, &mo.ActualEnd,
		&mo.MaterialCost, &mo.LaborCost, &mo.OverheadCost, &mo.TotalCost,
		&mo.ProgressPct,
		&mo.Notes,
		&mo.CreatedAt, &mo.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manufacturing order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load component lines
	compRows, _ := h.db.Query(ctx, `
		SELECT mcl.id, mcl.mo_id, mcl.component_id,
		       i.code, i.name,
		       mcl.required_qty, mcl.consumed_qty, mcl.unit_cost
		FROM mo_component_lines mcl
		JOIN items i ON mcl.component_id = i.id
		WHERE mcl.mo_id = $1
		ORDER BY i.name`, id,
	)
	if compRows != nil {
		defer compRows.Close()
		for compRows.Next() {
			var line MOComponentLine
			_ = compRows.Scan(
				&line.ID, &line.MOID, &line.ComponentID,
				&line.ComponentCode, &line.ComponentName,
				&line.RequiredQty, &line.ConsumedQty, &line.UnitCost,
			)
			mo.ComponentLines = append(mo.ComponentLines, line)
		}
	}
	if mo.ComponentLines == nil {
		mo.ComponentLines = []MOComponentLine{}
	}

	c.JSON(http.StatusOK, mo)
}

func (h *ManufacturingHandler) CreateOrder(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req struct {
		ManufacturingOrder
		ComponentLines []MOComponentLine `json:"component_lines"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ManufacturingOrder.ID = uuid.NewString()
	req.ManufacturingOrder.CompanyID = companyID
	if req.ManufacturingOrder.Status == "" {
		req.ManufacturingOrder.Status = "draft"
	}

	ctx := context.Background()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	// Generate number if not provided
	if req.ManufacturingOrder.Number == "" {
		var num string
		_ = tx.QueryRow(ctx, `SELECT next_mo_number($1)`, companyID).Scan(&num)
		req.ManufacturingOrder.Number = num
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO manufacturing_orders
		    (id, company_id, number, bom_id, product_id, warehouse_id,
		     planned_qty, produced_qty, status, planned_start, planned_end,
		     material_cost, labor_cost, overhead_cost, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		req.ManufacturingOrder.ID, req.ManufacturingOrder.CompanyID,
		req.ManufacturingOrder.Number, req.ManufacturingOrder.BOMID,
		req.ManufacturingOrder.ProductID, req.ManufacturingOrder.WarehouseID,
		req.ManufacturingOrder.PlannedQty, 0,
		req.ManufacturingOrder.Status,
		req.ManufacturingOrder.PlannedStart, req.ManufacturingOrder.PlannedEnd,
		req.ManufacturingOrder.MaterialCost, req.ManufacturingOrder.LaborCost,
		req.ManufacturingOrder.OverheadCost, req.ManufacturingOrder.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If component lines not provided, generate from BOM
	if len(req.ComponentLines) == 0 {
		bomRows, _ := tx.Query(ctx, `
			SELECT bc.component_id, bc.quantity * $1, i.cmup_cost
			FROM bom_components bc
			JOIN items i ON bc.component_id = i.id
			WHERE bc.bom_id = $2`,
			req.ManufacturingOrder.PlannedQty, req.ManufacturingOrder.BOMID,
		)
		if bomRows != nil {
			defer bomRows.Close()
			for bomRows.Next() {
				var compID string
				var reqQty, unitCost float64
				_ = bomRows.Scan(&compID, &reqQty, &unitCost)
				_, _ = tx.Exec(ctx, `
					INSERT INTO mo_component_lines (id, mo_id, component_id, required_qty, consumed_qty, unit_cost)
					VALUES ($1,$2,$3,$4,$5,$6)`,
					uuid.NewString(), req.ManufacturingOrder.ID, compID, reqQty, 0, unitCost,
				)
			}
		}
	} else {
		for _, line := range req.ComponentLines {
			line.ID = uuid.NewString()
			line.MOID = req.ManufacturingOrder.ID
			_, err = tx.Exec(ctx, `
				INSERT INTO mo_component_lines (id, mo_id, component_id, required_qty, consumed_qty, unit_cost)
				VALUES ($1,$2,$3,$4,$5,$6)`,
				line.ID, line.MOID, line.ComponentID,
				line.RequiredQty, line.ConsumedQty, line.UnitCost,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req.ManufacturingOrder)
}

func (h *ManufacturingHandler) StartOrder(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE manufacturing_orders
		SET status='in_progress', actual_start=NOW(), updated_at=NOW()
		WHERE id=$1 AND status IN ('draft','planned')`, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Manufacturing order started"})
}

func (h *ManufacturingHandler) CompleteOrder(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ProducedQty  float64 `json:"produced_qty"`
		MaterialCost float64 `json:"material_cost"`
		LaborCost    float64 `json:"labor_cost"`
		OverheadCost float64 `json:"overhead_cost"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE manufacturing_orders
		SET status='completed', produced_qty=$1,
		    material_cost=$2, labor_cost=$3, overhead_cost=$4,
		    actual_end=NOW(), updated_at=NOW()
		WHERE id=$5 AND status='in_progress'`,
		req.ProducedQty, req.MaterialCost, req.LaborCost, req.OverheadCost, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Manufacturing order completed"})
}

func (h *ManufacturingHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE manufacturing_orders
		SET status='cancelled', updated_at=NOW()
		WHERE id=$1 AND status IN ('draft','planned')`, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Manufacturing order cancelled"})
}

func (h *ManufacturingHandler) UpdateOrder(c *gin.Context) {
	id := c.Param("id")
	var req ManufacturingOrder
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE manufacturing_orders
		SET bom_id=$1, product_id=$2, warehouse_id=$3, planned_qty=$4,
		    planned_start=$5, planned_end=$6,
		    material_cost=$7, labor_cost=$8, overhead_cost=$9, notes=$10,
		    updated_at=NOW()
		WHERE id=$11`,
		req.BOMID, req.ProductID, req.WarehouseID, req.PlannedQty,
		req.PlannedStart, req.PlannedEnd,
		req.MaterialCost, req.LaborCost, req.OverheadCost, req.Notes,
		id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req.ID = id
	c.JSON(http.StatusOK, req)
}

// ─── Quality Inspections ──────────────────────────────────────────────────────

func (h *ManufacturingHandler) ListQualityInspections(c *gin.Context) {
	// Schema has no quality_inspections table; return empty
	c.JSON(http.StatusOK, []interface{}{})
}

func (h *ManufacturingHandler) CreateQualityInspection(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "Quality inspection module not yet configured"})
}

// ─── MRP Planning ─────────────────────────────────────────────────────────────

func (h *ManufacturingHandler) RunMRP(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	// Find all planned/draft orders and their component shortfalls
	rows, err := h.db.Query(ctx, `
		SELECT
			mo.id, mo.number, mo.product_id,
			COALESCE(i.code, i.sku, mo.product_id::TEXT) AS product_code,
			i.name,
			mo.planned_qty, mo.planned_start::text,
			mcl.component_id,
			COALESCE(ci.code, ci.sku, mcl.component_id::TEXT) AS component_code,
			ci.name,
			mcl.required_qty,
			COALESCE(sl.qty_available, 0) AS available_qty,
			GREATEST(mcl.required_qty - COALESCE(sl.qty_available, 0), 0) AS shortage_qty
		FROM manufacturing_orders mo
		JOIN items i ON mo.product_id = i.id
		JOIN mo_component_lines mcl ON mcl.mo_id = mo.id
		JOIN items ci ON mcl.component_id = ci.id
		LEFT JOIN (
			SELECT item_id, SUM(COALESCE(qty_on_hand,0) - COALESCE(qty_reserved,0)) AS qty_available
			FROM stock_levels
			WHERE company_id = $1
			GROUP BY item_id
		) sl ON sl.item_id = mcl.component_id
		WHERE mo.company_id = $1
		  AND mo.status IN ('draft', 'planned')
		ORDER BY mo.planned_start NULLS LAST, i.name, ci.name`, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type MRPLine struct {
		MOID          string  `json:"mo_id"`
		MONumber      string  `json:"mo_number"`
		ProductID     string  `json:"product_id"`
		ProductCode   string  `json:"product_code"`
		ProductName   string  `json:"product_name"`
		PlannedQty    float64 `json:"planned_qty"`
		PlannedStart  *string `json:"planned_start"`
		ComponentID   string  `json:"component_id"`
		ComponentCode string  `json:"component_code"`
		ComponentName string  `json:"component_name"`
		RequiredQty   float64 `json:"required_qty"`
		AvailableQty  float64 `json:"available_qty"`
		ShortageQty   float64 `json:"shortage_qty"`
		NeedsToPurch  bool    `json:"needs_to_purchase"`
	}

	var lines []MRPLine
	for rows.Next() {
		var l MRPLine
		_ = rows.Scan(
			&l.MOID, &l.MONumber, &l.ProductID, &l.ProductCode, &l.ProductName,
			&l.PlannedQty, &l.PlannedStart,
			&l.ComponentID, &l.ComponentCode, &l.ComponentName,
			&l.RequiredQty, &l.AvailableQty, &l.ShortageQty,
		)
		l.NeedsToPurch = l.ShortageQty > 0
		lines = append(lines, l)
	}
	if lines == nil {
		lines = []MRPLine{}
	}

	// Summary stats
	totalShortages := 0
	for _, l := range lines {
		if l.ShortageQty > 0 {
			totalShortages++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"lines":           lines,
		"total_lines":     len(lines),
		"total_shortages": totalShortages,
		"generated_at":    time.Now(),
	})
}

func (h *ManufacturingHandler) GetManufacturingDashboard(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	type DashStats struct {
		TotalOrders     int     `json:"total_orders"`
		DraftOrders     int     `json:"draft_orders"`
		InProgressOrders int    `json:"in_progress_orders"`
		CompletedOrders int     `json:"completed_orders"`
		TotalBOMs       int     `json:"total_boms"`
		ActiveWorkCenters int   `json:"active_work_centers"`
		TotalMaterialCost float64 `json:"total_material_cost"`
	}

	var stats DashStats
	_ = h.db.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'draft'),
			COUNT(*) FILTER (WHERE status = 'in_progress'),
			COUNT(*) FILTER (WHERE status = 'completed'),
			COALESCE(SUM(material_cost) FILTER (WHERE status = 'completed'), 0)
		FROM manufacturing_orders
		WHERE company_id = $1`, companyID,
	).Scan(
		&stats.TotalOrders, &stats.DraftOrders, &stats.InProgressOrders,
		&stats.CompletedOrders, &stats.TotalMaterialCost,
	)

	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM bill_of_materials WHERE company_id=$1 AND is_active=true`, companyID).Scan(&stats.TotalBOMs)
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM work_centers WHERE company_id=$1 AND is_active=true`, companyID).Scan(&stats.ActiveWorkCenters)

	// Recent orders
	orderRows, _ := h.db.Query(ctx, `
		SELECT mo.number, i.name, mo.planned_qty, mo.produced_qty, mo.status, mo.created_at
		FROM manufacturing_orders mo
		JOIN items i ON mo.product_id = i.id
		WHERE mo.company_id = $1
		ORDER BY mo.created_at DESC
		LIMIT 10`, companyID,
	)
	var recentOrders []map[string]interface{}
	if orderRows != nil {
		defer orderRows.Close()
		for orderRows.Next() {
			var num, product, status string
			var planned, produced float64
			var createdAt time.Time
			_ = orderRows.Scan(&num, &product, &planned, &produced, &status, &createdAt)
			recentOrders = append(recentOrders, map[string]interface{}{
				"number": num, "product": product,
				"planned_qty": planned, "produced_qty": produced,
				"status": status, "created_at": createdAt,
			})
		}
	}
	if recentOrders == nil {
		recentOrders = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":         stats,
		"recent_orders": recentOrders,
	})
}
