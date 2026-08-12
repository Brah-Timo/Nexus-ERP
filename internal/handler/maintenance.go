package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"nexus-erp/internal/middleware"
)

// MaintenanceHandler handles all maintenance HTTP requests.
type MaintenanceHandler struct {
	db *pgxpool.Pool
}

// ── map helpers (m-prefixed to avoid conflicts with other handlers) ────────────

func mStr(m map[string]interface{}, k string) string {
	if v, ok := m[k]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func mStrN(m map[string]interface{}, k string) *string {
	s := mStr(m, k)
	if s == "" {
		return nil
	}
	return &s
}

func mStrD(m map[string]interface{}, k, def string) string {
	s := mStr(m, k)
	if s == "" {
		return def
	}
	return s
}

func mF64D(m map[string]interface{}, k string, def float64) float64 {
	if v, ok := m[k]; ok && v != nil {
		switch t := v.(type) {
		case float64:
			return t
		case string:
			if f, err := strconv.ParseFloat(t, 64); err == nil {
				return f
			}
		}
	}
	return def
}

func mIntD(m map[string]interface{}, k string, def int) int {
	if v, ok := m[k]; ok && v != nil {
		switch t := v.(type) {
		case float64:
			return int(t)
		case int:
			return t
		case string:
			if n, err := strconv.Atoi(t); err == nil {
				return n
			}
		}
	}
	return def
}

func mIntN(m map[string]interface{}, k string) *int {
	if v, ok := m[k]; ok && v != nil {
		switch t := v.(type) {
		case float64:
			n := int(t)
			return &n
		case int:
			return &t
		case string:
			if n, err := strconv.Atoi(t); err == nil {
				return &n
			}
		}
	}
	return nil
}

func mBoolD(m map[string]interface{}, k string, def bool) bool {
	if v, ok := m[k]; ok && v != nil {
		switch t := v.(type) {
		case bool:
			return t
		case string:
			return t == "true" || t == "1"
		}
	}
	return def
}

func mDateN(m map[string]interface{}, k string) *string {
	s := mStr(m, k)
	if s == "" {
		return nil
	}
	return &s
}

func mDateD(m map[string]interface{}, k string) string {
	s := mStr(m, k)
	if s == "" {
		return time.Now().Format("2006-01-02")
	}
	return s
}

func mNullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func mPad(n, w int) string {
	return fmt.Sprintf("%0*d", w, n)
}

func mColorType(t string) string {
	switch t {
	case "corrective":
		return "#ef4444"
	case "preventive":
		return "#22c55e"
	case "inspection":
		return "#3b82f6"
	case "emergency":
		return "#f97316"
	case "upgrade":
		return "#8b5cf6"
	default:
		return "#6366f1"
	}
}

// ── EQUIPMENT ─────────────────────────────────────────────────────────────────

// ListEquipment GET /maintenance/equipment
func (h *MaintenanceHandler) ListEquipment(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}

	search := c.Query("search")
	status := c.Query("status")
	category := c.Query("category")
	pg, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	lim, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if pg < 1 {
		pg = 1
	}
	if lim < 1 || lim > 200 {
		lim = 50
	}
	offset := (pg - 1) * lim

	args := []interface{}{cid}
	where := []string{"e.company_id=$1", "e.is_active=TRUE"}
	idx := 2

	if search != "" {
		where = append(where, fmt.Sprintf("(e.code ILIKE $%d OR e.name ILIKE $%d OR e.serial_number ILIKE $%d)", idx, idx, idx))
		args = append(args, "%"+search+"%")
		idx++
	}
	if status != "" {
		where = append(where, fmt.Sprintf("e.status=$%d", idx))
		args = append(args, status)
		idx++
	}
	if category != "" {
		where = append(where, fmt.Sprintf("e.category=$%d", idx))
		args = append(args, category)
		idx++
	}
	wc := strings.Join(where, " AND ")

	var total int
	_ = h.db.QueryRow(context.Background(),
		fmt.Sprintf("SELECT COUNT(*) FROM equipment e WHERE %s", wc), args...).Scan(&total)

	args = append(args, lim, offset)
	rows, err := h.db.Query(context.Background(), fmt.Sprintf(`
		SELECT e.id,e.code,e.name,e.category,e.subcategory,e.location,e.department,e.status,
		       e.purchase_date,e.purchase_cost,e.current_value,e.warranty_expiry,
		       e.manufacturer,e.model,e.serial_number,e.asset_tag,
		       e.last_maintenance_date,e.next_maintenance_date,
		       e.maintenance_interval_days,e.expected_life_years,
		       e.notes,e.is_active,e.created_at,
		       COALESCE((SELECT COUNT(*) FROM maintenance_orders mo WHERE mo.equipment_id=e.id AND mo.status NOT IN('completed','cancelled')),0) AS open_orders,
		       COALESCE((SELECT COUNT(*) FROM maintenance_requests mr WHERE mr.equipment_id=e.id AND mr.status NOT IN('completed','rejected','cancelled')),0) AS open_requests
		FROM equipment e
		WHERE %s ORDER BY e.name ASC LIMIT $%d OFFSET $%d
	`, wc, idx, idx+1), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	now := time.Now()
	items := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, code, name, st                          string
			cat, sub, loc, dept                         *string
			pd, we, lm, nm                              *string
			pc, cv                                      float64
			mfr, mdl, sn, at, notes                     *string
			mi, ely                                     *int
			isActive                                    bool
			createdAt                                   time.Time
			openOrders, openRequests                    int
		)
		if err := rows.Scan(&id, &code, &name, &cat, &sub, &loc, &dept, &st,
			&pd, &pc, &cv, &we, &mfr, &mdl, &sn, &at,
			&lm, &nm, &mi, &ely, &notes, &isActive, &createdAt,
			&openOrders, &openRequests); err != nil {
			continue
		}
		overdue := false
		if nm != nil && *nm != "" {
			if t, err := time.Parse("2006-01-02", *nm); err == nil {
				overdue = now.After(t) && st == "active"
			}
		}
		items = append(items, map[string]interface{}{
			"id": id, "code": code, "name": name, "category": cat, "subcategory": sub,
			"location": loc, "department": dept, "status": st,
			"purchase_date": pd, "purchase_cost": pc, "current_value": cv,
			"warranty_expiry": we, "manufacturer": mfr, "model": mdl,
			"serial_number": sn, "asset_tag": at,
			"last_maintenance_date": lm, "next_maintenance_date": nm,
			"maintenance_interval_days": mi, "expected_life_years": ely,
			"notes": notes, "is_active": isActive, "created_at": createdAt,
			"open_orders": openOrders, "open_requests": openRequests, "overdue": overdue,
		})
	}

	// status summary
	srows, _ := h.db.Query(context.Background(),
		`SELECT status,COUNT(*),COALESCE(SUM(current_value),0) FROM equipment WHERE company_id=$1 AND is_active=TRUE GROUP BY status`, cid)
	defer srows.Close()
	statusSum := []map[string]interface{}{}
	for srows.Next() {
		var s string
		var cnt int
		var val float64
		if err := srows.Scan(&s, &cnt, &val); err == nil {
			statusSum = append(statusSum, map[string]interface{}{"status": s, "count": cnt, "value": val})
		}
	}

	// category summary
	crows, _ := h.db.Query(context.Background(),
		`SELECT COALESCE(category,'Uncategorized'),COUNT(*) FROM equipment WHERE company_id=$1 AND is_active=TRUE GROUP BY category ORDER BY 2 DESC`, cid)
	defer crows.Close()
	catSum := []map[string]interface{}{}
	for crows.Next() {
		var cat string
		var cnt int
		if err := crows.Scan(&cat, &cnt); err == nil {
			catSum = append(catSum, map[string]interface{}{"category": cat, "count": cnt})
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": pg, "limit": lim,
		"status_summary": statusSum, "category_summary": catSum})
}

