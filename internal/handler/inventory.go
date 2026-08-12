package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nexus-erp/internal/middleware"
)

// =============================================================================
// InventoryHandler — all inventory module endpoints
// =============================================================================

type InventoryHandler struct{ db *pgxpool.Pool }

// =============================================================================
// Item structs
// =============================================================================

type Item struct {
	ID              string   `json:"id"`
	CompanyID       string   `json:"company_id"`
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	Description     *string  `json:"description"`
	CategoryID      *string  `json:"category_id"`
	CategoryName    *string  `json:"category_name"`
	UomID           *string  `json:"uom_id"`
	UomCode         *string  `json:"uom_code"`
	UomName         *string  `json:"uom_name"`
	ItemType        string   `json:"item_type"`
	TrackInventory  bool     `json:"track_inventory"`
	TvaRate         float64  `json:"tva_rate"`
	CostMethod      string   `json:"cost_method"`
	StandardCost    float64  `json:"standard_cost"`
	CmupCost        float64  `json:"cmup_cost"`
	SalePrice       float64  `json:"sale_price"`
	MinStockQty     float64  `json:"min_stock_qty"`
	ReorderQty      float64  `json:"reorder_qty"`
	MaxStockQty     float64  `json:"max_stock_qty"`
	Barcode         *string  `json:"barcode"`
	InternalRef     *string  `json:"internal_ref"`
	HsCode          *string  `json:"hs_code"`
	IsActive        bool     `json:"is_active"`
	CreatedAt       *string  `json:"created_at"`
	UpdatedAt       *string  `json:"updated_at"`
}

type ItemCategory struct {
	ID        string  `json:"id"`
	CompanyID string  `json:"company_id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	ParentID  *string `json:"parent_id"`
}

type UnitOfMeasure struct {
	ID        string  `json:"id"`
	CompanyID string  `json:"company_id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Category  *string `json:"category"`
	Factor    float64 `json:"factor"`
}

// =============================================================================
// Warehouse structs
// =============================================================================

type Warehouse struct {
	ID        string  `json:"id"`
	CompanyID string  `json:"company_id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Address   *string `json:"address"`
	IsActive  bool    `json:"is_active"`
	CreatedAt *string `json:"created_at"`
}

type WarehouseLocation struct {
	ID          string  `json:"id"`
	WarehouseID string  `json:"warehouse_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	ParentID    *string `json:"parent_id"`
}

// =============================================================================
// Stock structs
// =============================================================================

type StockLevel struct {
	ID            string  `json:"id"`
	CompanyID     string  `json:"company_id"`
	ItemID        string  `json:"item_id"`
	ItemCode      string  `json:"item_code"`
	ItemName      string  `json:"item_name"`
	ItemType      string  `json:"item_type"`
	CategoryName  *string `json:"category_name"`
	UomCode       *string `json:"uom_code"`
	WarehouseID   string  `json:"warehouse_id"`
	WarehouseName string  `json:"warehouse_name"`
	LocationID    *string `json:"location_id"`
	QtyOnHand     float64 `json:"qty_on_hand"`
	QtyReserved   float64 `json:"qty_reserved"`
	QtyAvailable  float64 `json:"qty_available"`
	CmupCost      float64 `json:"cmup_cost"`
	TotalValue    float64 `json:"total_value"`
	MinStockQty   float64 `json:"min_stock_qty"`
	ReorderQty    float64 `json:"reorder_qty"`
	MaxStockQty   float64 `json:"max_stock_qty"`
	StockStatus   string  `json:"stock_status"`
	UpdatedAt     *string `json:"updated_at"`
}

type StockMovement struct {
	ID               string  `json:"id"`
	CompanyID        string  `json:"company_id"`
	Number           string  `json:"number"`
	Date             string  `json:"date"`
	Type             string  `json:"type"`
	ItemID           string  `json:"item_id"`
	ItemCode         string  `json:"item_code"`
	ItemName         string  `json:"item_name"`
	WarehouseID      string  `json:"warehouse_id"`
	WarehouseName    string  `json:"warehouse_name"`
	ToWarehouseID    *string `json:"to_warehouse_id"`
	ToWarehouseName  *string `json:"to_warehouse_name"`
	Quantity         float64 `json:"quantity"`
	UnitCost         float64 `json:"unit_cost"`
	TotalCost        float64 `json:"total_cost"`
	Reference        *string `json:"reference"`
	SourceType       *string `json:"source_type"`
	Notes            *string `json:"notes"`
	CreatedBy        *string `json:"created_by"`
	CreatedByName    *string `json:"created_by_name"`
	CreatedAt        string  `json:"created_at"`
}

// =============================================================================
// Inventory Count structs
// =============================================================================

type InventoryCount struct {
	ID          string               `json:"id"`
	CompanyID   string               `json:"company_id"`
	Number      string               `json:"number"`
	Date        string               `json:"date"`
	WarehouseID string               `json:"warehouse_id"`
	WarehouseName *string            `json:"warehouse_name"`
	Status      string               `json:"status"`
	Notes       *string              `json:"notes"`
	ValidatedBy *string              `json:"validated_by"`
	ValidatedAt *string              `json:"validated_at"`
	CreatedBy   *string              `json:"created_by"`
	CreatedAt   string               `json:"created_at"`
	Lines       []InventoryCountLine `json:"lines,omitempty"`
}

