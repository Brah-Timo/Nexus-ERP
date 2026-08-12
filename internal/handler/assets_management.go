package handler

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nexus-erp/internal/middleware"
)

// ─── Assets Handler ────────────────────────────────────────────────────────────

type AssetsHandler struct{ db *pgxpool.Pool }

// nullUUIDAsset returns nil for empty/null strings, else the string value.
// Uses a distinct name to avoid redeclaring the one in helpdesk.go.
func nullUUIDAsset(s string) interface{} {
	if s == "" || strings.ToLower(s) == "null" {
		return nil
	}
	return s
}

// ── Dashboard ─────────────────────────────────────────────────────────────────

func (h *AssetsHandler) GetDashboard(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	// KPI totals
	var totalAssets int
	var totalCost, totalNBV, totalAccumDep float64
	_ = h.db.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(purchase_cost),0),
			COALESCE(SUM(net_book_value),0),
			COALESCE(SUM(accumulated_depreciation),0)
		FROM fixed_assets
		WHERE company_id = $1 AND status NOT IN ('disposed','sold','written_off')
	`, companyID).Scan(&totalAssets, &totalCost, &totalNBV, &totalAccumDep)

	// Assets by status
	statusRows, err := h.db.Query(ctx, `
		SELECT status::text, COUNT(*), COALESCE(SUM(net_book_value),0)
		FROM fixed_assets
		WHERE company_id = $1
		GROUP BY status ORDER BY COUNT(*) DESC
	`, companyID)
	var byStatus []map[string]interface{}
	if err == nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var s string
			var cnt int
			var val float64
			_ = statusRows.Scan(&s, &cnt, &val)
			byStatus = append(byStatus, map[string]interface{}{
				"status": s, "count": cnt, "value": val,
			})
		}
	}
	if byStatus == nil {
		byStatus = []map[string]interface{}{}
	}

	// Assets by category
	catRows, err := h.db.Query(ctx, `
		SELECT COALESCE(ac.name,'Uncategorized'), COUNT(fa.id), COALESCE(SUM(fa.net_book_value),0)
		FROM fixed_assets fa
		LEFT JOIN asset_categories ac ON ac.id = fa.category_id
		WHERE fa.company_id = $1 AND fa.status NOT IN ('disposed','sold','written_off')
		GROUP BY ac.name ORDER BY COUNT(fa.id) DESC
		LIMIT 10
	`, companyID)
	var byCategory []map[string]interface{}
	if err == nil {
		defer catRows.Close()
		for catRows.Next() {
			var name string
			var cnt int
			var val float64
			_ = catRows.Scan(&name, &cnt, &val)
			byCategory = append(byCategory, map[string]interface{}{
				"category": name, "count": cnt, "value": val,
			})
		}
	}
	if byCategory == nil {
		byCategory = []map[string]interface{}{}
	}

	// Assets by location
	locRows, err := h.db.Query(ctx, `
		SELECT COALESCE(al.name,'Unknown'), COUNT(fa.id), COALESCE(SUM(fa.net_book_value),0)
		FROM fixed_assets fa
		LEFT JOIN asset_locations al ON al.id = fa.location_id
		WHERE fa.company_id = $1 AND fa.status NOT IN ('disposed','sold','written_off')
		GROUP BY al.name ORDER BY COUNT(fa.id) DESC
		LIMIT 10
	`, companyID)
	var byLocation []map[string]interface{}
	if err == nil {
		defer locRows.Close()
		for locRows.Next() {
			var name string
			var cnt int
			var val float64
			_ = locRows.Scan(&name, &cnt, &val)
			byLocation = append(byLocation, map[string]interface{}{
				"location": name, "count": cnt, "value": val,
			})
		}
	}
	if byLocation == nil {
		byLocation = []map[string]interface{}{}
	}

	// Recent assets (last 10)
	recentRows, err := h.db.Query(ctx, `
		SELECT fa.id, fa.asset_number, fa.name,
		       COALESCE(ac.name,''), COALESCE(al.name,''),
		       fa.status::text, fa.purchase_cost, fa.net_book_value,
		       fa.purchase_date
		FROM fixed_assets fa
		LEFT JOIN asset_categories ac ON ac.id = fa.category_id
		LEFT JOIN asset_locations al ON al.id = fa.location_id
		WHERE fa.company_id = $1
		ORDER BY fa.created_at DESC LIMIT 10
	`, companyID)
	var recentAssets []map[string]interface{}
	if err == nil {
		defer recentRows.Close()
		for recentRows.Next() {
			var id, num, name, cat, loc, status string
			var cost, nbv float64
			var purchaseDate interface{}
			_ = recentRows.Scan(&id, &num, &name, &cat, &loc, &status, &cost, &nbv, &purchaseDate)
			recentAssets = append(recentAssets, map[string]interface{}{
				"id": id, "asset_number": num, "name": name,
				"category": cat, "location": loc, "status": status,
				"purchase_cost": cost, "net_book_value": nbv, "purchase_date": purchaseDate,
			})
		}
	}
	if recentAssets == nil {
		recentAssets = []map[string]interface{}{}
	}

	// Due maintenance in next 30 days
	var dueMaintenance int
	_ = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM asset_maintenance_records
		WHERE company_id = $1
		  AND status NOT IN ('completed','cancelled')
		  AND scheduled_date <= CURRENT_DATE + INTERVAL '30 days'
	`, companyID).Scan(&dueMaintenance)

	// Pending transfers
	var pendingTransfers int
	_ = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM asset_transfers
		WHERE company_id = $1 AND status IN ('pending','approved')
	`, companyID).Scan(&pendingTransfers)

	// Depreciation this month
	var monthlyDepreciation float64
	now := time.Now()
	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(depreciation_amount),0)
		FROM asset_depreciation_schedules ads
		JOIN fixed_assets fa ON fa.id = ads.asset_id
		WHERE fa.company_id = $1
		  AND ads.period_year = $2
		  AND ads.period_month = $3
	`, companyID, now.Year(), int(now.Month())).Scan(&monthlyDepreciation)

	depRate := 0.0
	if totalCost > 0 {
		depRate = math.Round((totalAccumDep/totalCost)*10000) / 100
	}

	c.JSON(http.StatusOK, gin.H{
		"total_assets":          totalAssets,
		"total_cost":            totalCost,
		"total_net_book_value":  totalNBV,
		"total_accum_dep":       totalAccumDep,
		"depreciation_rate":     depRate,
		"monthly_depreciation":  monthlyDepreciation,
		"due_maintenance":       dueMaintenance,
		"pending_transfers":     pendingTransfers,
		"by_status":             byStatus,
		"by_category":           byCategory,
		"by_location":           byLocation,
		"recent_assets":         recentAssets,
	})
}

// ── Fixed Assets ──────────────────────────────────────────────────────────────