// GetEquipment GET /maintenance/equipment/:id
func (h *MaintenanceHandler) GetEquipment(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	var (
		code, name, st                              string
		cat, sub, loc, dept                         *string
		pd, we, lm, nm                              *string
		pc, cv                                      float64
		mfr, mdl, sn, at, notes, imgURL             *string
		mi, ely                                     *int
		specs                                       interface{}
		isActive                                    bool
		createdAt                                   time.Time
	)
	err := h.db.QueryRow(context.Background(), `
		SELECT code,name,category,subcategory,location,department,status,
		       purchase_date,purchase_cost,current_value,warranty_expiry,
		       manufacturer,model,serial_number,asset_tag,specifications,
		       last_maintenance_date,next_maintenance_date,
		       maintenance_interval_days,expected_life_years,
		       notes,image_url,is_active,created_at
		FROM equipment WHERE id=$1 AND company_id=$2 AND is_active=TRUE
	`, id, cid).Scan(&code, &name, &cat, &sub, &loc, &dept, &st,
		&pd, &pc, &cv, &we, &mfr, &mdl, &sn, &at, &specs,
		&lm, &nm, &mi, &ely, &notes, &imgURL, &isActive, &createdAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "equipment not found"})
		return
	}

	hrows, _ := h.db.Query(context.Background(),
		`SELECT id,title,history_type,performed_date,technician_name,total_cost
		 FROM maintenance_history WHERE equipment_id=$1 AND company_id=$2 AND is_active=TRUE
		 ORDER BY performed_date DESC LIMIT 10`, id, cid)
	defer hrows.Close()
	history := []map[string]interface{}{}
	for hrows.Next() {
		var hid, title, htype, tech string
		var perfDate *string
		var hcost float64
		if err := hrows.Scan(&hid, &title, &htype, &perfDate, &tech, &hcost); err == nil {
			history = append(history, map[string]interface{}{
				"id": hid, "title": title, "type": htype,
				"performed_date": perfDate, "technician": tech, "cost": hcost,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id, "code": code, "name": name, "category": cat, "subcategory": sub,
		"location": loc, "department": dept, "status": st,
		"purchase_date": pd, "purchase_cost": pc, "current_value": cv,
		"warranty_expiry": we, "manufacturer": mfr, "model": mdl,
		"serial_number": sn, "asset_tag": at, "specifications": specs,
		"last_maintenance_date": lm, "next_maintenance_date": nm,
		"maintenance_interval_days": mi, "expected_life_years": ely,
		"notes": notes, "image_url": imgURL, "is_active": isActive, "created_at": createdAt,
		"recent_history": history,
	})
}

// GetEquipmentCategories GET /maintenance/equipment/categories
func (h *MaintenanceHandler) GetEquipmentCategories(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	rows, err := h.db.Query(context.Background(),
		`SELECT COALESCE(category,'Uncategorized'),COUNT(*) FROM equipment
		 WHERE company_id=$1 AND is_active=TRUE GROUP BY category ORDER BY 2 DESC`, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	cats := []map[string]interface{}{}
	for rows.Next() {
		var cat string
		var cnt int
		if err := rows.Scan(&cat, &cnt); err == nil {
			cats = append(cats, map[string]interface{}{"category": cat, "count": cnt})
		}
	}
	c.JSON(http.StatusOK, gin.H{"categories": cats})
}

// CreateEquipment POST /maintenance/equipment
func (h *MaintenanceHandler) CreateEquipment(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := mStr(body, "name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	code := mStr(body, "code")
	if code == "" {
		var n int
		_ = h.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM equipment WHERE company_id=$1", cid).Scan(&n)
		code = "EQ-" + mPad(n+1, 4)
	}
	var id string
	err := h.db.QueryRow(context.Background(), `
		INSERT INTO equipment(company_id,code,name,category,subcategory,location,department,
		  status,purchase_date,purchase_cost,current_value,warranty_expiry,
		  manufacturer,model,serial_number,asset_tag,
		  last_maintenance_date,next_maintenance_date,
		  maintenance_interval_days,expected_life_years,notes,is_active)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		RETURNING id`,
		cid, code, name,
		mStrN(body, "category"), mStrN(body, "subcategory"),
		mStrN(body, "location"), mStrN(body, "department"),
		mStrD(body, "status", "active"),
		mDateN(body, "purchase_date"), mF64D(body, "purchase_cost", 0), mF64D(body, "current_value", 0),
		mDateN(body, "warranty_expiry"),
		mStrN(body, "manufacturer"), mStrN(body, "model"),
		mStrN(body, "serial_number"), mStrN(body, "asset_tag"),
		mDateN(body, "last_maintenance_date"), mDateN(body, "next_maintenance_date"),
		mIntD(body, "maintenance_interval_days", 90), mIntN(body, "expected_life_years"),
		mStrN(body, "notes"), mBoolD(body, "is_active", true),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "equipment created"})
}

// UpdateEquipment PUT /maintenance/equipment/:id
func (h *MaintenanceHandler) UpdateEquipment(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec(context.Background(), `
		UPDATE equipment SET
		  name=COALESCE($3,name), category=$4, subcategory=$5, location=$6, department=$7,
		  status=COALESCE($8,status),
		  purchase_date=$9, purchase_cost=$10, current_value=$11, warranty_expiry=$12,
		  manufacturer=$13, model=$14, serial_number=$15, asset_tag=$16,
		  last_maintenance_date=$17, next_maintenance_date=$18,
		  maintenance_interval_days=COALESCE($19,maintenance_interval_days),
		  expected_life_years=$20, notes=$21, is_active=$22, updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, cid,
		mStrN(body, "name"),
		mStrN(body, "category"), mStrN(body, "subcategory"),
		mStrN(body, "location"), mStrN(body, "department"),
		mStrN(body, "status"),
		mDateN(body, "purchase_date"), mF64D(body, "purchase_cost", 0), mF64D(body, "current_value", 0),
		mDateN(body, "warranty_expiry"),
		mStrN(body, "manufacturer"), mStrN(body, "model"),
		mStrN(body, "serial_number"), mStrN(body, "asset_tag"),
		mDateN(body, "last_maintenance_date"), mDateN(body, "next_maintenance_date"),
		mIntN(body, "maintenance_interval_days"), mIntN(body, "expected_life_years"),
		mStrN(body, "notes"), mBoolD(body, "is_active", true),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeleteEquipment DELETE /maintenance/equipment/:id
func (h *MaintenanceHandler) DeleteEquipment(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	_, err := h.db.Exec(context.Background(),
		"UPDATE equipment SET is_active=FALSE,updated_at=NOW() WHERE id=$1 AND company_id=$2",
		c.Param("id"), cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ── REQUESTS ──────────────────────────────────────────────────────────────────

// ListRequests GET /maintenance/requests
func (h *MaintenanceHandler) ListRequests(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	search := c.Query("search")
	status := c.Query("status")
	priority := c.Query("priority")
	pg, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	lim, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if pg < 1 {
		pg = 1
	}
	if lim < 1 || lim > 200 {
		lim = 50
	}
	offset := (pg - 1) * lim

	args := []interface{}{cid}
	where := []string{"mr.company_id=$1", "mr.is_active=TRUE"}
	idx := 2

	if search != "" {
		where = append(where, fmt.Sprintf("(mr.request_number ILIKE $%d OR mr.title ILIKE $%d)", idx, idx))
		args = append(args, "%"+search+"%")
		idx++
	}
	if status != "" {
		where = append(where, fmt.Sprintf("mr.status=$%d", idx))
		args = append(args, status)
		idx++
	}
	if priority != "" {
		where = append(where, fmt.Sprintf("mr.priority=$%d", idx))
		args = append(args, priority)
		idx++
	}
	wc := strings.Join(where, " AND ")

	var total int
	_ = h.db.QueryRow(context.Background(),
		fmt.Sprintf("SELECT COUNT(*) FROM maintenance_requests mr WHERE %s", wc), args...).Scan(&total)

	args = append(args, lim, offset)
	rows, err := h.db.Query(context.Background(), fmt.Sprintf(`
		SELECT mr.id,mr.request_number,mr.title,mr.description,mr.priority,mr.status,
		       mr.failure_type,mr.symptoms,mr.requested_by_name,mr.assigned_to_name,
		       mr.submitted_at,mr.approved_at,mr.completed_at,
		       mr.estimated_cost,mr.actual_cost,mr.notes,mr.created_at,
		       e.id,e.name,e.code,e.location
		FROM maintenance_requests mr
		LEFT JOIN equipment e ON e.id=mr.equipment_id
		WHERE %s ORDER BY mr.created_at DESC LIMIT $%d OFFSET $%d
	`, wc, idx, idx+1), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	items := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, rnum, title, pri, st                           string
			desc, ftype, sym, reqBy, assBy, notes              *string
			subAt, appAt, compAt                               *time.Time
			estCost, actCost                                    *float64
			createdAt                                          time.Time
			eqID, eqName, eqCode, eqLoc                        *string
		)
		if err := rows.Scan(&id, &rnum, &title, &desc, &pri, &st,
			&ftype, &sym, &reqBy, &assBy,
			&subAt, &appAt, &compAt, &estCost, &actCost, &notes, &createdAt,
			&eqID, &eqName, &eqCode, &eqLoc); err != nil {
			continue
		}
		items = append(items, map[string]interface{}{
			"id": id, "request_number": rnum, "title": title, "description": desc,
			"priority": pri, "status": st, "failure_type": ftype, "symptoms": sym,
			"requested_by_name": reqBy, "assigned_to_name": assBy,
			"submitted_at": subAt, "approved_at": appAt, "completed_at": compAt,
			"estimated_cost": estCost, "actual_cost": actCost, "notes": notes,
			"created_at": createdAt,
			"equipment_id": eqID, "equipment_name": eqName,
			"equipment_code": eqCode, "equipment_location": eqLoc,
		})
	}

	srows, _ := h.db.Query(context.Background(),
		`SELECT status,COUNT(*) FROM maintenance_requests WHERE company_id=$1 AND is_active=TRUE GROUP BY status`, cid)
	defer srows.Close()
	summary := []map[string]interface{}{}
	for srows.Next() {
		var s string
		var cnt int
		if err := srows.Scan(&s, &cnt); err == nil {
			summary = append(summary, map[string]interface{}{"status": s, "count": cnt})
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": pg, "limit": lim, "summary": summary})
}

// GetRequest GET /maintenance/requests/:id
func (h *MaintenanceHandler) GetRequest(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	var (
		rnum, title, pri, st                            string
		desc, ftype, sym, reqBy, assBy, notes            *string
		subAt, appAt, compAt                             *time.Time
		estCost, actCost                                  *float64
		createdAt                                        time.Time
		eqID, eqName, eqCode                             *string
	)
	err := h.db.QueryRow(context.Background(), `
		SELECT mr.request_number,mr.title,mr.priority,mr.status,
		       mr.description,mr.failure_type,mr.symptoms,
		       mr.requested_by_name,mr.assigned_to_name,
		       mr.submitted_at,mr.approved_at,mr.completed_at,
		       mr.estimated_cost,mr.actual_cost,mr.notes,mr.created_at,
		       e.id,e.name,e.code
		FROM maintenance_requests mr
		LEFT JOIN equipment e ON e.id=mr.equipment_id
		WHERE mr.id=$1 AND mr.company_id=$2 AND mr.is_active=TRUE
	`, id, cid).Scan(&rnum, &title, &pri, &st,
		&desc, &ftype, &sym, &reqBy, &assBy,
		&subAt, &appAt, &compAt, &estCost, &actCost, &notes, &createdAt,
		&eqID, &eqName, &eqCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": id, "request_number": rnum, "title": title, "priority": pri, "status": st,
		"description": desc, "failure_type": ftype, "symptoms": sym,
		"requested_by_name": reqBy, "assigned_to_name": assBy,
		"submitted_at": subAt, "approved_at": appAt, "completed_at": compAt,
		"estimated_cost": estCost, "actual_cost": actCost, "notes": notes,
		"created_at": createdAt,
		"equipment_id": eqID, "equipment_name": eqName, "equipment_code": eqCode,
	})
}

// CreateRequest POST /maintenance/requests
func (h *MaintenanceHandler) CreateRequest(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	title := mStr(body, "title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title required"})
		return
	}
	year := time.Now().Year()
	var n int
	_ = h.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM maintenance_requests WHERE company_id=$1 AND EXTRACT(YEAR FROM created_at)=$2",
		cid, year).Scan(&n)
	rnum := fmt.Sprintf("MR-%d-%s", year, mPad(n+1, 4))

	var id string
	err := h.db.QueryRow(context.Background(), `
		INSERT INTO maintenance_requests(company_id,equipment_id,request_number,title,description,
		  priority,status,failure_type,symptoms,requested_by_name,assigned_to_name,estimated_cost,notes)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING id`,
		cid, mNullStr(mStr(body, "equipment_id")), rnum, title,
		mStrN(body, "description"),
		mStrD(body, "priority", "medium"), mStrD(body, "status", "draft"),
		mStrN(body, "failure_type"), mStrN(body, "symptoms"),
		mStrN(body, "requested_by_name"), mStrN(body, "assigned_to_name"),
		mIntN(body, "estimated_cost"), mStrN(body, "notes"),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "request_number": rnum, "message": "created"})
}

// UpdateRequest PUT /maintenance/requests/:id
func (h *MaintenanceHandler) UpdateRequest(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ns := mStr(body, "status")
	var subAt, appAt, compAt interface{}
	if ns == "submitted" {
		subAt = time.Now()
	}
	if ns == "approved" {
		appAt = time.Now()
	}
	if ns == "completed" {
		compAt = time.Now()
	}
	_, err := h.db.Exec(context.Background(), `
		UPDATE maintenance_requests SET
		  equipment_id=COALESCE($3::uuid,equipment_id),
		  title=COALESCE($4,title), description=$5,
		  priority=COALESCE($6,priority), status=COALESCE($7,status),
		  failure_type=$8, symptoms=$9,
		  requested_by_name=$10, assigned_to_name=$11,
		  estimated_cost=$12, actual_cost=$13, notes=$14,
		  submitted_at=COALESCE($15,submitted_at),
		  approved_at=COALESCE($16,approved_at),
		  completed_at=COALESCE($17,completed_at),
		  updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, cid,
		mNullStr(mStr(body, "equipment_id")),
		mStrN(body, "title"), mStrN(body, "description"),
		mStrN(body, "priority"), mStrN(body, "status"),
		mStrN(body, "failure_type"), mStrN(body, "symptoms"),
		mStrN(body, "requested_by_name"), mStrN(body, "assigned_to_name"),
		mIntN(body, "estimated_cost"), mIntN(body, "actual_cost"), mStrN(body, "notes"),
		subAt, appAt, compAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeleteRequest DELETE /maintenance/requests/:id
func (h *MaintenanceHandler) DeleteRequest(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	_, err := h.db.Exec(context.Background(),
		"UPDATE maintenance_requests SET is_active=FALSE,updated_at=NOW() WHERE id=$1 AND company_id=$2",
		c.Param("id"), cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ── ORDERS ────────────────────────────────────────────────────────────────────

// ListOrders GET /maintenance/orders
func (h *MaintenanceHandler) ListOrders(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	search := c.Query("search")
	status := c.Query("status")
	oType := c.Query("order_type")
	pg, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	lim, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if pg < 1 {
		pg = 1
	}
	if lim < 1 || lim > 200 {
		lim = 50
	}
	offset := (pg - 1) * lim

	args := []interface{}{cid}
	where := []string{"mo.company_id=$1", "mo.is_active=TRUE"}
	idx := 2

	if search != "" {
		where = append(where, fmt.Sprintf("(mo.order_number ILIKE $%d OR mo.title ILIKE $%d)", idx, idx))
		args = append(args, "%"+search+"%")
		idx++
	}
	if status != "" {
		where = append(where, fmt.Sprintf("mo.status=$%d", idx))
		args = append(args, status)
		idx++
	}
	if oType != "" {
		where = append(where, fmt.Sprintf("mo.order_type=$%d", idx))
		args = append(args, oType)
		idx++
	}
	wc := strings.Join(where, " AND ")

	var total int
	_ = h.db.QueryRow(context.Background(),
		fmt.Sprintf("SELECT COUNT(*) FROM maintenance_orders mo WHERE %s", wc), args...).Scan(&total)

	args = append(args, lim, offset)
	rows, err := h.db.Query(context.Background(), fmt.Sprintf(`
		SELECT mo.id,mo.order_number,mo.order_type,mo.status,mo.priority,mo.title,
		       mo.description,mo.assigned_technician,
		       mo.planned_start_date,mo.planned_end_date,
		       mo.actual_start_date,mo.actual_end_date,
		       mo.estimated_hours,mo.actual_hours,
		       mo.labor_cost,mo.parts_cost,mo.other_cost,mo.total_cost,
		       mo.next_service_date,mo.color,mo.created_at,
		       e.id,e.name,e.code,e.location,
		       COALESCE((SELECT COUNT(*) FROM maintenance_order_lines l WHERE l.order_id=mo.id),0)
		FROM maintenance_orders mo
		LEFT JOIN equipment e ON e.id=mo.equipment_id
		WHERE %s ORDER BY mo.created_at DESC LIMIT $%d OFFSET $%d
	`, wc, idx, idx+1), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	now := time.Now()
	items := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, onum, ot, st, pri, title                    string
			desc, tech, color, nextSvc                       *string
			ps, pe, as2, ae                                  *string
			estH, actH, lc, pc2, oc, tc                     float64
			createdAt                                        time.Time
			eqID, eqName, eqCode, eqLoc                     *string
			lineCount                                        int
		)
		if err := rows.Scan(&id, &onum, &ot, &st, &pri, &title,
			&desc, &tech, &ps, &pe, &as2, &ae, &estH, &actH,
			&lc, &pc2, &oc, &tc, &nextSvc, &color, &createdAt,
			&eqID, &eqName, &eqCode, &eqLoc, &lineCount); err != nil {
			continue
		}
		overdue := false
		if pe != nil && *pe != "" && st != "completed" && st != "cancelled" {
			if t, err := time.Parse("2006-01-02", *pe); err == nil {
				overdue = now.After(t)
			}
		}
		items = append(items, map[string]interface{}{
			"id": id, "order_number": onum, "order_type": ot, "status": st,
			"priority": pri, "title": title, "description": desc,
			"assigned_technician": tech,
			"planned_start_date": ps, "planned_end_date": pe,
			"actual_start_date": as2, "actual_end_date": ae,
			"estimated_hours": estH, "actual_hours": actH,
			"labor_cost": lc, "parts_cost": pc2, "other_cost": oc, "total_cost": tc,
			"next_service_date": nextSvc, "color": color, "created_at": createdAt,
			"equipment_id": eqID, "equipment_name": eqName,
			"equipment_code": eqCode, "equipment_location": eqLoc,
			"line_count": lineCount, "overdue": overdue,
		})
	}

	srows, _ := h.db.Query(context.Background(),
		`SELECT status,COUNT(*) FROM maintenance_orders WHERE company_id=$1 AND is_active=TRUE GROUP BY status`, cid)
	defer srows.Close()
	summary := []map[string]interface{}{}
	for srows.Next() {
		var s string
		var cnt int
		if err := srows.Scan(&s, &cnt); err == nil {
			summary = append(summary, map[string]interface{}{"status": s, "count": cnt})
		}
	}

	trows, _ := h.db.Query(context.Background(),
		`SELECT order_type,COUNT(*) FROM maintenance_orders WHERE company_id=$1 AND is_active=TRUE GROUP BY order_type`, cid)
	defer trows.Close()
	typeSummary := []map[string]interface{}{}
	for trows.Next() {
		var t string
		var cnt int
		if err := trows.Scan(&t, &cnt); err == nil {
			typeSummary = append(typeSummary, map[string]interface{}{"type": t, "count": cnt})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items, "total": total, "page": pg, "limit": lim,
		"summary": summary, "type_summary": typeSummary,
	})
}

// GetOrder GET /maintenance/orders/:id
func (h *MaintenanceHandler) GetOrder(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	var (
		onum, ot, st, pri, title                            string
		desc, tech, wp, findings, color, nextSvc             *string
		ps, pe, as2, ae                                      *string
		estH, actH, lc, pc2, oc, tc                         float64
		createdAt                                            time.Time
		eqID, eqName, eqCode                                *string
	)
	err := h.db.QueryRow(context.Background(), `
		SELECT mo.order_number,mo.order_type,mo.status,mo.priority,mo.title,
		       mo.description,mo.assigned_technician,mo.work_performed,mo.findings,
		       mo.planned_start_date,mo.planned_end_date,mo.actual_start_date,mo.actual_end_date,
		       mo.estimated_hours,mo.actual_hours,
		       mo.labor_cost,mo.parts_cost,mo.other_cost,mo.total_cost,
		       mo.next_service_date,mo.color,mo.created_at,
		       e.id,e.name,e.code
		FROM maintenance_orders mo
		LEFT JOIN equipment e ON e.id=mo.equipment_id
		WHERE mo.id=$1 AND mo.company_id=$2 AND mo.is_active=TRUE
	`, id, cid).Scan(&onum, &ot, &st, &pri, &title,
		&desc, &tech, &wp, &findings, &ps, &pe, &as2, &ae,
		&estH, &actH, &lc, &pc2, &oc, &tc, &nextSvc, &color, &createdAt,
		&eqID, &eqName, &eqCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	lrows, _ := h.db.Query(context.Background(),
		`SELECT id,part_name,part_number,quantity,unit,unit_cost,total_cost,notes
		 FROM maintenance_order_lines WHERE order_id=$1 ORDER BY created_at`, id)
	defer lrows.Close()
	lines := []map[string]interface{}{}
	for lrows.Next() {
		var lid, pname string
		var pnum, unit, notes *string
		var qty, uc, totc float64
		if err := lrows.Scan(&lid, &pname, &pnum, &qty, &unit, &uc, &totc, &notes); err == nil {
			lines = append(lines, map[string]interface{}{
				"id": lid, "part_name": pname, "part_number": pnum,
				"quantity": qty, "unit": unit, "unit_cost": uc, "total_cost": totc, "notes": notes,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"id": id, "order_number": onum, "order_type": ot, "status": st, "priority": pri, "title": title,
		"description": desc, "assigned_technician": tech, "work_performed": wp, "findings": findings,
		"planned_start_date": ps, "planned_end_date": pe,
		"actual_start_date": as2, "actual_end_date": ae,
		"estimated_hours": estH, "actual_hours": actH,
		"labor_cost": lc, "parts_cost": pc2, "other_cost": oc, "total_cost": tc,
		"next_service_date": nextSvc, "color": color, "created_at": createdAt,
		"equipment_id": eqID, "equipment_name": eqName, "equipment_code": eqCode,
		"lines": lines,
	})
}

// CreateOrder POST /maintenance/orders
func (h *MaintenanceHandler) CreateOrder(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	title := mStr(body, "title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title required"})
		return
	}
	year := time.Now().Year()
	var n int
	_ = h.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM maintenance_orders WHERE company_id=$1 AND EXTRACT(YEAR FROM created_at)=$2",
		cid, year).Scan(&n)
	onum := fmt.Sprintf("WO-%d-%s", year, mPad(n+1, 4))
	ot := mStrD(body, "order_type", "corrective")
	color := mColorType(ot)
	var id string
	err := h.db.QueryRow(context.Background(), `
		INSERT INTO maintenance_orders(company_id,request_id,equipment_id,order_number,order_type,
		  status,priority,title,description,assigned_technician,
		  planned_start_date,planned_end_date,
		  estimated_hours,labor_cost,parts_cost,other_cost,total_cost,color)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		RETURNING id`,
		cid, mNullStr(mStr(body, "request_id")), mNullStr(mStr(body, "equipment_id")),
		onum, ot,
		mStrD(body, "status", "draft"), mStrD(body, "priority", "medium"),
		title, mStrN(body, "description"), mStrN(body, "assigned_technician"),
		mDateN(body, "planned_start_date"), mDateN(body, "planned_end_date"),
		mF64D(body, "estimated_hours", 0),
		mF64D(body, "labor_cost", 0), mF64D(body, "parts_cost", 0),
		mF64D(body, "other_cost", 0), mF64D(body, "total_cost", 0),
		color,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "order_number": onum, "message": "created"})
}

// UpdateOrder PUT /maintenance/orders/:id
func (h *MaintenanceHandler) UpdateOrder(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	lc := mF64D(body, "labor_cost", 0)
	pc2 := mF64D(body, "parts_cost", 0)
	oc := mF64D(body, "other_cost", 0)
	tc := lc + pc2 + oc
	ot := mStr(body, "order_type")
	clr := ""
	if ot != "" {
		clr = mColorType(ot)
	}
	_, err := h.db.Exec(context.Background(), `
		UPDATE maintenance_orders SET
		  equipment_id=COALESCE($3::uuid,equipment_id),
		  order_type=COALESCE($4,order_type), status=COALESCE($5,status),
		  priority=COALESCE($6,priority), title=COALESCE($7,title),
		  description=$8, assigned_technician=$9,
		  planned_start_date=$10, planned_end_date=$11,
		  actual_start_date=$12, actual_end_date=$13,
		  estimated_hours=$14, actual_hours=$15,
		  labor_cost=$16, parts_cost=$17, other_cost=$18, total_cost=$19,
		  next_service_date=$20,
		  color=CASE WHEN $21!='' THEN $21 ELSE color END,
		  updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, cid,
		mNullStr(mStr(body, "equipment_id")),
		mStrN(body, "order_type"), mStrN(body, "status"),
		mStrN(body, "priority"), mStrN(body, "title"),
		mStrN(body, "description"), mStrN(body, "assigned_technician"),
		mDateN(body, "planned_start_date"), mDateN(body, "planned_end_date"),
		mDateN(body, "actual_start_date"), mDateN(body, "actual_end_date"),
		mF64D(body, "estimated_hours", 0), mF64D(body, "actual_hours", 0),
		lc, pc2, oc, tc,
		mDateN(body, "next_service_date"), clr,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// CompleteOrder PUT /maintenance/orders/:id/complete
func (h *MaintenanceHandler) CompleteOrder(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	lc := mF64D(body, "labor_cost", 0)
	pc2 := mF64D(body, "parts_cost", 0)
	oc := mF64D(body, "other_cost", 0)
	tc := lc + pc2 + oc
	today := time.Now().Format("2006-01-02")
	nextSvc := mStrN(body, "next_service_date")

	tx, err := h.db.Begin(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
		UPDATE maintenance_orders SET
		  status='completed', actual_end_date=$3,
		  actual_hours=COALESCE($4,actual_hours),
		  labor_cost=$5, parts_cost=$6, other_cost=$7, total_cost=$8,
		  work_performed=$9, findings=$10, next_service_date=$11, updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, cid, today, mIntN(body, "actual_hours"),
		lc, pc2, oc, tc,
		mStrN(body, "work_performed"), mStrN(body, "findings"), nextSvc,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var eqID *string
	var oTitle, oType string
	_ = tx.QueryRow(context.Background(),
		"SELECT equipment_id::text,title,order_type FROM maintenance_orders WHERE id=$1", id).
		Scan(&eqID, &oTitle, &oType)

	tech := mStrD(body, "technician_name", "")
	_, err = tx.Exec(context.Background(), `
		INSERT INTO maintenance_history(
		  company_id,equipment_id,order_id,history_type,title,
		  work_performed,findings,technician_name,
		  performed_date,duration_hours,
		  labor_cost,parts_cost,other_cost,total_cost,next_service_date)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		cid, eqID, id, oType, oTitle,
		mStrN(body, "work_performed"), mStrN(body, "findings"),
		mNullStr(tech), today, mF64D(body, "actual_hours", 0),
		lc, pc2, oc, tc, nextSvc,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if eqID != nil && *eqID != "" {
		_, _ = tx.Exec(context.Background(), `
			UPDATE equipment SET
			  last_maintenance_date=$3, next_maintenance_date=$4,
			  status=CASE WHEN status='under_maintenance' THEN 'active' ELSE status END,
			  updated_at=NOW()
			WHERE id=$1 AND company_id=$2`, *eqID, cid, today, nextSvc)
	}

	if err := tx.Commit(context.Background()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "order completed"})
}

// DeleteOrder DELETE /maintenance/orders/:id
func (h *MaintenanceHandler) DeleteOrder(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	_, err := h.db.Exec(context.Background(),
		"UPDATE maintenance_orders SET is_active=FALSE,updated_at=NOW() WHERE id=$1 AND company_id=$2",
		c.Param("id"), cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ── PREVENTIVE PLANS ──────────────────────────────────────────────────────────

// ListPreventivePlans GET /maintenance/preventive-plans
func (h *MaintenanceHandler) ListPreventivePlans(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	rows, err := h.db.Query(context.Background(), `
		SELECT p.id,p.name,p.description,p.frequency_type,p.frequency_value,
		       p.estimated_hours,p.estimated_cost,p.last_performed,p.next_due,
		       p.lead_days,p.auto_create_order,p.assigned_to,p.is_active,p.created_at,
		       p.tasks,p.checklist,
		       e.id,e.name,e.code,e.location
		FROM preventive_maintenance_plans p
		LEFT JOIN equipment e ON e.id=p.equipment_id
		WHERE p.company_id=$1 ORDER BY p.next_due ASC NULLS LAST,p.name ASC
	`, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	now := time.Now()
	items := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, name, ft                             string
			desc, assignedTo                          *string
			lp, nd                                   *string
			estH, estC                               float64
			fv, ld                                   int
			auto2, isActive                          bool
			createdAt                                time.Time
			tasks, checklist                         interface{}
			eqID, eqName, eqCode, eqLoc              *string
		)
		if err := rows.Scan(&id, &name, &desc, &ft, &fv,
			&estH, &estC, &lp, &nd, &ld, &auto2, &assignedTo, &isActive, &createdAt,
			&tasks, &checklist,
			&eqID, &eqName, &eqCode, &eqLoc); err != nil {
			continue
		}
		days := 0
		planStatus := "upcoming"
		if nd != nil && *nd != "" {
			if t, err := time.Parse("2006-01-02", *nd); err == nil {
				days = int(t.Sub(now).Hours() / 24)
				if days < 0 {
					planStatus = "overdue"
				} else if days <= ld {
					planStatus = "due_soon"
				}
			}
		}
		if !isActive {
			planStatus = "inactive"
		}
		items = append(items, map[string]interface{}{
			"id": id, "name": name, "description": desc,
			"frequency_type": ft, "frequency_value": fv,
			"estimated_hours": estH, "estimated_cost": estC,
			"last_performed": lp, "next_due": nd, "lead_days": ld,
			"auto_create_order": auto2, "assigned_to": assignedTo,
			"is_active": isActive, "created_at": createdAt,
			"tasks": tasks, "checklist": checklist,
			"equipment_id": eqID, "equipment_name": eqName,
			"equipment_code": eqCode, "equipment_location": eqLoc,
			"days_until_due": days, "plan_status": planStatus,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

// CreatePreventivePlan POST /maintenance/preventive-plans
func (h *MaintenanceHandler) CreatePreventivePlan(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if mStr(body, "name") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}
	var id string
	err := h.db.QueryRow(context.Background(), `
		INSERT INTO preventive_maintenance_plans(
		  company_id,equipment_id,name,description,frequency_type,frequency_value,
		  estimated_hours,estimated_cost,last_performed,next_due,lead_days,
		  auto_create_order,assigned_to,is_active)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id`,
		cid, mNullStr(mStr(body, "equipment_id")),
		mStr(body, "name"), mStrN(body, "description"),
		mStrD(body, "frequency_type", "monthly"), mIntD(body, "frequency_value", 1),
		mF64D(body, "estimated_hours", 0), mF64D(body, "estimated_cost", 0),
		mDateN(body, "last_performed"), mDateN(body, "next_due"),
		mIntD(body, "lead_days", 7),
		mBoolD(body, "auto_create_order", false),
		mStrN(body, "assigned_to"), mBoolD(body, "is_active", true),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "created"})
}

// UpdatePreventivePlan PUT /maintenance/preventive-plans/:id
func (h *MaintenanceHandler) UpdatePreventivePlan(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec(context.Background(), `
		UPDATE preventive_maintenance_plans SET
		  equipment_id=COALESCE($3::uuid,equipment_id),
		  name=COALESCE($4,name), description=$5,
		  frequency_type=COALESCE($6,frequency_type),
		  frequency_value=COALESCE($7,frequency_value),
		  estimated_hours=$8, estimated_cost=$9,
		  last_performed=$10, next_due=$11,
		  lead_days=COALESCE($12,lead_days),
		  auto_create_order=$13, assigned_to=$14, is_active=$15, updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, cid,
		mNullStr(mStr(body, "equipment_id")),
		mStrN(body, "name"), mStrN(body, "description"),
		mStrN(body, "frequency_type"), mIntN(body, "frequency_value"),
		mF64D(body, "estimated_hours", 0), mF64D(body, "estimated_cost", 0),
		mDateN(body, "last_performed"), mDateN(body, "next_due"),
		mIntN(body, "lead_days"),
		mBoolD(body, "auto_create_order", false),
		mStrN(body, "assigned_to"), mBoolD(body, "is_active", true),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeletePreventivePlan DELETE /maintenance/preventive-plans/:id
func (h *MaintenanceHandler) DeletePreventivePlan(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	_, err := h.db.Exec(context.Background(),
		"UPDATE preventive_maintenance_plans SET is_active=FALSE,updated_at=NOW() WHERE id=$1 AND company_id=$2",
		c.Param("id"), cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ── CALENDAR ──────────────────────────────────────────────────────────────────

// GetCalendar GET /maintenance/calendar
func (h *MaintenanceHandler) GetCalendar(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	year, _ := strconv.Atoi(c.DefaultQuery("year", fmt.Sprintf("%d", time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", fmt.Sprintf("%d", int(time.Now().Month()))))
	if year < 2000 || year > 2100 {
		year = time.Now().Year()
	}
	if month < 1 || month > 12 {
		month = int(time.Now().Month())
	}
	start := fmt.Sprintf("%d-%02d-01", year, month)
	ey, em := year, month+1
	if em > 12 {
		em = 1
		ey++
	}
	end := fmt.Sprintf("%d-%02d-01", ey, em)

	events := []map[string]interface{}{}

	orows, _ := h.db.Query(context.Background(), `
		SELECT mo.id,mo.order_number,mo.title,mo.order_type,mo.status,
		       mo.planned_start_date,mo.planned_end_date,mo.color,
		       COALESCE(e.name,'')
		FROM maintenance_orders mo
		LEFT JOIN equipment e ON e.id=mo.equipment_id
		WHERE mo.company_id=$1 AND mo.is_active=TRUE
		  AND mo.planned_start_date IS NOT NULL
		  AND mo.planned_start_date>=$2::date AND mo.planned_start_date<$3::date
		ORDER BY mo.planned_start_date
	`, cid, start, end)
	if orows != nil {
		defer orows.Close()
		for orows.Next() {
			var id2, onum, title, ot, st string
			var ps, pe, clr *string
			var eqName string
			if err := orows.Scan(&id2, &onum, &title, &ot, &st, &ps, &pe, &clr, &eqName); err == nil {
				col := "#6366f1"
				if clr != nil {
					col = *clr
				}
				events = append(events, map[string]interface{}{
					"id": id2, "type": "order", "order_type": ot, "number": onum,
					"title": title, "status": st, "date": ps, "end_date": pe,
					"color": col, "equipment": eqName,
				})
			}
		}
	}

	prows, _ := h.db.Query(context.Background(), `
		SELECT p.id,p.name,p.frequency_type,p.next_due,p.is_active,
		       COALESCE(e.name,'')
		FROM preventive_maintenance_plans p
		LEFT JOIN equipment e ON e.id=p.equipment_id
		WHERE p.company_id=$1 AND p.next_due IS NOT NULL
		  AND p.next_due>=$2::date AND p.next_due<$3::date
		ORDER BY p.next_due
	`, cid, start, end)
	if prows != nil {
		defer prows.Close()
		for prows.Next() {
			var id2, name, ft string
			var nd *string
			var isActive bool
			var eqName string
			if err := prows.Scan(&id2, &name, &ft, &nd, &isActive, &eqName); err == nil {
				events = append(events, map[string]interface{}{
					"id": id2, "type": "preventive", "title": name,
					"status": "scheduled", "date": nd, "end_date": nd,
					"color": "#22c55e", "equipment": eqName, "is_active": isActive,
				})
			}
		}
	}

	wrows, _ := h.db.Query(context.Background(), `
		SELECT id,code,name,warranty_expiry FROM equipment
		WHERE company_id=$1 AND is_active=TRUE
		  AND warranty_expiry>=$2::date AND warranty_expiry<$3::date
		ORDER BY warranty_expiry
	`, cid, start, end)
	if wrows != nil {
		defer wrows.Close()
		for wrows.Next() {
			var eid, code, name string
			var we *string
			if err := wrows.Scan(&eid, &code, &name, &we); err == nil {
				events = append(events, map[string]interface{}{
					"id": eid, "type": "warranty",
					"title": name + " (" + code + ") - Warranty Expiry",
					"date": we, "color": "#f59e0b",
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"year": year, "month": month, "events": events, "total": len(events)})
}

// ── HISTORY ───────────────────────────────────────────────────────────────────

// ListHistory GET /maintenance/history
func (h *MaintenanceHandler) ListHistory(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	search := c.Query("search")
	htype := c.Query("history_type")
	eqID := c.Query("equipment_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	pg, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	lim, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if pg < 1 {
		pg = 1
	}
	if lim < 1 || lim > 200 {
		lim = 50
	}
	offset := (pg - 1) * lim

	args := []interface{}{cid}
	where := []string{"mh.company_id=$1", "mh.is_active=TRUE"}
	idx := 2

	if search != "" {
		where = append(where, fmt.Sprintf("(mh.title ILIKE $%d OR mh.technician_name ILIKE $%d)", idx, idx))
		args = append(args, "%"+search+"%")
		idx++
	}
	if htype != "" {
		where = append(where, fmt.Sprintf("mh.history_type=$%d", idx))
		args = append(args, htype)
		idx++
	}
	if eqID != "" {
		where = append(where, fmt.Sprintf("mh.equipment_id=$%d", idx))
		args = append(args, eqID)
		idx++
	}
	if dateFrom != "" {
		where = append(where, fmt.Sprintf("mh.performed_date>=$%d", idx))
		args = append(args, dateFrom)
		idx++
	}
	if dateTo != "" {
		where = append(where, fmt.Sprintf("mh.performed_date<=$%d", idx))
		args = append(args, dateTo)
		idx++
	}
	wc := strings.Join(where, " AND ")

	var total int
	_ = h.db.QueryRow(context.Background(),
		fmt.Sprintf("SELECT COUNT(*) FROM maintenance_history mh WHERE %s", wc), args...).Scan(&total)

	args = append(args, lim, offset)
	rows, err := h.db.Query(context.Background(), fmt.Sprintf(`
		SELECT mh.id,mh.history_type,mh.title,mh.description,
		       mh.work_performed,mh.findings,mh.technician_name,
		       mh.performed_date,mh.duration_hours,mh.downtime_hours,
		       mh.labor_cost,mh.parts_cost,mh.other_cost,mh.total_cost,
		       mh.next_service_date,mh.created_at,
		       e.id,e.name,e.code,e.location
		FROM maintenance_history mh
		LEFT JOIN equipment e ON e.id=mh.equipment_id
		WHERE %s ORDER BY mh.performed_date DESC LIMIT $%d OFFSET $%d
	`, wc, idx, idx+1), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	items := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, ht, title                                  string
			desc, wp, findings, tech, nextSvc              *string
			pd                                             *string
			dh, dwh, lc, pc2, oc, tc                      float64
			createdAt                                      time.Time
			eid, ename, ecode, eloc                        *string
		)
		if err := rows.Scan(&id, &ht, &title, &desc, &wp, &findings, &tech,
			&pd, &dh, &dwh, &lc, &pc2, &oc, &tc, &nextSvc, &createdAt,
			&eid, &ename, &ecode, &eloc); err != nil {
			continue
		}
		items = append(items, map[string]interface{}{
			"id": id, "history_type": ht, "title": title, "description": desc,
			"work_performed": wp, "findings": findings, "technician_name": tech,
			"performed_date": pd, "duration_hours": dh, "downtime_hours": dwh,
			"labor_cost": lc, "parts_cost": pc2, "other_cost": oc, "total_cost": tc,
			"next_service_date": nextSvc, "created_at": createdAt,
			"equipment_id": eid, "equipment_name": ename,
			"equipment_code": ecode, "equipment_location": eloc,
		})
	}

	var totCost, totHours, totDowntime float64
	_ = h.db.QueryRow(context.Background(), `
		SELECT COALESCE(SUM(total_cost),0),COALESCE(SUM(duration_hours),0),COALESCE(SUM(downtime_hours),0)
		FROM maintenance_history WHERE company_id=$1 AND is_active=TRUE
	`, cid).Scan(&totCost, &totHours, &totDowntime)

	c.JSON(http.StatusOK, gin.H{
		"items": items, "total": total, "page": pg, "limit": lim,
		"total_cost": totCost, "total_hours": totHours, "total_downtime": totDowntime,
	})
}

// CreateHistory POST /maintenance/history
func (h *MaintenanceHandler) CreateHistory(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if mStr(body, "title") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title required"})
		return
	}
	lc := mF64D(body, "labor_cost", 0)
	pc2 := mF64D(body, "parts_cost", 0)
	oc := mF64D(body, "other_cost", 0)
	tc := lc + pc2 + oc
	pd := mDateD(body, "performed_date")
	var id string
	err := h.db.QueryRow(context.Background(), `
		INSERT INTO maintenance_history(
		  company_id,equipment_id,history_type,title,description,
		  work_performed,findings,technician_name,
		  performed_date,duration_hours,downtime_hours,
		  labor_cost,parts_cost,other_cost,total_cost,next_service_date)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) RETURNING id`,
		cid, mNullStr(mStr(body, "equipment_id")),
		mStrD(body, "history_type", "corrective"),
		mStr(body, "title"), mStrN(body, "description"),
		mStrN(body, "work_performed"), mStrN(body, "findings"),
		mStrN(body, "technician_name"), pd,
		mF64D(body, "duration_hours", 0), mF64D(body, "downtime_hours", 0),
		lc, pc2, oc, tc, mDateN(body, "next_service_date"),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	eqID := mStr(body, "equipment_id")
	if eqID != "" {
		_, _ = h.db.Exec(context.Background(),
			"UPDATE equipment SET last_maintenance_date=$3,updated_at=NOW() WHERE id=$1 AND company_id=$2",
			eqID, cid, pd)
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "created"})
}

// ── DASHBOARD ─────────────────────────────────────────────────────────────────

// GetDashboard GET /maintenance/dashboard
func (h *MaintenanceHandler) GetDashboard(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	now := time.Now()
	msStart := fmt.Sprintf("%d-%02d-01", now.Year(), now.Month())

	var totEq, activeEq, underMaint, overdueEq int
	var totEqVal float64
	_ = h.db.QueryRow(context.Background(), `
		SELECT COUNT(*),
		       COUNT(*) FILTER(WHERE status='active'),
		       COUNT(*) FILTER(WHERE status='under_maintenance'),
		       COUNT(*) FILTER(WHERE next_maintenance_date<NOW() AND status='active'),
		       COALESCE(SUM(current_value),0)
		FROM equipment WHERE company_id=$1 AND is_active=TRUE
	`, cid).Scan(&totEq, &activeEq, &underMaint, &overdueEq, &totEqVal)

	var openOrders, completedMonth, overdueOrders int
	var costMonth float64
	_ = h.db.QueryRow(context.Background(), `
		SELECT COUNT(*) FILTER(WHERE status NOT IN('completed','cancelled')),
		       COUNT(*) FILTER(WHERE status='completed' AND updated_at>=$2::date),
		       COUNT(*) FILTER(WHERE status NOT IN('completed','cancelled') AND planned_end_date<NOW()),
		       COALESCE(SUM(total_cost) FILTER(WHERE status='completed' AND updated_at>=$2::date),0)
		FROM maintenance_orders WHERE company_id=$1 AND is_active=TRUE
	`, cid, msStart).Scan(&openOrders, &completedMonth, &overdueOrders, &costMonth)

	var openReqs, pendingApproval int
	_ = h.db.QueryRow(context.Background(), `
		SELECT COUNT(*) FILTER(WHERE status NOT IN('completed','rejected','cancelled')),
		       COUNT(*) FILTER(WHERE status='submitted')
		FROM maintenance_requests WHERE company_id=$1 AND is_active=TRUE
	`, cid).Scan(&openReqs, &pendingApproval)

	var upcomingPM int
	_ = h.db.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM preventive_maintenance_plans
		WHERE company_id=$1 AND is_active=TRUE
		  AND next_due>=NOW()::date AND next_due<=(NOW()+INTERVAL '30 days')::date
	`, cid).Scan(&upcomingPM)

	// monthly cost trend
	trows, _ := h.db.Query(context.Background(), `
		SELECT TO_CHAR(DATE_TRUNC('month',updated_at),'YYYY-MM'),
		       COALESCE(SUM(total_cost),0),COUNT(*)
		FROM maintenance_orders
		WHERE company_id=$1 AND status='completed' AND is_active=TRUE
		  AND updated_at>=(NOW()-INTERVAL '6 months')
		GROUP BY DATE_TRUNC('month',updated_at) ORDER BY 1
	`, cid)
	defer trows.Close()
	trend := []map[string]interface{}{}
	for trows.Next() {
		var m string
		var cost float64
		var cnt int
		if err := trows.Scan(&m, &cost, &cnt); err == nil {
			trend = append(trend, map[string]interface{}{"month": m, "cost": cost, "count": cnt})
		}
	}

	// equipment by status
	eqRows, _ := h.db.Query(context.Background(),
		`SELECT status,COUNT(*) FROM equipment WHERE company_id=$1 AND is_active=TRUE GROUP BY status`, cid)
	defer eqRows.Close()
	eqStatus := []map[string]interface{}{}
	for eqRows.Next() {
		var s string
		var cnt int
		if err := eqRows.Scan(&s, &cnt); err == nil {
			eqStatus = append(eqStatus, map[string]interface{}{"status": s, "count": cnt})
		}
	}

	// orders by type
	otRows, _ := h.db.Query(context.Background(),
		`SELECT order_type,COUNT(*),COALESCE(SUM(total_cost),0) FROM maintenance_orders WHERE company_id=$1 AND is_active=TRUE GROUP BY order_type`, cid)
	defer otRows.Close()
	ordersByType := []map[string]interface{}{}
	for otRows.Next() {
		var t string
		var cnt int
		var cost float64
		if err := otRows.Scan(&t, &cnt, &cost); err == nil {
			ordersByType = append(ordersByType, map[string]interface{}{"type": t, "count": cnt, "cost": cost})
		}
	}

	// top 5 equipment by cost
	topRows, _ := h.db.Query(context.Background(), `
		SELECT e.id,e.name,e.code,
		       COALESCE(SUM(mo.total_cost),0),COUNT(mo.id)
		FROM equipment e
		LEFT JOIN maintenance_orders mo ON mo.equipment_id=e.id AND mo.is_active=TRUE
		WHERE e.company_id=$1 AND e.is_active=TRUE
		GROUP BY e.id,e.name,e.code ORDER BY 4 DESC LIMIT 5
	`, cid)
	defer topRows.Close()
	topEq := []map[string]interface{}{}
	for topRows.Next() {
		var eid, ename, ecode string
		var cost float64
		var cnt int
		if err := topRows.Scan(&eid, &ename, &ecode, &cost, &cnt); err == nil {
			topEq = append(topEq, map[string]interface{}{
				"id": eid, "name": ename, "code": ecode, "total_cost": cost, "order_count": cnt,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"kpis": map[string]interface{}{
			"total_equipment": totEq, "active_equipment": activeEq,
			"under_maintenance": underMaint, "overdue_equipment": overdueEq,
			"total_equipment_value": totEqVal,
			"open_orders": openOrders, "completed_this_month": completedMonth,
			"overdue_orders": overdueOrders, "total_cost_this_month": costMonth,
			"open_requests": openReqs, "pending_approval": pendingApproval,
			"upcoming_pm": upcomingPM,
		},
		"monthly_trend":    trend,
		"equipment_status": eqStatus,
		"orders_by_type":   ordersByType,
		"top_equipment":    topEq,
	})
}

// ── REPORTS ───────────────────────────────────────────────────────────────────

// GetReports GET /maintenance/reports
func (h *MaintenanceHandler) GetReports(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	year, _ := strconv.Atoi(c.DefaultQuery("year", fmt.Sprintf("%d", time.Now().Year())))
	if year < 2000 || year > 2100 {
		year = time.Now().Year()
	}

	var totOrders, compOrders, prevCnt, corrCnt int
	var totCost, laborCost, partsCost, avgDur float64
	_ = h.db.QueryRow(context.Background(), `
		SELECT COUNT(*),
		       COUNT(*) FILTER(WHERE status='completed'),
		       COUNT(*) FILTER(WHERE order_type='preventive'),
		       COUNT(*) FILTER(WHERE order_type='corrective'),
		       COALESCE(SUM(total_cost),0),
		       COALESCE(SUM(labor_cost),0),
		       COALESCE(SUM(parts_cost),0),
		       COALESCE(AVG(actual_hours) FILTER(WHERE status='completed' AND actual_hours>0),0)
		FROM maintenance_orders
		WHERE company_id=$1 AND is_active=TRUE AND EXTRACT(YEAR FROM created_at)=$2
	`, cid, year).Scan(&totOrders, &compOrders, &prevCnt, &corrCnt,
		&totCost, &laborCost, &partsCost, &avgDur)

	mrows, _ := h.db.Query(context.Background(), `
		SELECT TO_CHAR(DATE_TRUNC('month',created_at),'YYYY-MM'),
		       COUNT(*),COUNT(*) FILTER(WHERE status='completed'),
		       COALESCE(SUM(total_cost),0),COALESCE(SUM(labor_cost),0),COALESCE(SUM(parts_cost),0),
		       COUNT(*) FILTER(WHERE order_type='preventive'),
		       COUNT(*) FILTER(WHERE order_type='corrective')
		FROM maintenance_orders
		WHERE company_id=$1 AND is_active=TRUE AND EXTRACT(YEAR FROM created_at)=$2
		GROUP BY DATE_TRUNC('month',created_at) ORDER BY 1
	`, cid, year)
	defer mrows.Close()
	monthly := []map[string]interface{}{}
	for mrows.Next() {
		var mo string
		var orders, completed, prev2, corr2 int
		var tc2, lc2, pc2 float64
		if err := mrows.Scan(&mo, &orders, &completed, &tc2, &lc2, &pc2, &prev2, &corr2); err == nil {
			monthly = append(monthly, map[string]interface{}{
				"month": mo, "orders": orders, "completed": completed,
				"total_cost": tc2, "labor_cost": lc2, "parts_cost": pc2,
				"preventive": prev2, "corrective": corr2,
			})
		}
	}

	catRows, _ := h.db.Query(context.Background(), `
		SELECT COALESCE(e.category,'Uncategorized'),COUNT(mo.id),
		       COALESCE(SUM(mo.total_cost),0),COALESCE(SUM(mo.actual_hours),0)
		FROM maintenance_orders mo
		LEFT JOIN equipment e ON e.id=mo.equipment_id
		WHERE mo.company_id=$1 AND mo.is_active=TRUE AND EXTRACT(YEAR FROM mo.created_at)=$2
		GROUP BY e.category ORDER BY 3 DESC
	`, cid, year)
	defer catRows.Close()
	byCat := []map[string]interface{}{}
	for catRows.Next() {
		var cat string
		var cnt int
		var cost, hrs float64
		if err := catRows.Scan(&cat, &cnt, &cost, &hrs); err == nil {
			byCat = append(byCat, map[string]interface{}{
				"category": cat, "count": cnt, "cost": cost, "hours": hrs,
			})
		}
	}

	stRows, _ := h.db.Query(context.Background(), `
		SELECT status,COUNT(*) FROM maintenance_orders
		WHERE company_id=$1 AND is_active=TRUE AND EXTRACT(YEAR FROM created_at)=$2
		GROUP BY status
	`, cid, year)
	defer stRows.Close()
	bySt := []map[string]interface{}{}
	for stRows.Next() {
		var s string
		var cnt int
		if err := stRows.Scan(&s, &cnt); err == nil {
			bySt = append(bySt, map[string]interface{}{"status": s, "count": cnt})
		}
	}

	mtbfRows, _ := h.db.Query(context.Background(), `
		SELECT e.id,e.name,e.code,
		       COUNT(mh.id),
		       COALESCE(SUM(mh.downtime_hours),0),
		       COALESCE(SUM(mh.total_cost),0)
		FROM equipment e
		LEFT JOIN maintenance_history mh ON mh.equipment_id=e.id AND mh.is_active=TRUE
		  AND EXTRACT(YEAR FROM mh.performed_date)=$2
		WHERE e.company_id=$1 AND e.is_active=TRUE
		GROUP BY e.id,e.name,e.code HAVING COUNT(mh.id)>0
		ORDER BY 4 DESC LIMIT 10
	`, cid, year)
	defer mtbfRows.Close()
	mtbf := []map[string]interface{}{}
	for mtbfRows.Next() {
		var eid, ename, ecode string
		var cnt int
		var dwn, cost float64
		if err := mtbfRows.Scan(&eid, &ename, &ecode, &cnt, &dwn, &cost); err == nil {
			mtbfVal := 0.0
			if cnt > 0 {
				mtbfVal = float64(365) / float64(cnt)
			}
			mtbf = append(mtbf, map[string]interface{}{
				"id": eid, "name": ename, "code": ecode,
				"failure_count": cnt, "total_downtime_hours": dwn,
				"total_cost": cost, "mtbf_days": mtbfVal,
			})
		}
	}

	pmRows, _ := h.db.Query(context.Background(), `
		SELECT p.id,p.name,p.next_due::text,
		       COALESCE(e.name,'N/A'),
		       (p.next_due-NOW()::date)
		FROM preventive_maintenance_plans p
		LEFT JOIN equipment e ON e.id=p.equipment_id
		WHERE p.company_id=$1 AND p.is_active=TRUE AND p.next_due IS NOT NULL
		  AND p.next_due>=NOW()::date AND p.next_due<=(NOW()+INTERVAL '60 days')::date
		ORDER BY p.next_due LIMIT 10
	`, cid)
	defer pmRows.Close()
	upcomingPM := []map[string]interface{}{}
	for pmRows.Next() {
		var pid, pname, nd, eqName string
		var daysLeft int
		if err := pmRows.Scan(&pid, &pname, &nd, &eqName, &daysLeft); err == nil {
			upcomingPM = append(upcomingPM, map[string]interface{}{
				"id": pid, "name": pname, "next_due": nd,
				"equipment": eqName, "days_left": daysLeft,
			})
		}
	}

	compRate := 0.0
	if totOrders > 0 {
		compRate = float64(compOrders) / float64(totOrders) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"year": year,
		"kpis": map[string]interface{}{
			"total_orders": totOrders, "completed_orders": compOrders,
			"preventive_count": prevCnt, "corrective_count": corrCnt,
			"total_cost": totCost, "labor_cost": laborCost, "parts_cost": partsCost,
			"avg_duration_hours": avgDur, "completion_rate": compRate,
		},
		"monthly": monthly, "by_category": byCat, "by_status": bySt,
		"mtbf": mtbf, "upcoming_pm": upcomingPM,
	})
}