type InventoryCountLine struct {
	ID         string   `json:"id"`
	CountID    string   `json:"count_id"`
	ItemID     string   `json:"item_id"`
	ItemCode   string   `json:"item_code"`
	ItemName   string   `json:"item_name"`
	UomCode    *string  `json:"uom_code"`
	LocationID *string  `json:"location_id"`
	BookQty    float64  `json:"book_qty"`
	CountedQty *float64 `json:"counted_qty"`
	Difference float64  `json:"difference"`
	UnitCost   float64  `json:"unit_cost"`
}

// =============================================================================
// Items
// =============================================================================

func (h *InventoryHandler) ListItems(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT
			i.id, i.company_id, i.code, i.name, i.description,
			i.category_id, ic.name AS category_name,
			i.uom_id, u.code AS uom_code, u.name AS uom_name,
			i.item_type, i.track_inventory, i.tva_rate, i.cost_method,
			i.standard_cost, i.cmup_cost, i.sale_price,
			i.min_stock_qty, i.reorder_qty, i.max_stock_qty,
			i.barcode, i.internal_ref, i.hs_code,
			i.is_active,
			TO_CHAR(i.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			TO_CHAR(i.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM items i
		LEFT JOIN item_categories ic ON ic.id = i.category_id
		LEFT JOIN units_of_measure u ON u.id = i.uom_id
		WHERE i.company_id = $1
		ORDER BY i.code`,
		companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var it Item
		_ = rows.Scan(
			&it.ID, &it.CompanyID, &it.Code, &it.Name, &it.Description,
			&it.CategoryID, &it.CategoryName,
			&it.UomID, &it.UomCode, &it.UomName,
			&it.ItemType, &it.TrackInventory, &it.TvaRate, &it.CostMethod,
			&it.StandardCost, &it.CmupCost, &it.SalePrice,
			&it.MinStockQty, &it.ReorderQty, &it.MaxStockQty,
			&it.Barcode, &it.InternalRef, &it.HsCode,
			&it.IsActive, &it.CreatedAt, &it.UpdatedAt,
		)
		items = append(items, it)
	}
	if items == nil {
		items = []Item{}
	}
	c.JSON(http.StatusOK, items)
}

func (h *InventoryHandler) GetItem(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var it Item
	err := h.db.QueryRow(ctx, `
		SELECT
			i.id, i.company_id, i.code, i.name, i.description,
			i.category_id, ic.name AS category_name,
			i.uom_id, u.code AS uom_code, u.name AS uom_name,
			i.item_type, i.track_inventory, i.tva_rate, i.cost_method,
			i.standard_cost, i.cmup_cost, i.sale_price,
			i.min_stock_qty, i.reorder_qty, i.max_stock_qty,
			i.barcode, i.internal_ref, i.hs_code,
			i.is_active,
			TO_CHAR(i.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			TO_CHAR(i.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM items i
		LEFT JOIN item_categories ic ON ic.id = i.category_id
		LEFT JOIN units_of_measure u ON u.id = i.uom_id
		WHERE i.id = $1`, id).Scan(
		&it.ID, &it.CompanyID, &it.Code, &it.Name, &it.Description,
		&it.CategoryID, &it.CategoryName,
		&it.UomID, &it.UomCode, &it.UomName,
		&it.ItemType, &it.TrackInventory, &it.TvaRate, &it.CostMethod,
		&it.StandardCost, &it.CmupCost, &it.SalePrice,
		&it.MinStockQty, &it.ReorderQty, &it.MaxStockQty,
		&it.Barcode, &it.InternalRef, &it.HsCode,
		&it.IsActive, &it.CreatedAt, &it.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}
	c.JSON(http.StatusOK, it)
}

func (h *InventoryHandler) CreateItem(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var body struct {
		Code           string   `json:"code"`
		Name           string   `json:"name"`
		Description    *string  `json:"description"`
		CategoryID     *string  `json:"category_id"`
		UomID          *string  `json:"uom_id"`
		ItemType       string   `json:"item_type"`
		TrackInventory *bool    `json:"track_inventory"`
		TvaRate        *float64 `json:"tva_rate"`
		CostMethod     *string  `json:"cost_method"`
		StandardCost   *float64 `json:"standard_cost"`
		SalePrice      *float64 `json:"sale_price"`
		MinStockQty    *float64 `json:"min_stock_qty"`
		ReorderQty     *float64 `json:"reorder_qty"`
		MaxStockQty    *float64 `json:"max_stock_qty"`
		Barcode        *string  `json:"barcode"`
		InternalRef    *string  `json:"internal_ref"`
		HsCode         *string  `json:"hs_code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Code == "" || body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code and name are required"})
		return
	}
	itemType := body.ItemType
	if itemType == "" {
		itemType = "storable"
	}
	trackInventory := true
	if body.TrackInventory != nil {
		trackInventory = *body.TrackInventory
	}
	tvaRate := 19.0
	if body.TvaRate != nil {
		tvaRate = *body.TvaRate
	}
	costMethod := "cmup"
	if body.CostMethod != nil {
		costMethod = *body.CostMethod
	}
	standardCost := 0.0
	if body.StandardCost != nil {
		standardCost = *body.StandardCost
	}
	salePrice := 0.0
	if body.SalePrice != nil {
		salePrice = *body.SalePrice
	}
	minStockQty := 0.0
	if body.MinStockQty != nil {
		minStockQty = *body.MinStockQty
	}
	reorderQty := 0.0
	if body.ReorderQty != nil {
		reorderQty = *body.ReorderQty
	}
	maxStockQty := 0.0
	if body.MaxStockQty != nil {
		maxStockQty = *body.MaxStockQty
	}

	id := uuid.New().String()
	_, err := h.db.Exec(ctx, `
		INSERT INTO items (
			id, company_id, code, name, description,
			category_id, uom_id, item_type, track_inventory, tva_rate,
			cost_method, standard_cost, sale_price,
			min_stock_qty, reorder_qty, max_stock_qty,
			barcode, internal_ref, hs_code
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13,
			$14, $15, $16,
			$17, $18, $19
		)`,
		id, companyID, body.Code, body.Name, body.Description,
		body.CategoryID, body.UomID, itemType, trackInventory, tvaRate,
		costMethod, standardCost, salePrice,
		minStockQty, reorderQty, maxStockQty,
		body.Barcode, body.InternalRef, body.HsCode,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Item created"})
}

func (h *InventoryHandler) UpdateItem(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var body struct {
		Code           string   `json:"code"`
		Name           string   `json:"name"`
		Description    *string  `json:"description"`
		CategoryID     *string  `json:"category_id"`
		UomID          *string  `json:"uom_id"`
		ItemType       string   `json:"item_type"`
		TrackInventory *bool    `json:"track_inventory"`
		TvaRate        *float64 `json:"tva_rate"`
		CostMethod     *string  `json:"cost_method"`
		StandardCost   *float64 `json:"standard_cost"`
		SalePrice      *float64 `json:"sale_price"`
		MinStockQty    *float64 `json:"min_stock_qty"`
		ReorderQty     *float64 `json:"reorder_qty"`
		MaxStockQty    *float64 `json:"max_stock_qty"`
		Barcode        *string  `json:"barcode"`
		InternalRef    *string  `json:"internal_ref"`
		HsCode         *string  `json:"hs_code"`
		IsActive       *bool    `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.Exec(ctx, `
		UPDATE items SET
			code = COALESCE(NULLIF($2,''), code),
			name = COALESCE(NULLIF($3,''), name),
			description    = $4,
			category_id    = $5,
			uom_id         = $6,
			item_type      = COALESCE(NULLIF($7,''), item_type),
			track_inventory= COALESCE($8, track_inventory),
			tva_rate       = COALESCE($9, tva_rate),
			cost_method    = COALESCE($10, cost_method),
			standard_cost  = COALESCE($11, standard_cost),
			sale_price     = COALESCE($12, sale_price),
			min_stock_qty  = COALESCE($13, min_stock_qty),
			reorder_qty    = COALESCE($14, reorder_qty),
			max_stock_qty  = COALESCE($15, max_stock_qty),
			barcode        = $16,
			internal_ref   = $17,
			hs_code        = $18,
			is_active      = COALESCE($19, is_active),
			updated_at     = NOW()
		WHERE id = $1`,
		id,
		body.Code, body.Name, body.Description,
		body.CategoryID, body.UomID, body.ItemType,
		body.TrackInventory, body.TvaRate, body.CostMethod,
		body.StandardCost, body.SalePrice,
		body.MinStockQty, body.ReorderQty, body.MaxStockQty,
		body.Barcode, body.InternalRef, body.HsCode,
		body.IsActive,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Item updated"})
}

func (h *InventoryHandler) DeactivateItem(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx,
		`UPDATE items SET is_active = FALSE, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Item deactivated"})
}

// =============================================================================
// Categories
// =============================================================================

func (h *InventoryHandler) ListCategories(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx,
		`SELECT id, company_id, code, name, parent_id
		 FROM item_categories
		 WHERE company_id = $1
		 ORDER BY name`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var cats []ItemCategory
	for rows.Next() {
		var cat ItemCategory
		_ = rows.Scan(&cat.ID, &cat.CompanyID, &cat.Code, &cat.Name, &cat.ParentID)
		cats = append(cats, cat)
	}
	if cats == nil {
		cats = []ItemCategory{}
	}
	c.JSON(http.StatusOK, cats)
}

func (h *InventoryHandler) CreateCategory(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var body struct {
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		ParentID *string `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(ctx,
		`INSERT INTO item_categories (id, company_id, code, name, parent_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		id, companyID, body.Code, body.Name, body.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Category created"})
}

// =============================================================================
// Units of Measure
// =============================================================================

func (h *InventoryHandler) ListUnits(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx,
		`SELECT id, company_id, code, name, category, factor
		 FROM units_of_measure
		 WHERE company_id = $1
		 ORDER BY category, name`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var units []UnitOfMeasure
	for rows.Next() {
		var u UnitOfMeasure
		_ = rows.Scan(&u.ID, &u.CompanyID, &u.Code, &u.Name, &u.Category, &u.Factor)
		units = append(units, u)
	}
	if units == nil {
		units = []UnitOfMeasure{}
	}
	c.JSON(http.StatusOK, units)
}

func (h *InventoryHandler) CreateUnit(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var body struct {
		Code     string   `json:"code"`
		Name     string   `json:"name"`
		Category *string  `json:"category"`
		Factor   *float64 `json:"factor"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	factor := 1.0
	if body.Factor != nil {
		factor = *body.Factor
	}
	id := uuid.New().String()
	_, err := h.db.Exec(ctx,
		`INSERT INTO units_of_measure (id, company_id, code, name, category, factor)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id, companyID, body.Code, body.Name, body.Category, factor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Unit created"})
}

// =============================================================================
// Warehouses
// =============================================================================

func (h *InventoryHandler) ListWarehouses(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx,
		`SELECT id, company_id, code, name, address, is_active,
		        TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		 FROM warehouses
		 WHERE company_id = $1
		 ORDER BY name`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var warehouses []Warehouse
	for rows.Next() {
		var w Warehouse
		_ = rows.Scan(&w.ID, &w.CompanyID, &w.Code, &w.Name,
			&w.Address, &w.IsActive, &w.CreatedAt)
		warehouses = append(warehouses, w)
	}
	if warehouses == nil {
		warehouses = []Warehouse{}
	}
	c.JSON(http.StatusOK, warehouses)
}

func (h *InventoryHandler) GetWarehouse(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var w Warehouse
	err := h.db.QueryRow(ctx,
		`SELECT id, company_id, code, name, address, is_active,
		        TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		 FROM warehouses WHERE id = $1`, id).Scan(
		&w.ID, &w.CompanyID, &w.Code, &w.Name,
		&w.Address, &w.IsActive, &w.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse not found"})
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *InventoryHandler) CreateWarehouse(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var body struct {
		Code    string  `json:"code"`
		Name    string  `json:"name"`
		Address *string `json:"address"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Code == "" || body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code and name are required"})
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(ctx,
		`INSERT INTO warehouses (id, company_id, code, name, address)
		 VALUES ($1, $2, $3, $4, $5)`,
		id, companyID, body.Code, body.Name, body.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Warehouse created"})
}

func (h *InventoryHandler) UpdateWarehouse(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var body struct {
		Code     *string `json:"code"`
		Name     *string `json:"name"`
		Address  *string `json:"address"`
		IsActive *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec(ctx, `
		UPDATE warehouses SET
			code      = COALESCE($2, code),
			name      = COALESCE($3, name),
			address   = COALESCE($4, address),
			is_active = COALESCE($5, is_active)
		WHERE id = $1`,
		id, body.Code, body.Name, body.Address, body.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Warehouse updated"})
}

// =============================================================================
// Warehouse Locations
// =============================================================================

func (h *InventoryHandler) ListLocations(c *gin.Context) {
	warehouseID := c.Query("warehouse_id")
	ctx := context.Background()

	query := `SELECT id, warehouse_id, code, name, parent_id
			  FROM warehouse_locations`
	args := []interface{}{}
	if warehouseID != "" {
		query += ` WHERE warehouse_id = $1`
		args = append(args, warehouseID)
	}
	query += ` ORDER BY code`

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var locations []WarehouseLocation
	for rows.Next() {
		var loc WarehouseLocation
		_ = rows.Scan(&loc.ID, &loc.WarehouseID, &loc.Code, &loc.Name, &loc.ParentID)
		locations = append(locations, loc)
	}
	if locations == nil {
		locations = []WarehouseLocation{}
	}
	c.JSON(http.StatusOK, locations)
}

func (h *InventoryHandler) CreateLocation(c *gin.Context) {
	ctx := context.Background()

	var body struct {
		WarehouseID string  `json:"warehouse_id"`
		Code        string  `json:"code"`
		Name        string  `json:"name"`
		ParentID    *string `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(ctx,
		`INSERT INTO warehouse_locations (id, warehouse_id, code, name, parent_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		id, body.WarehouseID, body.Code, body.Name, body.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Location created"})
}

// =============================================================================
// Stock Levels
// =============================================================================

func (h *InventoryHandler) GetStockLevels(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	warehouseID := c.Query("warehouse_id")
	itemID := c.Query("item_id")

	query := `
		SELECT
			sl.id, sl.company_id,
			sl.item_id,      i.code AS item_code,      i.name AS item_name, i.item_type,
			ic.name AS category_name,
			u.code  AS uom_code,
			sl.warehouse_id, w.name AS warehouse_name,
			sl.location_id,
			sl.qty_on_hand, sl.qty_reserved, sl.qty_available,
			sl.cmup_cost,
			ROUND(sl.qty_on_hand * sl.cmup_cost, 2)      AS total_value,
			i.min_stock_qty, i.reorder_qty, i.max_stock_qty,
			CASE
				WHEN sl.qty_available <= 0                          THEN 'out_of_stock'
				WHEN sl.qty_available < i.min_stock_qty             THEN 'low_stock'
				WHEN i.max_stock_qty > 0 AND sl.qty_available > i.max_stock_qty THEN 'over_stock'
				ELSE 'normal'
			END AS stock_status,
			TO_CHAR(sl.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM stock_levels sl
		JOIN items         i  ON i.id  = sl.item_id
		JOIN warehouses    w  ON w.id  = sl.warehouse_id
		LEFT JOIN item_categories ic ON ic.id = i.category_id
		LEFT JOIN units_of_measure u ON u.id  = i.uom_id
		WHERE sl.company_id = $1`

	args := []interface{}{companyID}
	paramIdx := 2

	if warehouseID != "" {
		query += fmt.Sprintf(` AND sl.warehouse_id = $%d`, paramIdx)
		args = append(args, warehouseID)
		paramIdx++
	}
	if itemID != "" {
		query += fmt.Sprintf(` AND sl.item_id = $%d`, paramIdx)
		args = append(args, itemID)
	}
	query += ` ORDER BY i.code, w.name`

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var levels []StockLevel
	for rows.Next() {
		var sl StockLevel
		_ = rows.Scan(
			&sl.ID, &sl.CompanyID,
			&sl.ItemID, &sl.ItemCode, &sl.ItemName, &sl.ItemType,
			&sl.CategoryName, &sl.UomCode,
			&sl.WarehouseID, &sl.WarehouseName,
			&sl.LocationID,
			&sl.QtyOnHand, &sl.QtyReserved, &sl.QtyAvailable,
			&sl.CmupCost, &sl.TotalValue,
			&sl.MinStockQty, &sl.ReorderQty, &sl.MaxStockQty,
			&sl.StockStatus, &sl.UpdatedAt,
		)
		levels = append(levels, sl)
	}
	if levels == nil {
		levels = []StockLevel{}
	}
	c.JSON(http.StatusOK, levels)
}

// =============================================================================
// Stock Movements
// =============================================================================

func (h *InventoryHandler) ListMovements(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	movType := c.Query("type")
	warehouseID := c.Query("warehouse_id")
	itemID := c.Query("item_id")

	query := `
		SELECT
			sm.id, sm.company_id, sm.number,
			TO_CHAR(sm.date, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			sm.type,
			sm.item_id,      i.code AS item_code,   i.name AS item_name,
			sm.warehouse_id, w.name AS warehouse_name,
			sm.to_warehouse_id,
			w2.name AS to_warehouse_name,
			sm.quantity, sm.unit_cost, sm.total_cost,
			sm.reference, sm.source_type, sm.notes,
			sm.created_by,
			COALESCE(u.first_name || ' ' || u.last_name, ''),
			TO_CHAR(sm.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM stock_movements sm
		JOIN items       i  ON i.id  = sm.item_id
		JOIN warehouses  w  ON w.id  = sm.warehouse_id
		LEFT JOIN warehouses w2 ON w2.id = sm.to_warehouse_id
		LEFT JOIN users      u  ON u.id  = sm.created_by
		WHERE sm.company_id = $1`

	args := []interface{}{companyID}
	paramIdx := 2

	if movType != "" {
		query += fmt.Sprintf(` AND sm.type = $%d`, paramIdx)
		args = append(args, movType)
		paramIdx++
	}
	if warehouseID != "" {
		query += fmt.Sprintf(` AND sm.warehouse_id = $%d`, paramIdx)
		args = append(args, warehouseID)
		paramIdx++
	}
	if itemID != "" {
		query += fmt.Sprintf(` AND sm.item_id = $%d`, paramIdx)
		args = append(args, itemID)
	}
	query += ` ORDER BY sm.date DESC, sm.created_at DESC LIMIT 500`

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var movements []StockMovement
	for rows.Next() {
		var m StockMovement
		_ = rows.Scan(
			&m.ID, &m.CompanyID, &m.Number,
			&m.Date, &m.Type,
			&m.ItemID, &m.ItemCode, &m.ItemName,
			&m.WarehouseID, &m.WarehouseName,
			&m.ToWarehouseID, &m.ToWarehouseName,
			&m.Quantity, &m.UnitCost, &m.TotalCost,
			&m.Reference, &m.SourceType, &m.Notes,
			&m.CreatedBy, &m.CreatedByName,
			&m.CreatedAt,
		)
		movements = append(movements, m)
	}
	if movements == nil {
		movements = []StockMovement{}
	}
	c.JSON(http.StatusOK, movements)
}

// generateMovementNumber creates a sequential movement number MOV-YYYY-XXXXXX
func generateMovementNumber(ctx context.Context, db *pgxpool.Pool, companyID string) string {
	year := time.Now().Format("2006")
	var seq int64
	_ = db.QueryRow(ctx, `
		SELECT COALESCE(MAX(
			CAST(SUBSTRING(number FROM 'MOV-\d{4}-(\d+)') AS BIGINT)
		), 0) + 1
		FROM stock_movements
		WHERE company_id = $1 AND number LIKE $2`,
		companyID, "MOV-"+year+"-%",
	).Scan(&seq)
	return fmt.Sprintf("MOV-%s-%06d", year, seq)
}

// AdjustStock — manual inventory adjustment
func (h *InventoryHandler) AdjustStock(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	ctx := context.Background()

	var body struct {
		ItemID      string   `json:"item_id"`
		WarehouseID string   `json:"warehouse_id"`
		LocationID  *string  `json:"location_id"`
		Quantity    float64  `json:"quantity"`
		UnitCost    *float64 `json:"unit_cost"`
		Reference   *string  `json:"reference"`
		Notes       *string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.ItemID == "" || body.WarehouseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id and warehouse_id are required"})
		return
	}

	unitCost := 0.0
	if body.UnitCost != nil {
		unitCost = *body.UnitCost
	}

	number := generateMovementNumber(ctx, h.db, companyID)
	movID := uuid.New().String()

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	// Insert stock movement (total_cost is GENERATED — do not insert it)
	_, err = tx.Exec(ctx, `
		INSERT INTO stock_movements (
			id, company_id, number, date, type,
			item_id, warehouse_id, from_location_id,
			quantity, unit_cost,
			reference, source_type, notes, created_by
		) VALUES (
			$1, $2, $3, NOW(), 'adjustment',
			$4, $5, $6,
			$7, $8,
			$9, 'manual', $10, $11
		)`,
		movID, companyID, number,
		body.ItemID, body.WarehouseID, body.LocationID,
		body.Quantity, unitCost,
		body.Reference, body.Notes, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Upsert stock_levels
	_, err = tx.Exec(ctx, `
		INSERT INTO stock_levels (id, company_id, item_id, warehouse_id, location_id, qty_on_hand, cmup_cost)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (item_id, warehouse_id, COALESCE(location_id, '00000000-0000-0000-0000-000000000000'::UUID))
		DO UPDATE SET
			qty_on_hand = stock_levels.qty_on_hand + $6,
			cmup_cost   = CASE WHEN $7 > 0 THEN $7 ELSE stock_levels.cmup_cost END,
			updated_at  = NOW()`,
		uuid.New().String(), companyID, body.ItemID, body.WarehouseID, body.LocationID,
		body.Quantity, unitCost,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": movID, "number": number, "message": "Adjustment recorded"})
}

// TransferStock — inter-warehouse transfer
func (h *InventoryHandler) TransferStock(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	ctx := context.Background()

	var body struct {
		ItemID        string   `json:"item_id"`
		FromWarehouse string   `json:"warehouse_id"`
		ToWarehouse   string   `json:"to_warehouse_id"`
		FromLocation  *string  `json:"from_location_id"`
		ToLocation    *string  `json:"to_location_id"`
		Quantity      float64  `json:"quantity"`
		UnitCost      *float64 `json:"unit_cost"`
		Reference     *string  `json:"reference"`
		Notes         *string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.ItemID == "" || body.FromWarehouse == "" || body.ToWarehouse == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id, warehouse_id, and to_warehouse_id are required"})
		return
	}
	if body.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be positive"})
		return
	}

	unitCost := 0.0
	if body.UnitCost != nil {
		unitCost = *body.UnitCost
	} else {
		// Read current cmup_cost
		_ = h.db.QueryRow(ctx,
			`SELECT cmup_cost FROM stock_levels WHERE item_id = $1 AND warehouse_id = $2`,
			body.ItemID, body.FromWarehouse).Scan(&unitCost)
	}

	number := generateMovementNumber(ctx, h.db, companyID)
	movID := uuid.New().String()

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	// Stock movement (total_cost GENERATED — do not insert)
	_, err = tx.Exec(ctx, `
		INSERT INTO stock_movements (
			id, company_id, number, date, type,
			item_id, warehouse_id, from_location_id,
			to_warehouse_id, to_location_id,
			quantity, unit_cost,
			reference, source_type, notes, created_by
		) VALUES (
			$1, $2, $3, NOW(), 'transfer',
			$4, $5, $6,
			$7, $8,
			$9, $10,
			$11, 'transfer', $12, $13
		)`,
		movID, companyID, number,
		body.ItemID, body.FromWarehouse, body.FromLocation,
		body.ToWarehouse, body.ToLocation,
		body.Quantity, unitCost,
		body.Reference, body.Notes, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Decrease from source warehouse
	_, err = tx.Exec(ctx, `
		UPDATE stock_levels SET qty_on_hand = qty_on_hand - $1, updated_at = NOW()
		WHERE item_id = $2 AND warehouse_id = $3`,
		body.Quantity, body.ItemID, body.FromWarehouse)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Upsert destination warehouse
	_, err = tx.Exec(ctx, `
		INSERT INTO stock_levels (id, company_id, item_id, warehouse_id, location_id, qty_on_hand, cmup_cost)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (item_id, warehouse_id, COALESCE(location_id, '00000000-0000-0000-0000-000000000000'::UUID))
		DO UPDATE SET
			qty_on_hand = stock_levels.qty_on_hand + $6,
			cmup_cost   = CASE WHEN $7 > 0 THEN $7 ELSE stock_levels.cmup_cost END,
			updated_at  = NOW()`,
		uuid.New().String(), companyID, body.ItemID, body.ToWarehouse, body.ToLocation,
		body.Quantity, unitCost,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": movID, "number": number, "message": "Transfer recorded"})
}

// ListLots — stub (lots/serial management not in v1 schema)
func (h *InventoryHandler) ListLots(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

// =============================================================================
// Inventory Counts
// =============================================================================

func generateCountNumber(ctx context.Context, db *pgxpool.Pool, companyID string) string {
	year := time.Now().Format("2006")
	var seq int64
	_ = db.QueryRow(ctx, `
		SELECT COALESCE(MAX(
			CAST(SUBSTRING(number FROM 'IC-\d{4}-(\d+)') AS BIGINT)
		), 0) + 1
		FROM inventory_counts
		WHERE company_id = $1 AND number LIKE $2`,
		companyID, "IC-"+year+"-%",
	).Scan(&seq)
	return fmt.Sprintf("IC-%s-%06d", year, seq)
}

func (h *InventoryHandler) ListInventoryCounts(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT
			ic.id, ic.company_id, ic.number,
			TO_CHAR(ic.date, 'YYYY-MM-DD'),
			ic.warehouse_id, w.name AS warehouse_name,
			ic.status, ic.notes,
			ic.validated_by,
			TO_CHAR(ic.validated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			ic.created_by,
			TO_CHAR(ic.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM inventory_counts ic
		JOIN warehouses w ON w.id = ic.warehouse_id
		WHERE ic.company_id = $1
		ORDER BY ic.date DESC, ic.created_at DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var counts []InventoryCount
	for rows.Next() {
		var ic InventoryCount
		_ = rows.Scan(
			&ic.ID, &ic.CompanyID, &ic.Number,
			&ic.Date,
			&ic.WarehouseID, &ic.WarehouseName,
			&ic.Status, &ic.Notes,
			&ic.ValidatedBy, &ic.ValidatedAt,
			&ic.CreatedBy, &ic.CreatedAt,
		)
		counts = append(counts, ic)
	}
	if counts == nil {
		counts = []InventoryCount{}
	}
	c.JSON(http.StatusOK, counts)
}

func (h *InventoryHandler) GetInventoryCount(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var ic InventoryCount
	err := h.db.QueryRow(ctx, `
		SELECT
			ic.id, ic.company_id, ic.number,
			TO_CHAR(ic.date, 'YYYY-MM-DD'),
			ic.warehouse_id, w.name AS warehouse_name,
			ic.status, ic.notes,
			ic.validated_by,
			TO_CHAR(ic.validated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			ic.created_by,
			TO_CHAR(ic.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM inventory_counts ic
		JOIN warehouses w ON w.id = ic.warehouse_id
		WHERE ic.id = $1`, id).Scan(
		&ic.ID, &ic.CompanyID, &ic.Number,
		&ic.Date,
		&ic.WarehouseID, &ic.WarehouseName,
		&ic.Status, &ic.Notes,
		&ic.ValidatedBy, &ic.ValidatedAt,
		&ic.CreatedBy, &ic.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory count not found"})
		return
	}

	// Load lines
	rows, err := h.db.Query(ctx, `
		SELECT
			icl.id, icl.count_id,
			icl.item_id, i.code AS item_code, i.name AS item_name,
			u.code AS uom_code,
			icl.location_id,
			icl.book_qty, icl.counted_qty, icl.difference,
			icl.unit_cost
		FROM inventory_count_lines icl
		JOIN items i ON i.id = icl.item_id
		LEFT JOIN units_of_measure u ON u.id = i.uom_id
		WHERE icl.count_id = $1
		ORDER BY i.code`, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var line InventoryCountLine
			_ = rows.Scan(
				&line.ID, &line.CountID,
				&line.ItemID, &line.ItemCode, &line.ItemName,
				&line.UomCode, &line.LocationID,
				&line.BookQty, &line.CountedQty, &line.Difference,
				&line.UnitCost,
			)
			ic.Lines = append(ic.Lines, line)
		}
	}
	if ic.Lines == nil {
		ic.Lines = []InventoryCountLine{}
	}
	c.JSON(http.StatusOK, ic)
}

func (h *InventoryHandler) CreateInventoryCount(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	ctx := context.Background()

	var body struct {
		Date        string `json:"date"`
		WarehouseID string `json:"warehouse_id"`
		Notes       *string `json:"notes"`
		Lines       []struct {
			ItemID     string   `json:"item_id"`
			LocationID *string  `json:"location_id"`
			BookQty    float64  `json:"book_qty"`
			CountedQty *float64 `json:"counted_qty"`
			UnitCost   float64  `json:"unit_cost"`
		} `json:"lines"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.WarehouseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "warehouse_id is required"})
		return
	}
	date := body.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	number := generateCountNumber(ctx, h.db, companyID)
	countID := uuid.New().String()

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO inventory_counts (id, company_id, number, date, warehouse_id, status, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, 'draft', $6, $7)`,
		countID, companyID, number, date, body.WarehouseID, body.Notes, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, line := range body.Lines {
		lineID := uuid.New().String()
		// difference is GENERATED — do not insert it
		_, err = tx.Exec(ctx, `
			INSERT INTO inventory_count_lines (id, count_id, item_id, location_id, book_qty, counted_qty, unit_cost)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			lineID, countID, line.ItemID, line.LocationID,
			line.BookQty, line.CountedQty, line.UnitCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": countID, "number": number, "message": "Inventory count created"})
}

func (h *InventoryHandler) ValidateInventoryCount(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	// Verify draft status
	var status string
	var warehouseID string
	err := h.db.QueryRow(ctx,
		`SELECT status, warehouse_id FROM inventory_counts WHERE id = $1 AND company_id = $2`,
		id, companyID).Scan(&status, &warehouseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory count not found"})
		return
	}
	if status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft counts can be validated"})
		return
	}

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	// Load lines
	rows, err := tx.Query(ctx, `
		SELECT item_id, location_id, counted_qty, difference, unit_cost
		FROM inventory_count_lines
		WHERE count_id = $1 AND counted_qty IS NOT NULL`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type countLine struct {
		ItemID     string
		LocationID *string
		CountedQty float64
		Difference float64
		UnitCost   float64
	}
	var lines []countLine
	for rows.Next() {
		var l countLine
		_ = rows.Scan(&l.ItemID, &l.LocationID, &l.CountedQty, &l.Difference, &l.UnitCost)
		lines = append(lines, l)
	}
	rows.Close()

	// For each line, create an adjustment movement and update stock_levels
	year := time.Now().Format("2006")
	var movSeq int64
	_ = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(CAST(SUBSTRING(number FROM 'MOV-\d{4}-(\d+)') AS BIGINT)), 0)
		FROM stock_movements WHERE company_id = $1 AND number LIKE $2`,
		companyID, "MOV-"+year+"-%").Scan(&movSeq)

	for _, line := range lines {
		if line.Difference == 0 {
			continue
		}
		movSeq++
		movNumber := fmt.Sprintf("MOV-%s-%06d", year, movSeq)
		movID := uuid.New().String()

		_, err = tx.Exec(ctx, `
			INSERT INTO stock_movements (
				id, company_id, number, date, type,
				item_id, warehouse_id, from_location_id,
				quantity, unit_cost,
				reference, source_type, source_id, created_by
			) VALUES (
				$1, $2, $3, NOW(), 'adjustment',
				$4, $5, $6,
				$7, $8,
				$9, 'inventory_count', $10, $11
			)`,
			movID, companyID, movNumber,
			line.ItemID, warehouseID, line.LocationID,
			line.Difference, line.UnitCost,
			"IC-VALIDATE", id, userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Upsert stock level
		_, err = tx.Exec(ctx, `
			INSERT INTO stock_levels (id, company_id, item_id, warehouse_id, location_id, qty_on_hand, cmup_cost)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (item_id, warehouse_id, COALESCE(location_id, '00000000-0000-0000-0000-000000000000'::UUID))
			DO UPDATE SET
				qty_on_hand = $6,
				cmup_cost   = CASE WHEN $7 > 0 THEN $7 ELSE stock_levels.cmup_cost END,
				updated_at  = NOW()`,
			uuid.New().String(), companyID, line.ItemID, warehouseID, line.LocationID,
			line.CountedQty, line.UnitCost,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Mark count as validated
	_, err = tx.Exec(ctx, `
		UPDATE inventory_counts
		SET status = 'validated', validated_by = $1, validated_at = NOW()
		WHERE id = $2`, userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Inventory count validated and stock updated"})
}

// =============================================================================
// Reports
// =============================================================================

func (h *InventoryHandler) ValuationReport(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT
			sl.item_id,
			i.code AS item_code,
			i.name AS item_name,
			i.item_type,
			COALESCE(ic.name, '') AS category_name,
			COALESCE(u.code, '') AS uom_code,
			SUM(sl.qty_on_hand)    AS total_qty_on_hand,
			SUM(sl.qty_available)  AS total_qty_available,
			AVG(sl.cmup_cost)      AS avg_cmup_cost,
			SUM(sl.qty_on_hand * sl.cmup_cost) AS total_value
		FROM stock_levels sl
		JOIN items i ON i.id = sl.item_id
		LEFT JOIN item_categories ic ON ic.id = i.category_id
		LEFT JOIN units_of_measure u ON u.id = i.uom_id
		WHERE sl.company_id = $1
		GROUP BY sl.item_id, i.code, i.name, i.item_type, ic.name, u.code
		ORDER BY total_value DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	type ValuationRow struct {
		ItemID          string  `json:"item_id"`
		ItemCode        string  `json:"item_code"`
		ItemName        string  `json:"item_name"`
		ItemType        string  `json:"item_type"`
		CategoryName    string  `json:"category_name"`
		UomCode         string  `json:"uom_code"`
		TotalQtyOnHand  float64 `json:"total_qty_on_hand"`
		TotalQtyAvail   float64 `json:"total_qty_available"`
		AvgCmupCost     float64 `json:"avg_cmup_cost"`
		TotalValue      float64 `json:"total_value"`
	}

	var report []ValuationRow
	for rows.Next() {
		var r ValuationRow
		_ = rows.Scan(
			&r.ItemID, &r.ItemCode, &r.ItemName, &r.ItemType,
			&r.CategoryName, &r.UomCode,
			&r.TotalQtyOnHand, &r.TotalQtyAvail,
			&r.AvgCmupCost, &r.TotalValue,
		)
		report = append(report, r)
	}
	if report == nil {
		report = []ValuationRow{}
	}
	c.JSON(http.StatusOK, report)
}

func (h *InventoryHandler) GetInventoryDashboard(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	type DashKPI struct {
		TotalItems        int     `json:"total_items"`
		ActiveItems       int     `json:"active_items"`
		TotalWarehouses   int     `json:"total_warehouses"`
		TotalStockValue   float64 `json:"total_stock_value"`
		LowStockCount     int     `json:"low_stock_count"`
		OutOfStockCount   int     `json:"out_of_stock_count"`
		MovementsThisMonth int    `json:"movements_this_month"`
	}

	var kpi DashKPI
	_ = h.db.QueryRow(ctx,
		`SELECT COUNT(*), SUM(CASE WHEN is_active THEN 1 ELSE 0 END)
		 FROM items WHERE company_id = $1`, companyID).
		Scan(&kpi.TotalItems, &kpi.ActiveItems)

	_ = h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM warehouses WHERE company_id = $1 AND is_active = TRUE`, companyID).
		Scan(&kpi.TotalWarehouses)

	_ = h.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(qty_on_hand * cmup_cost), 0)
		 FROM stock_levels WHERE company_id = $1`, companyID).
		Scan(&kpi.TotalStockValue)

	_ = h.db.QueryRow(ctx, `
		SELECT
			SUM(CASE WHEN sl.qty_available < i.min_stock_qty AND sl.qty_available > 0 THEN 1 ELSE 0 END),
			SUM(CASE WHEN sl.qty_available <= 0 THEN 1 ELSE 0 END)
		FROM stock_levels sl
		JOIN items i ON i.id = sl.item_id
		WHERE sl.company_id = $1`, companyID).
		Scan(&kpi.LowStockCount, &kpi.OutOfStockCount)

	_ = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM stock_movements
		WHERE company_id = $1
		  AND date >= DATE_TRUNC('month', NOW())`, companyID).
		Scan(&kpi.MovementsThisMonth)

	c.JSON(http.StatusOK, kpi)
}