func (h *AssetsHandler) ListAssets(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	where := []string{"fa.company_id = $1"}
	args := []interface{}{companyID}
	n := 2

	if s := c.Query("status"); s != "" {
		where = append(where, fmt.Sprintf("fa.status = $%d::asset_status", n))
		args = append(args, s)
		n++
	}
	if cat := c.Query("category_id"); cat != "" {
		where = append(where, fmt.Sprintf("fa.category_id = $%d", n))
		args = append(args, cat)
		n++
	}
	if loc := c.Query("location_id"); loc != "" {
		where = append(where, fmt.Sprintf("fa.location_id = $%d", n))
		args = append(args, loc)
		n++
	}
	if q := c.Query("q"); q != "" {
		where = append(where, fmt.Sprintf("(fa.name ILIKE $%d OR fa.asset_number ILIKE $%d OR fa.serial_number ILIKE $%d)", n, n, n))
		args = append(args, "%"+q+"%")
		n++
	}
	_ = n

	query := `
		SELECT fa.id, fa.asset_number, fa.name,
		       COALESCE(fa.description,''), fa.status::text, fa.condition::text,
		       fa.depreciation_method::text,
		       COALESCE(fa.category_id::text,''), COALESCE(ac.name,''),
		       COALESCE(fa.location_id::text,''), COALESCE(al.name,''),
		       fa.purchase_cost, fa.salvage_value, fa.useful_life_years,
		       fa.depreciation_rate, fa.accumulated_depreciation, fa.net_book_value,
		       fa.purchase_date, COALESCE(fa.in_service_date,''),
		       COALESCE(fa.serial_number,''), COALESCE(fa.model,''),
		       COALESCE(fa.manufacturer,''), COALESCE(fa.warranty_expiry_date,''),
		       COALESCE(fa.notes,''), fa.is_depreciable, fa.created_at
		FROM fixed_assets fa
		LEFT JOIN asset_categories ac ON ac.id = fa.category_id
		LEFT JOIN asset_locations al ON al.id = fa.location_id
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY fa.asset_number`

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var id, assetNum, name, desc, status, condition, depMethod string
		var catID, catName, locID, locName string
		var cost, salvage, useful, depRate, accumDep, nbv float64
		var purchaseDate, inServiceDate interface{}
		var serial, model, manufacturer, warrantyExpiry, notes string
		var isDepreciable bool
		var createdAt interface{}

		_ = rows.Scan(&id, &assetNum, &name, &desc, &status, &condition, &depMethod,
			&catID, &catName, &locID, &locName,
			&cost, &salvage, &useful, &depRate, &accumDep, &nbv,
			&purchaseDate, &inServiceDate,
			&serial, &model, &manufacturer, &warrantyExpiry, &notes,
			&isDepreciable, &createdAt)

		list = append(list, map[string]interface{}{
			"id": id, "asset_number": assetNum, "name": name, "description": desc,
			"status": status, "condition": condition, "depreciation_method": depMethod,
			"category_id": catID, "category_name": catName,
			"location_id": locID, "location_name": locName,
			"purchase_cost": cost, "salvage_value": salvage,
			"useful_life_years": useful, "depreciation_rate": depRate,
			"accumulated_depreciation": accumDep, "net_book_value": nbv,
			"purchase_date": purchaseDate, "in_service_date": inServiceDate,
			"serial_number": serial, "model": model, "manufacturer": manufacturer,
			"warranty_expiry_date": warrantyExpiry, "notes": notes,
			"is_depreciable": isDepreciable, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *AssetsHandler) GetAsset(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var assetNum, name, desc, status, condition, depMethod string
	var catID, catName, locID, locName string
	var cost, salvage, useful, depRate, accumDep, nbv float64
	var purchaseDate, inServiceDate interface{}
	var serial, model, manufacturer, warrantyExpiry, notes string
	var isDepreciable bool
	var createdAt, updatedAt interface{}

	err := h.db.QueryRow(ctx, `
		SELECT fa.id, fa.asset_number, fa.name,
		       COALESCE(fa.description,''), fa.status::text, fa.condition::text,
		       fa.depreciation_method::text,
		       COALESCE(fa.category_id::text,''), COALESCE(ac.name,''),
		       COALESCE(fa.location_id::text,''), COALESCE(al.name,''),
		       fa.purchase_cost, fa.salvage_value, fa.useful_life_years,
		       fa.depreciation_rate, fa.accumulated_depreciation, fa.net_book_value,
		       fa.purchase_date, COALESCE(fa.in_service_date,''),
		       COALESCE(fa.serial_number,''), COALESCE(fa.model,''),
		       COALESCE(fa.manufacturer,''), COALESCE(fa.warranty_expiry_date,''),
		       COALESCE(fa.notes,''), fa.is_depreciable,
		       fa.created_at, fa.updated_at
		FROM fixed_assets fa
		LEFT JOIN asset_categories ac ON ac.id = fa.category_id
		LEFT JOIN asset_locations al ON al.id = fa.location_id
		WHERE fa.id = $1 AND fa.company_id = $2
	`, id, companyID).Scan(
		&id, &assetNum, &name, &desc, &status, &condition, &depMethod,
		&catID, &catName, &locID, &locName,
		&cost, &salvage, &useful, &depRate, &accumDep, &nbv,
		&purchaseDate, &inServiceDate,
		&serial, &model, &manufacturer, &warrantyExpiry, &notes,
		&isDepreciable, &createdAt, &updatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
		return
	}

	// Fetch depreciation schedule
	schedRows, _ := h.db.Query(ctx, `
		SELECT id, period_year, period_month,
		       opening_book_value, depreciation_amount,
		       accumulated_depreciation, closing_book_value, is_posted
		FROM asset_depreciation_schedules
		WHERE asset_id = $1
		ORDER BY period_year, period_month
	`, id)
	var schedule []map[string]interface{}
	if schedRows != nil {
		defer schedRows.Close()
		for schedRows.Next() {
			var sid string
			var yr, mo int
			var open, dep, accum, close float64
			var posted bool
			_ = schedRows.Scan(&sid, &yr, &mo, &open, &dep, &accum, &close, &posted)
			schedule = append(schedule, map[string]interface{}{
				"id": sid, "period_year": yr, "period_month": mo,
				"opening_book_value": open, "depreciation_amount": dep,
				"accumulated_depreciation": accum, "closing_book_value": close,
				"is_posted": posted,
			})
		}
	}
	if schedule == nil {
		schedule = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id, "asset_number": assetNum, "name": name, "description": desc,
		"status": status, "condition": condition, "depreciation_method": depMethod,
		"category_id": catID, "category_name": catName,
		"location_id": locID, "location_name": locName,
		"purchase_cost": cost, "salvage_value": salvage,
		"useful_life_years": useful, "depreciation_rate": depRate,
		"accumulated_depreciation": accumDep, "net_book_value": nbv,
		"purchase_date": purchaseDate, "in_service_date": inServiceDate,
		"serial_number": serial, "model": model, "manufacturer": manufacturer,
		"warranty_expiry_date": warrantyExpiry, "notes": notes,
		"is_depreciable": isDepreciable,
		"created_at": createdAt, "updated_at": updatedAt,
		"depreciation_schedule": schedule,
	})
}

func (h *AssetsHandler) CreateAsset(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate asset number
	var seq int64
	_ = h.db.QueryRow(ctx, `SELECT nextval('asset_number_seq')`).Scan(&seq)
	assetNumber := fmt.Sprintf("AST-%06d", seq)

	id := uuid.NewString()

	depMethod := strValDefault(req, "depreciation_method", "straight_line")
	status := strValDefault(req, "status", "active")
	condition := strValDefault(req, "condition", "good")

	cost := assetFloatVal(req, "purchase_cost")
	salvage := assetFloatVal(req, "salvage_value")
	useful := assetFloatValDefault(req, "useful_life_years", 5.0)

	// Auto-calculate depreciation rate for straight-line if not provided
	depRate := assetFloatVal(req, "depreciation_rate")
	if depRate == 0 && useful > 0 {
		depRate = math.Round((1.0/useful)*10000) / 100
	}

	purchaseDate := strVal(req, "purchase_date")
	if purchaseDate == "" {
		purchaseDate = time.Now().Format("2006-01-02")
	}

	_, err := h.db.Exec(ctx, `
		INSERT INTO fixed_assets (
			id, company_id, asset_number, name, description,
			category_id, location_id, status, condition,
			purchase_cost, salvage_value, useful_life_years,
			depreciation_method, depreciation_rate, accumulated_depreciation,
			purchase_date, in_service_date,
			serial_number, model, manufacturer, warranty_expiry_date,
			notes, is_depreciable
		) VALUES (
			$1,$2,$3,$4,$5,
			$6,$7,$8::asset_status,$9::asset_condition,
			$10,$11,$12,
			$13::depreciation_method,$14,0,
			$15,$16,
			$17,$18,$19,$20,
			$21,$22
		)`,
		id, companyID, assetNumber,
		strVal(req, "name"), strVal(req, "description"),
		nullUUIDAsset(strVal(req, "category_id")),
		nullUUIDAsset(strVal(req, "location_id")),
		status, condition,
		cost, salvage, useful,
		depMethod, depRate,
		purchaseDate,
		assetNullableStr(strVal(req, "in_service_date")),
		strVal(req, "serial_number"), strVal(req, "model"),
		strVal(req, "manufacturer"),
		assetNullableStr(strVal(req, "warranty_expiry_date")),
		strVal(req, "notes"),
		assetBoolValDefault(req, "is_depreciable", true),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req["id"] = id
	req["asset_number"] = assetNumber
	c.JSON(http.StatusCreated, req)
}

func (h *AssetsHandler) UpdateAsset(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	depMethod := strValDefault(req, "depreciation_method", "straight_line")
	status := strValDefault(req, "status", "active")
	condition := strValDefault(req, "condition", "good")
	cost := assetFloatVal(req, "purchase_cost")
	salvage := assetFloatVal(req, "salvage_value")
	useful := assetFloatValDefault(req, "useful_life_years", 5.0)
	depRate := assetFloatVal(req, "depreciation_rate")
	if depRate == 0 && useful > 0 {
		depRate = math.Round((1.0/useful)*10000) / 100
	}

	_, err := h.db.Exec(ctx, `
		UPDATE fixed_assets SET
			name = $1, description = $2,
			category_id = $3, location_id = $4,
			status = $5::asset_status, condition = $6::asset_condition,
			purchase_cost = $7, salvage_value = $8, useful_life_years = $9,
			depreciation_method = $10::depreciation_method, depreciation_rate = $11,
			purchase_date = $12, in_service_date = $13,
			serial_number = $14, model = $15, manufacturer = $16,
			warranty_expiry_date = $17, notes = $18, is_depreciable = $19,
			updated_at = NOW()
		WHERE id = $20 AND company_id = $21
	`,
		strVal(req, "name"), strVal(req, "description"),
		nullUUIDAsset(strVal(req, "category_id")),
		nullUUIDAsset(strVal(req, "location_id")),
		status, condition,
		cost, salvage, useful,
		depMethod, depRate,
		strVal(req, "purchase_date"),
		assetNullableStr(strVal(req, "in_service_date")),
		strVal(req, "serial_number"), strVal(req, "model"),
		strVal(req, "manufacturer"),
		assetNullableStr(strVal(req, "warranty_expiry_date")),
		strVal(req, "notes"),
		assetBoolValDefault(req, "is_depreciable", true),
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req["id"] = id
	c.JSON(http.StatusOK, req)
}

func (h *AssetsHandler) DeleteAsset(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	// Only allow deletion if not active (must be disposed/written_off/sold)
	_, err := h.db.Exec(ctx,
		`DELETE FROM fixed_assets WHERE id = $1 AND company_id = $2`,
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Asset deleted"})
}

func (h *AssetsHandler) DisposeAsset(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req struct {
		Status         string  `json:"status"` // disposed, sold, written_off
		DisposalDate   string  `json:"disposal_date"`
		DisposalAmount float64 `json:"disposal_amount"`
		Reason         string  `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status == "" {
		req.Status = "disposed"
	}
	if req.DisposalDate == "" {
		req.DisposalDate = time.Now().Format("2006-01-02")
	}

	_, err := h.db.Exec(ctx, `
		UPDATE fixed_assets SET
			status = $1::asset_status,
			notes = CASE WHEN notes = '' THEN $2 ELSE notes || E'\n' || $2 END,
			updated_at = NOW()
		WHERE id = $3 AND company_id = $4
	`, req.Status,
		fmt.Sprintf("Disposed on %s. Amount: %.2f. Reason: %s", req.DisposalDate, req.DisposalAmount, req.Reason),
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Asset disposed", "status": req.Status})
}

// ── Asset Categories ──────────────────────────────────────────────────────────

func (h *AssetsHandler) ListCategories(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT ac.id, ac.name, COALESCE(ac.description,''),
		       ac.depreciation_method::text, ac.useful_life_years,
		       ac.depreciation_rate, ac.is_active,
		       COALESCE(ac.parent_id::text,''), COALESCE(p.name,''),
		       COUNT(fa.id) AS asset_count,
		       COALESCE(SUM(fa.net_book_value),0),
		       ac.created_at
		FROM asset_categories ac
		LEFT JOIN asset_categories p ON p.id = ac.parent_id
		LEFT JOIN fixed_assets fa ON fa.category_id = ac.id AND fa.company_id = $1
		WHERE ac.company_id = $1
		GROUP BY ac.id, ac.name, ac.description, ac.depreciation_method,
		         ac.useful_life_years, ac.depreciation_rate, ac.is_active,
		         ac.parent_id, p.name, ac.created_at
		ORDER BY ac.name
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var id, name, desc, depMethod string
		var useful, depRate float64
		var isActive bool
		var parentID, parentName string
		var assetCount int
		var totalNBV float64
		var createdAt interface{}
		_ = rows.Scan(&id, &name, &desc, &depMethod, &useful, &depRate, &isActive,
			&parentID, &parentName, &assetCount, &totalNBV, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "name": name, "description": desc,
			"depreciation_method": depMethod,
			"useful_life_years": useful, "depreciation_rate": depRate,
			"is_active": isActive,
			"parent_id": parentID, "parent_name": parentName,
			"asset_count": assetCount, "total_net_book_value": totalNBV,
			"created_at": createdAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *AssetsHandler) CreateCategory(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.NewString()
	depMethod := strValDefault(req, "depreciation_method", "straight_line")
	useful := assetFloatValDefault(req, "useful_life_years", 5.0)
	depRate := assetFloatVal(req, "depreciation_rate")
	if depRate == 0 && useful > 0 {
		depRate = math.Round((1.0/useful)*10000) / 100
	}

	_, err := h.db.Exec(ctx, `
		INSERT INTO asset_categories (
			id, company_id, name, description, parent_id,
			depreciation_method, useful_life_years, depreciation_rate, is_active
		) VALUES ($1,$2,$3,$4,$5,$6::depreciation_method,$7,$8,$9)
	`, id, companyID,
		strVal(req, "name"), strVal(req, "description"),
		nullUUIDAsset(strVal(req, "parent_id")),
		depMethod, useful, depRate,
		assetBoolValDefault(req, "is_active", true),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req["id"] = id
	c.JSON(http.StatusCreated, req)
}

func (h *AssetsHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	depMethod := strValDefault(req, "depreciation_method", "straight_line")
	useful := assetFloatValDefault(req, "useful_life_years", 5.0)
	depRate := assetFloatVal(req, "depreciation_rate")
	if depRate == 0 && useful > 0 {
		depRate = math.Round((1.0/useful)*10000) / 100
	}

	_, err := h.db.Exec(ctx, `
		UPDATE asset_categories SET
			name = $1, description = $2, parent_id = $3,
			depreciation_method = $4::depreciation_method,
			useful_life_years = $5, depreciation_rate = $6,
			is_active = $7, updated_at = NOW()
		WHERE id = $8 AND company_id = $9
	`,
		strVal(req, "name"), strVal(req, "description"),
		nullUUIDAsset(strVal(req, "parent_id")),
		depMethod, useful, depRate,
		assetBoolValDefault(req, "is_active", true),
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req["id"] = id
	c.JSON(http.StatusOK, req)
}

func (h *AssetsHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	// Check if any assets use this category
	var cnt int
	_ = h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM fixed_assets WHERE category_id = $1 AND company_id = $2`,
		id, companyID,
	).Scan(&cnt)
	if cnt > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Cannot delete: %d assets are assigned to this category", cnt)})
		return
	}

	_, err := h.db.Exec(ctx,
		`DELETE FROM asset_categories WHERE id = $1 AND company_id = $2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}

// ── Asset Locations ───────────────────────────────────────────────────────────

func (h *AssetsHandler) ListLocations(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT al.id, al.name, COALESCE(al.code,''), COALESCE(al.description,''),
		       COALESCE(al.address,''), COALESCE(al.city,''), COALESCE(al.country,''),
		       COALESCE(al.parent_id::text,''), COALESCE(p.name,''),
		       al.is_active,
		       COUNT(fa.id) AS asset_count,
		       COALESCE(SUM(fa.net_book_value),0),
		       al.created_at
		FROM asset_locations al
		LEFT JOIN asset_locations p ON p.id = al.parent_id
		LEFT JOIN fixed_assets fa ON fa.location_id = al.id AND fa.company_id = $1
		WHERE al.company_id = $1
		GROUP BY al.id, al.name, al.code, al.description, al.address,
		         al.city, al.country, al.parent_id, p.name, al.is_active, al.created_at
		ORDER BY al.name
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var id, name, code, desc, address, city, country string
		var parentID, parentName string
		var isActive bool
		var assetCount int
		var totalNBV float64
		var createdAt interface{}
		_ = rows.Scan(&id, &name, &code, &desc, &address, &city, &country,
			&parentID, &parentName, &isActive, &assetCount, &totalNBV, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "name": name, "code": code, "description": desc,
			"address": address, "city": city, "country": country,
			"parent_id": parentID, "parent_name": parentName,
			"is_active": isActive,
			"asset_count": assetCount, "total_net_book_value": totalNBV,
			"created_at": createdAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *AssetsHandler) CreateLocation(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := uuid.NewString()

	_, err := h.db.Exec(ctx, `
		INSERT INTO asset_locations (
			id, company_id, name, code, description,
			address, city, country, parent_id, is_active
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`, id, companyID,
		strVal(req, "name"), strVal(req, "code"), strVal(req, "description"),
		strVal(req, "address"), strVal(req, "city"), strVal(req, "country"),
		nullUUIDAsset(strVal(req, "parent_id")),
		assetBoolValDefault(req, "is_active", true),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req["id"] = id
	c.JSON(http.StatusCreated, req)
}

func (h *AssetsHandler) UpdateLocation(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.Exec(ctx, `
		UPDATE asset_locations SET
			name = $1, code = $2, description = $3,
			address = $4, city = $5, country = $6,
			parent_id = $7, is_active = $8, updated_at = NOW()
		WHERE id = $9 AND company_id = $10
	`,
		strVal(req, "name"), strVal(req, "code"), strVal(req, "description"),
		strVal(req, "address"), strVal(req, "city"), strVal(req, "country"),
		nullUUIDAsset(strVal(req, "parent_id")),
		assetBoolValDefault(req, "is_active", true),
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req["id"] = id
	c.JSON(http.StatusOK, req)
}

func (h *AssetsHandler) DeleteLocation(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var cnt int
	_ = h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM fixed_assets WHERE location_id = $1 AND company_id = $2`,
		id, companyID,
	).Scan(&cnt)
	if cnt > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Cannot delete: %d assets are at this location", cnt)})
		return
	}

	_, err := h.db.Exec(ctx,
		`DELETE FROM asset_locations WHERE id = $1 AND company_id = $2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Location deleted"})
}

// ── Asset Transfers ───────────────────────────────────────────────────────────

func (h *AssetsHandler) ListTransfers(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	where := []string{"at2.company_id = $1"}
	args := []interface{}{companyID}
	n := 2

	if s := c.Query("status"); s != "" {
		where = append(where, fmt.Sprintf("at2.status = $%d::transfer_status", n))
		args = append(args, s)
		n++
	}
	if assetID := c.Query("asset_id"); assetID != "" {
		where = append(where, fmt.Sprintf("at2.asset_id = $%d", n))
		args = append(args, assetID)
		n++
	}
	_ = n

	rows, err := h.db.Query(ctx, `
		SELECT at2.id, at2.transfer_number, at2.asset_id,
		       fa.asset_number, fa.name,
		       COALESCE(at2.from_location_id::text,''), COALESCE(fl.name,''),
		       COALESCE(at2.to_location_id::text,''), COALESCE(tl.name,''),
		       at2.status::text, at2.transfer_date,
		       COALESCE(at2.reason,''), COALESCE(at2.notes,''),
		       COALESCE(at2.approved_by::text,''), COALESCE(at2.approved_at,''),
		       at2.created_at
		FROM asset_transfers at2
		JOIN fixed_assets fa ON fa.id = at2.asset_id
		LEFT JOIN asset_locations fl ON fl.id = at2.from_location_id
		LEFT JOIN asset_locations tl ON tl.id = at2.to_location_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY at2.created_at DESC
	`, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var id, num, assetID, assetNum, assetName string
		var fromLocID, fromLocName, toLocID, toLocName string
		var status, reason, notes string
		var transferDate, approvedBy, approvedAt, createdAt interface{}
		_ = rows.Scan(&id, &num, &assetID, &assetNum, &assetName,
			&fromLocID, &fromLocName, &toLocID, &toLocName,
			&status, &transferDate, &reason, &notes,
			&approvedBy, &approvedAt, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "transfer_number": num, "asset_id": assetID,
			"asset_number": assetNum, "asset_name": assetName,
			"from_location_id": fromLocID, "from_location_name": fromLocName,
			"to_location_id": toLocID, "to_location_name": toLocName,
			"status": status, "transfer_date": transferDate,
			"reason": reason, "notes": notes,
			"approved_by": approvedBy, "approved_at": approvedAt,
			"created_at": createdAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *AssetsHandler) CreateTransfer(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var seq int64
	_ = h.db.QueryRow(ctx, `SELECT nextval('asset_transfer_seq')`).Scan(&seq)
	transferNumber := fmt.Sprintf("TRF-%04d", seq)

	id := uuid.NewString()
	transferDate := strVal(req, "transfer_date")
	if transferDate == "" {
		transferDate = time.Now().Format("2006-01-02")
	}

	_, err := h.db.Exec(ctx, `
		INSERT INTO asset_transfers (
			id, company_id, transfer_number, asset_id,
			from_location_id, to_location_id,
			status, transfer_date, reason, notes
		) VALUES ($1,$2,$3,$4,$5,$6,'pending'::transfer_status,$7,$8,$9)
	`, id, companyID, transferNumber,
		strVal(req, "asset_id"),
		nullUUIDAsset(strVal(req, "from_location_id")),
		strVal(req, "to_location_id"),
		transferDate,
		strVal(req, "reason"), strVal(req, "notes"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req["id"] = id
	req["transfer_number"] = transferNumber
	req["status"] = "pending"
	c.JSON(http.StatusCreated, req)
}

func (h *AssetsHandler) ApproveTransfer(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		UPDATE asset_transfers SET
			status = 'approved'::transfer_status,
			approved_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND company_id = $2 AND status = 'pending'
	`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transfer approved", "id": id})
}

func (h *AssetsHandler) CompleteTransfer(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	// Get the transfer to find the asset + new location
	var assetID, toLocID string
	err := h.db.QueryRow(ctx, `
		SELECT asset_id, COALESCE(to_location_id::text,'')
		FROM asset_transfers
		WHERE id = $1 AND company_id = $2
	`, id, companyID).Scan(&assetID, &toLocID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transfer not found"})
		return
	}

	// Update transfer status
	_, err = h.db.Exec(ctx, `
		UPDATE asset_transfers SET
			status = 'completed'::transfer_status,
			updated_at = NOW()
		WHERE id = $1 AND company_id = $2
	`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Move asset to new location
	if toLocID != "" {
		_, _ = h.db.Exec(ctx, `
			UPDATE fixed_assets SET location_id = $1, updated_at = NOW()
			WHERE id = $2 AND company_id = $3
		`, toLocID, assetID, companyID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer completed", "id": id})
}

func (h *AssetsHandler) DeleteTransfer(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	_, err := h.db.Exec(ctx,
		`DELETE FROM asset_transfers WHERE id = $1 AND company_id = $2 AND status = 'pending'`,
		id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transfer deleted"})
}

// ── Asset Depreciation ────────────────────────────────────────────────────────

func (h *AssetsHandler) ListDepreciation(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	where := []string{"fa.company_id = $1"}
	args := []interface{}{companyID}
	n := 2

	if assetID := c.Query("asset_id"); assetID != "" {
		where = append(where, fmt.Sprintf("ads.asset_id = $%d", n))
		args = append(args, assetID)
		n++
	}
	if yr := c.Query("year"); yr != "" {
		where = append(where, fmt.Sprintf("ads.period_year = $%d", n))
		args = append(args, yr)
		n++
	}
	if posted := c.Query("posted"); posted != "" {
		b := posted == "true" || posted == "1"
		where = append(where, fmt.Sprintf("ads.is_posted = $%d", n))
		args = append(args, b)
		n++
	}
	_ = n

	rows, err := h.db.Query(ctx, `
		SELECT ads.id, ads.asset_id, fa.asset_number, fa.name,
		       ads.period_year, ads.period_month,
		       ads.opening_book_value, ads.depreciation_amount,
		       ads.accumulated_depreciation, ads.closing_book_value,
		       ads.is_posted, ads.posted_at
		FROM asset_depreciation_schedules ads
		JOIN fixed_assets fa ON fa.id = ads.asset_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY ads.period_year DESC, ads.period_month DESC, fa.asset_number
	`, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var id, assetID, assetNum, assetName string
		var yr, mo int
		var open, dep, accum, close float64
		var isPosted bool
		var postedAt interface{}
		_ = rows.Scan(&id, &assetID, &assetNum, &assetName,
			&yr, &mo, &open, &dep, &accum, &close, &isPosted, &postedAt)
		list = append(list, map[string]interface{}{
			"id": id, "asset_id": assetID,
			"asset_number": assetNum, "asset_name": assetName,
			"period_year": yr, "period_month": mo,
			"opening_book_value": open, "depreciation_amount": dep,
			"accumulated_depreciation": accum, "closing_book_value": close,
			"is_posted": isPosted, "posted_at": postedAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *AssetsHandler) GenerateDepreciation(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req struct {
		Year  int `json:"year"`
		Month int `json:"month"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Year == 0 {
		req.Year = time.Now().Year()
	}
	if req.Month == 0 {
		req.Month = int(time.Now().Month())
	}

	// Fetch all depreciable, active assets
	rows, err := h.db.Query(ctx, `
		SELECT id, purchase_cost, salvage_value, useful_life_years,
		       depreciation_method::text, depreciation_rate, accumulated_depreciation,
		       net_book_value
		FROM fixed_assets
		WHERE company_id = $1
		  AND is_depreciable = true
		  AND status IN ('active','in_use','in_storage')
		  AND net_book_value > 0
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	generated := 0
	skipped := 0

	for rows.Next() {
		var assetID string
		var cost, salvage, useful, depRate, accumDep, nbv float64
		var depMethod string
		_ = rows.Scan(&assetID, &cost, &salvage, &useful, &depMethod, &depRate, &accumDep, &nbv)

		// Skip if schedule already exists for this period
		var existing int
		_ = h.db.QueryRow(ctx, `
			SELECT COUNT(*) FROM asset_depreciation_schedules
			WHERE asset_id = $1 AND period_year = $2 AND period_month = $3
		`, assetID, req.Year, req.Month).Scan(&existing)
		if existing > 0 {
			skipped++
			continue
		}

		// Calculate monthly depreciation
		var monthlyDep float64
		switch depMethod {
		case "straight_line":
			if useful > 0 {
				annual := (cost - salvage) / useful
				monthlyDep = annual / 12
			}
		case "declining_balance":
			monthlyDep = (nbv * depRate / 100) / 12
		case "double_declining":
			ddRate := 0.0
			if useful > 0 {
				ddRate = 2.0 / useful
			}
			monthlyDep = nbv * ddRate / 12
		case "sum_of_years":
			// SYD = (remaining_life / SYD_total) * (cost - salvage) / 12
			remainingLife := math.Max(0, useful-(accumDep/((cost-salvage)/useful)))
			syd := useful * (useful + 1) / 2
			if syd > 0 && cost != salvage {
				monthlyDep = (remainingLife / syd) * (cost - salvage) / 12
			}
		default:
			if useful > 0 {
				monthlyDep = (cost - salvage) / useful / 12
			}
		}

		// Cap at remaining depreciable amount
		remaining := nbv - salvage
		if monthlyDep > remaining {
			monthlyDep = remaining
		}
		if monthlyDep < 0 {
			monthlyDep = 0
		}

		// Round to 4 decimal places
		monthlyDep = math.Round(monthlyDep*10000) / 10000

		openNBV := nbv
		newAccum := accumDep + monthlyDep
		closeNBV := cost - newAccum

		schedID := uuid.NewString()
		_, err := h.db.Exec(ctx, `
			INSERT INTO asset_depreciation_schedules (
				id, asset_id, period_year, period_month,
				opening_book_value, depreciation_amount,
				accumulated_depreciation, closing_book_value, is_posted
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,false)
			ON CONFLICT (asset_id, period_year, period_month) DO NOTHING
		`, schedID, assetID, req.Year, req.Month,
			openNBV, monthlyDep, newAccum, closeNBV)
		if err == nil {
			generated++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Depreciation generated for %d/%d", req.Month, req.Year),
		"generated": generated,
		"skipped":   skipped,
	})
}

func (h *AssetsHandler) PostDepreciation(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req struct {
		Year  int      `json:"year"`
		Month int      `json:"month"`
		IDs   []string `json:"ids"` // optional: specific schedule IDs
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var rows interface{ Close() }
	var posted int

	if len(req.IDs) > 0 {
		// Post specific schedules
		for _, schedID := range req.IDs {
			var assetID string
			var depAmt, newAccum float64
			err := h.db.QueryRow(ctx, `
				SELECT ads.asset_id, ads.depreciation_amount, ads.accumulated_depreciation
				FROM asset_depreciation_schedules ads
				JOIN fixed_assets fa ON fa.id = ads.asset_id
				WHERE ads.id = $1 AND fa.company_id = $2 AND ads.is_posted = false
			`, schedID, companyID).Scan(&assetID, &depAmt, &newAccum)
			if err != nil {
				continue
			}
			_, err = h.db.Exec(ctx, `
				UPDATE asset_depreciation_schedules
				SET is_posted = true, posted_at = NOW()
				WHERE id = $1
			`, schedID)
			if err == nil {
				// Update asset's accumulated depreciation
				_, _ = h.db.Exec(ctx, `
					UPDATE fixed_assets SET
						accumulated_depreciation = $1,
						updated_at = NOW()
					WHERE id = $2 AND company_id = $3
				`, newAccum, assetID, companyID)
				posted++
			}
		}
	} else {
		// Post all unposted for the given month/year
		schedRows, err := h.db.Query(ctx, `
			SELECT ads.id, ads.asset_id, ads.depreciation_amount, ads.accumulated_depreciation
			FROM asset_depreciation_schedules ads
			JOIN fixed_assets fa ON fa.id = ads.asset_id
			WHERE fa.company_id = $1
			  AND ads.period_year = $2
			  AND ads.period_month = $3
			  AND ads.is_posted = false
		`, companyID, req.Year, req.Month)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rows = schedRows
		defer schedRows.Close()
		for schedRows.Next() {
			var schedID, assetID string
			var depAmt, newAccum float64
			_ = schedRows.Scan(&schedID, &assetID, &depAmt, &newAccum)
			_, err = h.db.Exec(ctx, `
				UPDATE asset_depreciation_schedules
				SET is_posted = true, posted_at = NOW()
				WHERE id = $1
			`, schedID)
			if err == nil {
				_, _ = h.db.Exec(ctx, `
					UPDATE fixed_assets SET
						accumulated_depreciation = $1,
						updated_at = NOW()
					WHERE id = $2 AND company_id = $3
				`, newAccum, assetID, companyID)
				posted++
			}
		}
	}
	_ = rows

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Posted %d depreciation entries", posted),
		"posted":  posted,
	})
}

// ── Asset Maintenance ─────────────────────────────────────────────────────────

func (h *AssetsHandler) ListMaintenance(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	where := []string{"amr.company_id = $1"}
	args := []interface{}{companyID}
	n := 2

	if s := c.Query("status"); s != "" {
		where = append(where, fmt.Sprintf("amr.status = $%d::maintenance_status", n))
		args = append(args, s)
		n++
	}
	if t := c.Query("type"); t != "" {
		where = append(where, fmt.Sprintf("amr.maintenance_type = $%d::maintenance_type", n))
		args = append(args, t)
		n++
	}
	if assetID := c.Query("asset_id"); assetID != "" {
		where = append(where, fmt.Sprintf("amr.asset_id = $%d", n))
		args = append(args, assetID)
		n++
	}
	_ = n

	rows, err := h.db.Query(ctx, `
		SELECT amr.id, amr.asset_id, fa.asset_number, fa.name,
		       amr.maintenance_type::text, amr.status::text,
		       amr.scheduled_date, amr.completed_date,
		       COALESCE(amr.description,''), COALESCE(amr.technician,''),
		       COALESCE(amr.vendor,''), amr.cost,
		       COALESCE(amr.parts_used,''), COALESCE(amr.notes,''),
		       COALESCE(amr.next_maintenance_date,''),
		       amr.created_at
		FROM asset_maintenance_records amr
		JOIN fixed_assets fa ON fa.id = amr.asset_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY amr.scheduled_date DESC, amr.created_at DESC
	`, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var id, assetID, assetNum, assetName string
		var mType, status, desc, technician, vendor, partsUsed, notes string
		var cost float64
		var scheduledDate, completedDate, nextMainDate, createdAt interface{}
		_ = rows.Scan(&id, &assetID, &assetNum, &assetName,
			&mType, &status,
			&scheduledDate, &completedDate,
			&desc, &technician, &vendor, &cost,
			&partsUsed, &notes, &nextMainDate,
			&createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "asset_id": assetID,
			"asset_number": assetNum, "asset_name": assetName,
			"maintenance_type": mType, "status": status,
			"scheduled_date": scheduledDate, "completed_date": completedDate,
			"description": desc, "technician": technician, "vendor": vendor,
			"cost": cost, "parts_used": partsUsed, "notes": notes,
			"next_maintenance_date": nextMainDate,
			"created_at": createdAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *AssetsHandler) CreateMaintenance(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.NewString()
	mType := strValDefault(req, "maintenance_type", "preventive")
	status := strValDefault(req, "status", "scheduled")
	scheduledDate := strVal(req, "scheduled_date")
	if scheduledDate == "" {
		scheduledDate = time.Now().Format("2006-01-02")
	}

	_, err := h.db.Exec(ctx, `
		INSERT INTO asset_maintenance_records (
			id, company_id, asset_id, maintenance_type, status,
			scheduled_date, completed_date,
			description, technician, vendor, cost,
			parts_used, notes, next_maintenance_date
		) VALUES (
			$1,$2,$3,$4::maintenance_type,$5::maintenance_status,
			$6,$7,$8,$9,$10,$11,$12,$13,$14
		)
	`, id, companyID,
		strVal(req, "asset_id"),
		mType, status,
		scheduledDate,
		assetNullableStr(strVal(req, "completed_date")),
		strVal(req, "description"),
		strVal(req, "technician"), strVal(req, "vendor"),
		assetFloatVal(req, "cost"),
		strVal(req, "parts_used"), strVal(req, "notes"),
		assetNullableStr(strVal(req, "next_maintenance_date")),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req["id"] = id
	c.JSON(http.StatusCreated, req)
}

func (h *AssetsHandler) UpdateMaintenance(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mType := strValDefault(req, "maintenance_type", "preventive")
	status := strValDefault(req, "status", "scheduled")

	_, err := h.db.Exec(ctx, `
		UPDATE asset_maintenance_records SET
			maintenance_type = $1::maintenance_type,
			status = $2::maintenance_status,
			scheduled_date = $3,
			completed_date = $4,
			description = $5,
			technician = $6, vendor = $7, cost = $8,
			parts_used = $9, notes = $10,
			next_maintenance_date = $11,
			updated_at = NOW()
		WHERE id = $12 AND company_id = $13
	`,
		mType, status,
		strVal(req, "scheduled_date"),
		assetNullableStr(strVal(req, "completed_date")),
		strVal(req, "description"),
		strVal(req, "technician"), strVal(req, "vendor"),
		assetFloatVal(req, "cost"),
		strVal(req, "parts_used"), strVal(req, "notes"),
		assetNullableStr(strVal(req, "next_maintenance_date")),
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req["id"] = id
	c.JSON(http.StatusOK, req)
}

func (h *AssetsHandler) CompleteMaintenance(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req struct {
		CompletedDate      string  `json:"completed_date"`
		Cost               float64 `json:"cost"`
		Notes              string  `json:"notes"`
		NextMaintenanceDate string `json:"next_maintenance_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.CompletedDate == "" {
		req.CompletedDate = time.Now().Format("2006-01-02")
	}

	_, err := h.db.Exec(ctx, `
		UPDATE asset_maintenance_records SET
			status = 'completed'::maintenance_status,
			completed_date = $1,
			cost = $2, notes = $3,
			next_maintenance_date = $4,
			updated_at = NOW()
		WHERE id = $5 AND company_id = $6
	`,
		req.CompletedDate, req.Cost, req.Notes,
		assetNullableStr(req.NextMaintenanceDate),
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Maintenance completed", "id": id})
}

func (h *AssetsHandler) DeleteMaintenance(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	_, err := h.db.Exec(ctx,
		`DELETE FROM asset_maintenance_records WHERE id = $1 AND company_id = $2`,
		id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Maintenance record deleted"})
}

// ── Reports ───────────────────────────────────────────────────────────────────

func (h *AssetsHandler) GetReports(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	reportType := c.Query("type")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	if dateFrom == "" {
		dateFrom = time.Now().AddDate(0, -12, 0).Format("2006-01-02")
	}
	if dateTo == "" {
		dateTo = time.Now().Format("2006-01-02")
	}

	switch reportType {

	case "asset_register":
		rows, err := h.db.Query(ctx, `
			SELECT fa.asset_number, fa.name,
			       COALESCE(ac.name,''), COALESCE(al.name,''),
			       fa.status::text, fa.condition::text,
			       fa.purchase_date, fa.purchase_cost,
			       fa.accumulated_depreciation, fa.net_book_value,
			       fa.depreciation_method::text, fa.useful_life_years,
			       COALESCE(fa.serial_number,''), COALESCE(fa.manufacturer,'')
			FROM fixed_assets fa
			LEFT JOIN asset_categories ac ON ac.id = fa.category_id
			LEFT JOIN asset_locations al ON al.id = fa.location_id
			WHERE fa.company_id = $1
			ORDER BY fa.asset_number
		`, companyID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var list []map[string]interface{}
		for rows.Next() {
			var assetNum, name, cat, loc, status, condition, depMethod, serial, manufacturer string
			var purchaseDate interface{}
			var cost, accumDep, nbv, useful float64
			_ = rows.Scan(&assetNum, &name, &cat, &loc, &status, &condition,
				&purchaseDate, &cost, &accumDep, &nbv, &depMethod, &useful, &serial, &manufacturer)
			list = append(list, map[string]interface{}{
				"asset_number": assetNum, "name": name,
				"category": cat, "location": loc,
				"status": status, "condition": condition,
				"purchase_date": purchaseDate, "purchase_cost": cost,
				"accumulated_depreciation": accumDep, "net_book_value": nbv,
				"depreciation_method": depMethod, "useful_life_years": useful,
				"serial_number": serial, "manufacturer": manufacturer,
			})
		}
		if list == nil {
			list = []map[string]interface{}{}
		}
		c.JSON(http.StatusOK, gin.H{"type": "asset_register", "data": list})

	case "depreciation_summary":
		yr := c.Query("year")
		if yr == "" {
			yr = strconv.Itoa(time.Now().Year())
		}
		yearInt, _ := strconv.Atoi(yr)
		rows, err := h.db.Query(ctx, `
			SELECT COALESCE(ac.name,'Uncategorized'),
			       COUNT(DISTINCT fa.id),
			       COALESCE(SUM(fa.purchase_cost),0),
			       COALESCE(SUM(fa.accumulated_depreciation),0),
			       COALESCE(SUM(fa.net_book_value),0),
			       COALESCE(SUM(ads.depreciation_amount),0) AS year_dep
			FROM fixed_assets fa
			LEFT JOIN asset_categories ac ON ac.id = fa.category_id
			LEFT JOIN asset_depreciation_schedules ads
			       ON ads.asset_id = fa.id AND ads.period_year = $2
			WHERE fa.company_id = $1
			GROUP BY ac.name ORDER BY ac.name
		`, companyID, yearInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var list []map[string]interface{}
		for rows.Next() {
			var cat string
			var cnt int
			var cost, accumDep, nbv, yearDep float64
			_ = rows.Scan(&cat, &cnt, &cost, &accumDep, &nbv, &yearDep)
			list = append(list, map[string]interface{}{
				"category": cat, "asset_count": cnt,
				"purchase_cost": cost, "accumulated_depreciation": accumDep,
				"net_book_value": nbv, "year_depreciation": yearDep,
			})
		}
		if list == nil {
			list = []map[string]interface{}{}
		}
		c.JSON(http.StatusOK, gin.H{"type": "depreciation_summary", "year": yearInt, "data": list})

	case "maintenance_costs":
		rows, err := h.db.Query(ctx, `
			SELECT fa.asset_number, fa.name,
			       COALESCE(ac.name,''),
			       COUNT(amr.id),
			       COALESCE(SUM(amr.cost),0),
			       COUNT(CASE WHEN amr.status = 'completed' THEN 1 END),
			       COUNT(CASE WHEN amr.status NOT IN ('completed','cancelled') THEN 1 END)
			FROM fixed_assets fa
			LEFT JOIN asset_categories ac ON ac.id = fa.category_id
			LEFT JOIN asset_maintenance_records amr ON amr.asset_id = fa.id
			    AND amr.scheduled_date BETWEEN $2 AND $3
			WHERE fa.company_id = $1
			GROUP BY fa.asset_number, fa.name, ac.name
			HAVING COUNT(amr.id) > 0
			ORDER BY COALESCE(SUM(amr.cost),0) DESC
		`, companyID, dateFrom, dateTo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var list []map[string]interface{}
		for rows.Next() {
			var assetNum, name, cat string
			var total, cost float64
			var completed, pending int
			_ = rows.Scan(&assetNum, &name, &cat, &total, &cost, &completed, &pending)
			list = append(list, map[string]interface{}{
				"asset_number": assetNum, "name": name, "category": cat,
				"total_records": total, "total_cost": cost,
				"completed": completed, "pending": pending,
			})
		}
		if list == nil {
			list = []map[string]interface{}{}
		}
		c.JSON(http.StatusOK, gin.H{"type": "maintenance_costs", "data": list, "date_from": dateFrom, "date_to": dateTo})

	case "transfer_history":
		rows, err := h.db.Query(ctx, `
			SELECT at2.transfer_number, fa.asset_number, fa.name,
			       COALESCE(fl.name,''), COALESCE(tl.name,''),
			       at2.status::text, at2.transfer_date,
			       COALESCE(at2.reason,'')
			FROM asset_transfers at2
			JOIN fixed_assets fa ON fa.id = at2.asset_id
			LEFT JOIN asset_locations fl ON fl.id = at2.from_location_id
			LEFT JOIN asset_locations tl ON tl.id = at2.to_location_id
			WHERE at2.company_id = $1
			  AND at2.transfer_date BETWEEN $2 AND $3
			ORDER BY at2.transfer_date DESC
		`, companyID, dateFrom, dateTo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var list []map[string]interface{}
		for rows.Next() {
			var tNum, assetNum, name, from, to, status, reason string
			var tDate interface{}
			_ = rows.Scan(&tNum, &assetNum, &name, &from, &to, &status, &tDate, &reason)
			list = append(list, map[string]interface{}{
				"transfer_number": tNum, "asset_number": assetNum, "asset_name": name,
				"from_location": from, "to_location": to,
				"status": status, "transfer_date": tDate, "reason": reason,
			})
		}
		if list == nil {
			list = []map[string]interface{}{}
		}
		c.JSON(http.StatusOK, gin.H{"type": "transfer_history", "data": list})

	default:
		// Summary report
		var totalAssets, activeAssets, disposedAssets int
		var totalCost, totalNBV, totalAccumDep float64
		_ = h.db.QueryRow(ctx, `
			SELECT
				COUNT(*),
				COUNT(CASE WHEN status IN ('active','in_use') THEN 1 END),
				COUNT(CASE WHEN status IN ('disposed','sold','written_off') THEN 1 END),
				COALESCE(SUM(purchase_cost),0),
				COALESCE(SUM(net_book_value),0),
				COALESCE(SUM(accumulated_depreciation),0)
			FROM fixed_assets WHERE company_id = $1
		`, companyID).Scan(
			&totalAssets, &activeAssets, &disposedAssets,
			&totalCost, &totalNBV, &totalAccumDep,
		)

		// Monthly depreciation trend (last 12 months)
		trendRows, _ := h.db.Query(ctx, `
			SELECT ads.period_year, ads.period_month, COALESCE(SUM(ads.depreciation_amount),0)
			FROM asset_depreciation_schedules ads
			JOIN fixed_assets fa ON fa.id = ads.asset_id
			WHERE fa.company_id = $1
			  AND (ads.period_year * 100 + ads.period_month) >= $2
			GROUP BY ads.period_year, ads.period_month
			ORDER BY ads.period_year, ads.period_month
		`, companyID,
			(time.Now().Year()-1)*100+int(time.Now().Month()),
		)
		var trend []map[string]interface{}
		if trendRows != nil {
			defer trendRows.Close()
			for trendRows.Next() {
				var yr, mo int
				var dep float64
				_ = trendRows.Scan(&yr, &mo, &dep)
				trend = append(trend, map[string]interface{}{
					"year": yr, "month": mo, "depreciation": dep,
					"label": fmt.Sprintf("%d-%02d", yr, mo),
				})
			}
		}
		if trend == nil {
			trend = []map[string]interface{}{}
		}

		// Maintenance cost by month
		mainRows, _ := h.db.Query(ctx, `
			SELECT TO_CHAR(completed_date,'YYYY-MM'), COALESCE(SUM(cost),0), COUNT(*)
			FROM asset_maintenance_records
			WHERE company_id = $1
			  AND completed_date BETWEEN $2 AND $3
			  AND status = 'completed'
			GROUP BY TO_CHAR(completed_date,'YYYY-MM')
			ORDER BY 1
		`, companyID, dateFrom, dateTo)
		var mainCosts []map[string]interface{}
		if mainRows != nil {
			defer mainRows.Close()
			for mainRows.Next() {
				var label string
				var cost float64
				var cnt int
				_ = mainRows.Scan(&label, &cost, &cnt)
				mainCosts = append(mainCosts, map[string]interface{}{
					"period": label, "cost": cost, "count": cnt,
				})
			}
		}
		if mainCosts == nil {
			mainCosts = []map[string]interface{}{}
		}

		c.JSON(http.StatusOK, gin.H{
			"type":             "summary",
			"total_assets":     totalAssets,
			"active_assets":    activeAssets,
			"disposed_assets":  disposedAssets,
			"total_cost":       totalCost,
			"total_nbv":        totalNBV,
			"total_accum_dep":  totalAccumDep,
			"depreciation_trend": trend,
			"maintenance_costs":  mainCosts,
		})
	}
}

// ── Local helpers ─────────────────────────────────────────────────────────────

func assetFloatVal(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case float64:
			return val
		case float32:
			return float64(val)
		case int:
			return float64(val)
		case int64:
			return float64(val)
		case string:
			f, _ := strconv.ParseFloat(val, 64)
			return f
		}
	}
	return 0
}

func assetFloatValDefault(m map[string]interface{}, key string, def float64) float64 {
	v := assetFloatVal(m, key)
	if v == 0 {
		return def
	}
	return v
}

func assetBoolValDefault(m map[string]interface{}, key string, def bool) bool {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case bool:
			return val
		case string:
			return val == "true" || val == "1"
		case float64:
			return val != 0
		}
	}
	return def
}

func assetNullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
