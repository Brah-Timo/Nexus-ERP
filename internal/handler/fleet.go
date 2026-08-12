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

// FleetHandler handles all fleet management endpoints.
type FleetHandler struct{ db *pgxpool.Pool }

// ── helpers (fl-prefix to avoid conflicts) ───────────────────────────────────

func flStr(m map[string]interface{}, k string) string {
	if v, ok := m[k]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func flStrD(m map[string]interface{}, k, def string) string {
	s := flStr(m, k)
	if s == "" {
		return def
	}
	return s
}

func flStrN(m map[string]interface{}, k string) *string {
	s := flStr(m, k)
	if s == "" {
		return nil
	}
	return &s
}

func flF64D(m map[string]interface{}, k string, def float64) float64 {
	if v, ok := m[k]; ok && v != nil {
		switch x := v.(type) {
		case float64:
			return x
		case float32:
			return float64(x)
		case int:
			return float64(x)
		case string:
			if f, err := strconv.ParseFloat(x, 64); err == nil {
				return f
			}
		}
	}
	return def
}

func flIntD(m map[string]interface{}, k string, def int) int {
	if v, ok := m[k]; ok && v != nil {
		switch x := v.(type) {
		case float64:
			return int(x)
		case int:
			return x
		case string:
			if i, err := strconv.Atoi(x); err == nil {
				return i
			}
		}
	}
	return def
}

func flIntN(m map[string]interface{}, k string) *int {
	if v, ok := m[k]; ok && v != nil {
		switch x := v.(type) {
		case float64:
			i := int(x)
			return &i
		case int:
			return &x
		case string:
			if i, err := strconv.Atoi(x); err == nil {
				return &i
			}
		}
	}
	return nil
}

func flBoolD(m map[string]interface{}, k string, def bool) bool {
	if v, ok := m[k]; ok && v != nil {
		switch x := v.(type) {
		case bool:
			return x
		case string:
			return strings.EqualFold(x, "true") || x == "1"
		}
	}
	return def
}

func flDateN(m map[string]interface{}, k string) *string {
	s := flStr(m, k)
	if s == "" || s == "0001-01-01" {
		return nil
	}
	if len(s) >= 10 {
		d := s[:10]
		return &d
	}
	return &s
}

func flPad(n, w int) string {
	return fmt.Sprintf("%0*d", w, n)
}

// ── Vehicles ─────────────────────────────────────────────────────────────────

// ListVehicles GET /fleet/vehicles
func (h *FleetHandler) ListVehicles(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}

	search := c.Query("search")
	status := c.Query("status")
	vtype := c.Query("type")
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
	idx := 2
	wheres := []string{"v.company_id=$1"}

	if search != "" {
		wheres = append(wheres, fmt.Sprintf(
			"(v.plate_number ILIKE $%d OR v.make ILIKE $%d OR v.model ILIKE $%d OR v.vin ILIKE $%d)",
			idx, idx, idx, idx))
		args = append(args, "%"+search+"%")
		idx++
	}
	if status != "" {
		wheres = append(wheres, fmt.Sprintf("v.status=$%d", idx))
		args = append(args, status)
		idx++
	}
	if vtype != "" {
		wheres = append(wheres, fmt.Sprintf("v.vehicle_type=$%d", idx))
		args = append(args, vtype)
		idx++
	}

	where := strings.Join(wheres, " AND ")

	var total int
	_ = h.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM fleet_vehicles v WHERE "+where, args...).Scan(&total)

	args = append(args, lim, offset)
	rows, err := h.db.Query(context.Background(), `
		SELECT v.id, v.plate_number, v.vin, v.make, v.model, v.year, v.color,
		       v.vehicle_type, v.fuel_type, v.status, v.mileage_at_fill,
		       v.purchase_date, v.purchase_price, v.current_value,
		       v.insurance_policy, v.insurance_expiry, v.registration_expiry,
		       v.technical_visit_expiry, v.department, v.notes, v.image_url,
		       v.assigned_driver_id,
		       CONCAT(d.first_name,' ',d.last_name) AS driver_name,
		       v.created_at
		FROM fleet_vehicles v
		LEFT JOIN fleet_drivers d ON v.assigned_driver_id = d.id
		WHERE `+where+`
		ORDER BY v.created_at DESC
		LIMIT $`+strconv.Itoa(idx)+` OFFSET $`+strconv.Itoa(idx+1),
		args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	vehicles := []map[string]interface{}{}
	statusSum := map[string]int{}
	for rows.Next() {
		var (
			id, plate, mk, mdl, vt, ft, st, dept string
			yr, odo                               int
			vin, clr, pd, iv, pp, cv             *string
			ie, re, tve, notes, imgURL           *string
			adID, driverName                     *string
			createdAt                            time.Time
		)
		if err := rows.Scan(&id, &plate, &vin, &mk, &mdl, &yr, &clr,
			&vt, &ft, &st, &odo,
			&pd, &pp, &cv,
			&iv, &ie, &re, &tve, &dept, &notes, &imgURL,
			&adID, &driverName, &createdAt); err != nil {
			continue
		}
		statusSum[st]++
		vehicles = append(vehicles, map[string]interface{}{
			"id": id, "plate_number": plate, "vin": vin, "make": mk, "model": mdl,
			"year": yr, "color": clr, "vehicle_type": vt, "fuel_type": ft,
			"status": st, "mileage_at_fill": odo,
			"purchase_date": pd, "purchase_price": pp, "current_value": cv,
			"insurance_policy": iv, "insurance_expiry": ie,
			"registration_expiry": re, "technical_visit_expiry": tve,
			"department": dept, "notes": notes, "image_url": imgURL,
			"assigned_driver_id": adID, "driver_name": driverName,
			"created_at": createdAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"vehicles": vehicles, "total": total,
		"page": pg, "limit": lim, "status_summary": statusSum,
	})
}

// GetVehicle GET /fleet/vehicles/:id
func (h *FleetHandler) GetVehicle(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	var (
		plate, mk, mdl, vt, ft, st      string
		yr, odo                          int
		vin, clr, dept, notes, imgURL   *string
		pd, pp, cv                       *string
		iv, ie, re, tve                  *string
		ftc                              *float64
		sc                               *int
		adID, driverName                 *string
		createdAt                        time.Time
	)
	err := h.db.QueryRow(context.Background(), `
		SELECT v.plate_number, v.vin, v.make, v.model, v.year, v.color,
		       v.vehicle_type, v.fuel_type, v.status, v.mileage_at_fill,
		       v.fuel_tank_capacity, v.seating_capacity,
		       v.purchase_date, v.purchase_price, v.current_value,
		       v.insurance_policy, v.insurance_expiry, v.registration_expiry,
		       v.technical_visit_expiry, v.department, v.notes, v.image_url,
		       v.assigned_driver_id,
		       CONCAT(d.first_name,' ',d.last_name) AS driver_name,
		       v.created_at
		FROM fleet_vehicles v
		LEFT JOIN fleet_drivers d ON v.assigned_driver_id = d.id
		WHERE v.id=$1 AND v.company_id=$2
	`, id, cid).Scan(&plate, &vin, &mk, &mdl, &yr, &clr,
		&vt, &ft, &st, &odo, &ftc, &sc,
		&pd, &pp, &cv, &iv, &ie, &re, &tve,
		&dept, &notes, &imgURL, &adID, &driverName, &createdAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vehicle not found"})
		return
	}

	// Recent fuel logs
	frows, _ := h.db.Query(context.Background(), `
		SELECT id, fill_date, mileage_at_fill, liters, total_cost, fuel_type
		FROM fleet_fuel_logs
		WHERE vehicle_id=$1 AND company_id=$2
		ORDER BY fill_date DESC LIMIT 5`, id, cid)
	defer frows.Close()
	fuelLogs := []map[string]interface{}{}
	for frows.Next() {
		var fid, fdate, fftype string
		var fodo int
		var fliters, ftotalcost float64
		if err := frows.Scan(&fid, &fdate, &fodo, &fliters, &ftotalcost, &fftype); err == nil {
			fuelLogs = append(fuelLogs, map[string]interface{}{
				"id": fid, "fill_date": fdate, "mileage_at_fill": fodo,
				"liters": fliters, "total_cost": ftotalcost, "fuel_type": fftype,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id, "plate_number": plate, "vin": vin, "make": mk, "model": mdl,
		"year": yr, "color": clr, "vehicle_type": vt, "fuel_type": ft,
		"status": st, "mileage_at_fill": odo, "fuel_tank_capacity": ftc,
		"seating_capacity": sc,
		"purchase_date": pd, "purchase_price": pp, "current_value": cv,
		"insurance_policy": iv, "insurance_expiry": ie,
		"registration_expiry": re, "technical_visit_expiry": tve,
		"department": dept, "notes": notes, "image_url": imgURL,
		"assigned_driver_id": adID, "driver_name": driverName,
		"recent_fuel": fuelLogs, "created_at": createdAt,
	})
}

// CreateVehicle POST /fleet/vehicles
func (h *FleetHandler) CreateVehicle(c *gin.Context) {
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

	plate := flStr(body, "plate_number")
	if plate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plate_number is required"})
		return
	}
	make_ := flStrD(body, "make", "Unknown")
	model_ := flStrD(body, "model", "Unknown")
	year_ := flIntD(body, "year", time.Now().Year())
	odo := flIntD(body, "mileage_at_fill", 0)
	vtype := flStrD(body, "vehicle_type", "car")
	ftype := flStrD(body, "fuel_type", "diesel")
	status := flStrD(body, "status", "active")

	var id string
	err := h.db.QueryRow(context.Background(), `
		INSERT INTO fleet_vehicles (
			company_id, plate_number, vin, make, model, year, color,
			vehicle_type, fuel_type, status, mileage_at_fill,
			fuel_tank_capacity, seating_capacity,
			purchase_date, purchase_price, current_value,
			insurance_policy, insurance_expiry, registration_expiry,
			technical_visit_expiry, department, notes, image_url
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,
			$8,$9,$10,$11,
			$12,$13,
			$14,$15,$16,
			$17,$18,$19,
			$20,$21,$22,$23
		) RETURNING id`,
		cid, plate,
		flStrN(body, "vin"), make_, model_, year_, flStrN(body, "color"),
		vtype, ftype, status, odo,
		flIntN(body, "fuel_tank_capacity"), flIntN(body, "seating_capacity"),
		flDateN(body, "purchase_date"), flF64D(body, "purchase_price", 0), flF64D(body, "current_value", 0),
		flStrN(body, "insurance_policy"), flDateN(body, "insurance_expiry"),
		flDateN(body, "registration_expiry"), flDateN(body, "technical_visit_expiry"),
		flStrN(body, "department"), flStrN(body, "notes"), flStrN(body, "image_url"),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "vehicle created"})
}

// UpdateVehicle PUT /fleet/vehicles/:id
func (h *FleetHandler) UpdateVehicle(c *gin.Context) {
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
		UPDATE fleet_vehicles SET
			plate_number=$3, vin=$4, make=$5, model=$6, year=$7, color=$8,
			vehicle_type=$9, fuel_type=$10, status=$11, mileage_at_fill=$12,
			fuel_tank_capacity=$13, seating_capacity=$14,
			purchase_date=$15, purchase_price=$16, current_value=$17,
			insurance_policy=$18, insurance_expiry=$19,
			registration_expiry=$20, technical_visit_expiry=$21,
			department=$22, notes=$23, image_url=$24,
			updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, cid,
		flStr(body, "plate_number"),
		flStrN(body, "vin"),
		flStrD(body, "make", "Unknown"),
		flStrD(body, "model", "Unknown"),
		flIntD(body, "year", time.Now().Year()),
		flStrN(body, "color"),
		flStrD(body, "vehicle_type", "car"),
		flStrD(body, "fuel_type", "diesel"),
		flStrD(body, "status", "active"),
		flIntD(body, "mileage_at_fill", 0),
		flIntN(body, "fuel_tank_capacity"),
		flIntN(body, "seating_capacity"),
		flDateN(body, "purchase_date"),
		flF64D(body, "purchase_price", 0),
		flF64D(body, "current_value", 0),
		flStrN(body, "insurance_policy"),
		flDateN(body, "insurance_expiry"),
		flDateN(body, "registration_expiry"),
		flDateN(body, "technical_visit_expiry"),
		flStrN(body, "department"),
		flStrN(body, "notes"),
		flStrN(body, "image_url"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "vehicle updated"})
}

// DeleteVehicle DELETE /fleet/vehicles/:id
func (h *FleetHandler) DeleteVehicle(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	_, err := h.db.Exec(context.Background(),
		"DELETE FROM fleet_vehicles WHERE id=$1 AND company_id=$2",
		id, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "vehicle deleted"})
}

// ── Drivers ──────────────────────────────────────────────────────────────────

// ListDrivers GET /fleet/drivers
func (h *FleetHandler) ListDrivers(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	search := c.Query("search")
	status := c.Query("status")
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
	idx := 2
	wheres := []string{"company_id=$1"}

	if search != "" {
		wheres = append(wheres, fmt.Sprintf(
			"(first_name ILIKE $%d OR last_name ILIKE $%d OR license_number ILIKE $%d OR phone ILIKE $%d)",
			idx, idx, idx, idx))
		args = append(args, "%"+search+"%")
		idx++
	}
	if status != "" {
		wheres = append(wheres, fmt.Sprintf("status=$%d", idx))
		args = append(args, status)
		idx++
	}
	where := strings.Join(wheres, " AND ")

	var total int
	_ = h.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM fleet_drivers WHERE "+where, args...).Scan(&total)

	args = append(args, lim, offset)
	rows, err := h.db.Query(context.Background(), `
		SELECT id, first_name, last_name, phone, email, national_id,
		       license_number, license_class, license_expiry, license_issue_date,
		       status, hire_date, address, emergency_contact, emergency_phone,
		       notes, photo_url, created_at
		FROM fleet_drivers WHERE `+where+`
		ORDER BY created_at DESC
		LIMIT $`+strconv.Itoa(idx)+` OFFSET $`+strconv.Itoa(idx+1),
		args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	drivers := []map[string]interface{}{}
	statusSum := map[string]int{}
	for rows.Next() {
		var (
			id, fn, ln, lic, lc, st    string
			phone, email, natID        *string
			le, lid, hireDate, addr    *string
			ec, ep, notes, photo       *string
			createdAt                  time.Time
		)
		if err := rows.Scan(&id, &fn, &ln, &phone, &email, &natID,
			&lic, &lc, &le, &lid, &st, &hireDate, &addr,
			&ec, &ep, &notes, &photo, &createdAt); err != nil {
			continue
		}
		statusSum[st]++
		drivers = append(drivers, map[string]interface{}{
			"id": id, "first_name": fn, "last_name": ln,
			"full_name":        fn + " " + ln,
			"phone":            phone, "email": email, "national_id": natID,
			"license_number":   lic, "license_class": lc,
			"license_expiry":   le, "license_issue_date": lid,
			"status": st, "hire_date": hireDate, "address": addr,
			"emergency_contact": ec, "emergency_phone": ep,
			"notes": notes, "photo_url": photo, "created_at": createdAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"drivers": drivers, "total": total,
		"page": pg, "limit": lim, "status_summary": statusSum,
	})
}

// CreateDriver POST /fleet/drivers
func (h *FleetHandler) CreateDriver(c *gin.Context) {
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
	fn := flStr(body, "first_name")
	ln := flStr(body, "last_name")
	lic := flStr(body, "license_number")
	if fn == "" || ln == "" || lic == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "first_name, last_name, license_number required"})
		return
	}
	var id string
	err := h.db.QueryRow(context.Background(), `
		INSERT INTO fleet_drivers (
			company_id, first_name, last_name, phone, email, national_id,
			license_number, license_class, license_expiry, license_issue_date,
			status, hire_date, address, emergency_contact, emergency_phone, notes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING id`,
		cid, fn, ln,
		flStrN(body, "phone"), flStrN(body, "email"), flStrN(body, "national_id"),
		lic, flStrD(body, "license_class", "B"),
		flDateN(body, "license_expiry"), flDateN(body, "license_issue_date"),
		flStrD(body, "status", "active"),
		flDateN(body, "hire_date"),
		flStrN(body, "address"),
		flStrN(body, "emergency_contact"), flStrN(body, "emergency_phone"),
		flStrN(body, "notes"),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "driver created"})
}

// UpdateDriver PUT /fleet/drivers/:id
func (h *FleetHandler) UpdateDriver(c *gin.Context) {
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
		UPDATE fleet_drivers SET
			first_name=$3, last_name=$4, phone=$5, email=$6, national_id=$7,
			license_number=$8, license_class=$9, license_expiry=$10, license_issue_date=$11,
			status=$12, hire_date=$13, address=$14,
			emergency_contact=$15, emergency_phone=$16, notes=$17,
			updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, cid,
		flStr(body, "first_name"), flStr(body, "last_name"),
		flStrN(body, "phone"), flStrN(body, "email"), flStrN(body, "national_id"),
		flStr(body, "license_number"), flStrD(body, "license_class", "B"),
		flDateN(body, "license_expiry"), flDateN(body, "license_issue_date"),
		flStrD(body, "status", "active"),
		flDateN(body, "hire_date"), flStrN(body, "address"),
		flStrN(body, "emergency_contact"), flStrN(body, "emergency_phone"),
		flStrN(body, "notes"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "driver updated"})
}

// DeleteDriver DELETE /fleet/drivers/:id
func (h *FleetHandler) DeleteDriver(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	_, err := h.db.Exec(context.Background(),
		"DELETE FROM fleet_drivers WHERE id=$1 AND company_id=$2",
		id, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "driver deleted"})
}

// ── Assignments ───────────────────────────────────────────────────────────────

// ListAssignments GET /fleet/assignments
func (h *FleetHandler) ListAssignments(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	status := c.Query("status")
	vehicleID := c.Query("vehicle_id")
	driverID := c.Query("driver_id")
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
	idx := 2
	wheres := []string{"a.company_id=$1"}

	if status != "" {
		wheres = append(wheres, fmt.Sprintf("a.status=$%d", idx))
		args = append(args, status)
		idx++
	}
	if vehicleID != "" {
		wheres = append(wheres, fmt.Sprintf("a.vehicle_id=$%d", idx))
		args = append(args, vehicleID)
		idx++
	}
	if driverID != "" {
		wheres = append(wheres, fmt.Sprintf("a.driver_id=$%d", idx))
		args = append(args, driverID)
		idx++
	}
	where := strings.Join(wheres, " AND ")

	var total int
	_ = h.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM fleet_assignments a WHERE "+where, args...).Scan(&total)

	args = append(args, lim, offset)
	rows, err := h.db.Query(context.Background(), `
		SELECT a.id, a.vehicle_id, a.driver_id,
		       v.plate_number, v.make, v.model,
		       CONCAT(d.first_name,' ',d.last_name) AS driver_name,
		       a.start_date, a.end_date, a.start_odometer, a.end_odometer,
		       a.purpose, a.destination, a.notes, a.status, a.created_at
		FROM fleet_assignments a
		JOIN fleet_vehicles v ON a.vehicle_id = v.id
		JOIN fleet_drivers  d ON a.driver_id  = d.id
		WHERE `+where+`
		ORDER BY a.start_date DESC
		LIMIT $`+strconv.Itoa(idx)+` OFFSET $`+strconv.Itoa(idx+1),
		args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, vid, did, plate, mk, mdl, dname, st string
			startDate                                string
			endDate, purpose, dest, notes           *string
			startOdo                                 int
			endOdo                                   *int
			createdAt                                time.Time
		)
		if err := rows.Scan(&id, &vid, &did, &plate, &mk, &mdl, &dname,
			&startDate, &endDate, &startOdo, &endOdo,
			&purpose, &dest, &notes, &st, &createdAt); err != nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"id": id, "vehicle_id": vid, "driver_id": did,
			"plate_number": plate, "vehicle_name": mk + " " + mdl, "driver_name": dname,
			"start_date": startDate, "end_date": endDate,
			"start_odometer": startOdo, "end_odometer": endOdo,
			"purpose": purpose, "destination": dest, "notes": notes,
			"status": st, "created_at": createdAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"assignments": list, "total": total, "page": pg, "limit": lim})
}

// CreateAssignment POST /fleet/assignments
func (h *FleetHandler) CreateAssignment(c *gin.Context) {
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
	vid := flStr(body, "vehicle_id")
	did := flStr(body, "driver_id")
	sd := flStr(body, "start_date")
	if vid == "" || did == "" || sd == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vehicle_id, driver_id, start_date required"})
		return
	}
	var id string
	err := h.db.QueryRow(context.Background(), `
		INSERT INTO fleet_assignments (
			company_id, vehicle_id, driver_id, start_date, end_date,
			start_odometer, purpose, destination, notes, status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		cid, vid, did, sd,
		flDateN(body, "end_date"),
		flIntD(body, "start_odometer", 0),
		flStrN(body, "purpose"), flStrN(body, "destination"), flStrN(body, "notes"),
		flStrD(body, "status", "active"),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Update vehicle assigned driver
	_, _ = h.db.Exec(context.Background(),
		"UPDATE fleet_vehicles SET assigned_driver_id=$1, updated_at=NOW() WHERE id=$2 AND company_id=$3",
		did, vid, cid)
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "assignment created"})
}

// UpdateAssignment PUT /fleet/assignments/:id
func (h *FleetHandler) UpdateAssignment(c *gin.Context) {
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
		UPDATE fleet_assignments SET
			start_date=$3, end_date=$4,
			start_odometer=$5, end_odometer=$6,
			purpose=$7, destination=$8, notes=$9, status=$10,
			updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, cid,
		flStr(body, "start_date"),
		flDateN(body, "end_date"),
		flIntD(body, "start_odometer", 0),
		flIntN(body, "end_odometer"),
		flStrN(body, "purpose"), flStrN(body, "destination"), flStrN(body, "notes"),
		flStrD(body, "status", "active"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "assignment updated"})
}

// DeleteAssignment DELETE /fleet/assignments/:id
func (h *FleetHandler) DeleteAssignment(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	_, err := h.db.Exec(context.Background(),
		"DELETE FROM fleet_assignments WHERE id=$1 AND company_id=$2",
		id, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "assignment deleted"})
}

// ── Fuel Logs ─────────────────────────────────────────────────────────────────

// ListFuelLogs GET /fleet/fuel
func (h *FleetHandler) ListFuelLogs(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	vehicleID := c.Query("vehicle_id")
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
	idx := 2
	wheres := []string{"f.company_id=$1"}

	if vehicleID != "" {
		wheres = append(wheres, fmt.Sprintf("f.vehicle_id=$%d", idx))
		args = append(args, vehicleID)
		idx++
	}
	if dateFrom != "" {
		wheres = append(wheres, fmt.Sprintf("f.fill_date>=$%d", idx))
		args = append(args, dateFrom)
		idx++
	}
	if dateTo != "" {
		wheres = append(wheres, fmt.Sprintf("f.fill_date<=$%d", idx))
		args = append(args, dateTo)
		idx++
	}
	where := strings.Join(wheres, " AND ")

	var total int
	var totalLiters, totalCost float64
	_ = h.db.QueryRow(context.Background(),
		"SELECT COUNT(*), COALESCE(SUM(liters),0), COALESCE(SUM(total_cost),0) FROM fleet_fuel_logs f WHERE "+where, args...).
		Scan(&total, &totalLiters, &totalCost)

	args = append(args, lim, offset)
	rows, err := h.db.Query(context.Background(), `
		SELECT f.id, f.vehicle_id, f.driver_id,
		       v.plate_number, v.make, v.model,
		       COALESCE(CONCAT(d.first_name,' ',d.last_name),'') AS driver_name,
		       f.fill_date, f.mileage_at_fill, f.liters, f.price_per_liter, f.total_cost,
		       f.fuel_type, f.fuel_station, f.is_full_tank, f.notes, f.created_at
		FROM fleet_fuel_logs f
		JOIN fleet_vehicles v ON f.vehicle_id = v.id
		LEFT JOIN fleet_drivers d ON f.driver_id = d.id
		WHERE `+where+`
		ORDER BY f.fill_date DESC
		LIMIT $`+strconv.Itoa(idx)+` OFFSET $`+strconv.Itoa(idx+1),
		args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, vid, plate, mk, mdl, dname, logDate, ftype string
			odo                                             int
			liters, ppl, totalC                            float64
			did, station, notes                            *string
			fullTank                                       bool
			createdAt                                      time.Time
		)
		if err := rows.Scan(&id, &vid, &did, &plate, &mk, &mdl, &dname,
			&logDate, &odo, &liters, &ppl, &totalC,
			&ftype, &station, &fullTank, &notes, &createdAt); err != nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"id": id, "vehicle_id": vid, "driver_id": did,
			"plate_number": plate, "vehicle_name": mk + " " + mdl,
			"driver_name":   dname,
			"fill_date":      logDate, "mileage_at_fill": odo,
			"liters":        liters, "price_per_liter": ppl, "total_cost": totalC,
			"fuel_type":     ftype, "fuel_station": station,
			"is_full_tank":     fullTank, "notes": notes, "created_at": createdAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"fuel_logs":    list,
		"total":        total,
		"total_liters": totalLiters,
		"total_cost":   totalCost,
		"page": pg, "limit": lim,
	})
}

// CreateFuelLog POST /fleet/fuel
func (h *FleetHandler) CreateFuelLog(c *gin.Context) {
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
	vid := flStr(body, "vehicle_id")
	if vid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vehicle_id required"})
		return
	}
	logDate := flStrD(body, "fill_date", time.Now().Format("2006-01-02"))
	odo := flIntD(body, "mileage_at_fill", 0)
	liters := flF64D(body, "liters", 0)
	ppl := flF64D(body, "price_per_liter", 0)

	var id string
	err := h.db.QueryRow(context.Background(), `
		INSERT INTO fleet_fuel_logs (
			company_id, vehicle_id, driver_id, fill_date, mileage_at_fill,
			liters, price_per_liter, fuel_type, fuel_station, is_full_tank, notes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`,
		cid, vid, flStrN(body, "driver_id"), logDate, odo,
		liters, ppl,
		flStrD(body, "fuel_type", "diesel"),
		flStrN(body, "fuel_station"),
		flBoolD(body, "is_full_tank", true),
		flStrN(body, "notes"),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Update vehicle odometer if new reading is higher
	if odo > 0 {
		_, _ = h.db.Exec(context.Background(),
			"UPDATE fleet_vehicles SET current_mileage=GREATEST(current_mileage,$1), updated_at=NOW() WHERE id=$2 AND company_id=$3",
			odo, vid, cid)
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "fuel log created"})
}

// UpdateFuelLog PUT /fleet/fuel/:id
func (h *FleetHandler) UpdateFuelLog(c *gin.Context) {
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
		UPDATE fleet_fuel_logs SET
			fill_date=$3, mileage_at_fill=$4, liters=$5, price_per_liter=$6,
			fuel_type=$7, fuel_station=$8, is_full_tank=$9, notes=$10,
			updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, cid,
		flStrD(body, "fill_date", time.Now().Format("2006-01-02")),
		flIntD(body, "mileage_at_fill", 0),
		flF64D(body, "liters", 0),
		flF64D(body, "price_per_liter", 0),
		flStrD(body, "fuel_type", "diesel"),
		flStrN(body, "fuel_station"),
		flBoolD(body, "is_full_tank", true),
		flStrN(body, "notes"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "fuel log updated"})
}

// DeleteFuelLog DELETE /fleet/fuel/:id
func (h *FleetHandler) DeleteFuelLog(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	_, err := h.db.Exec(context.Background(),
		"DELETE FROM fleet_fuel_logs WHERE id=$1 AND company_id=$2",
		id, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "fuel log deleted"})
}

// ── Fleet Maintenance ─────────────────────────────────────────────────────────

// ListFleetMaintenance GET /fleet/maintenance
func (h *FleetHandler) ListFleetMaintenance(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	vehicleID := c.Query("vehicle_id")
	status := c.Query("status")
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
	idx := 2
	wheres := []string{"m.company_id=$1"}

	if vehicleID != "" {
		wheres = append(wheres, fmt.Sprintf("m.vehicle_id=$%d", idx))
		args = append(args, vehicleID)
		idx++
	}
	if status != "" {
		wheres = append(wheres, fmt.Sprintf("m.status=$%d", idx))
		args = append(args, status)
		idx++
	}
	where := strings.Join(wheres, " AND ")

	var total int
	_ = h.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM fleet_maintenance m WHERE "+where, args...).Scan(&total)

	args = append(args, lim, offset)
	rows, err := h.db.Query(context.Background(), `
		SELECT m.id, m.vehicle_id, v.plate_number, v.make, v.model,
		       m.title, m.description, m.maintenance_type, m.status,
		       m.scheduled_date, m.completed_date, m.mileage_at_fill,
		       m.next_service_km, m.next_service_date,
		       m.technician, m.garage_name,
		       m.labor_cost, m.parts_cost, m.total_cost,
		       m.work_performed, m.notes, m.created_at
		FROM fleet_maintenance m
		JOIN fleet_vehicles v ON m.vehicle_id = v.id
		WHERE `+where+`
		ORDER BY m.scheduled_date DESC
		LIMIT $`+strconv.Itoa(idx)+` OFFSET $`+strconv.Itoa(idx+1),
		args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	statusSum := map[string]int{}
	for rows.Next() {
		var (
			id, vid, plate, mk, mdl, title, mtype, st string
			laborC, partsC, totalC                    float64
			odo                                        int
			desc, sdate, cdate, tech, garage          *string
			nsk                                        *int
			nsdate, workPerf, notes                   *string
			createdAt                                  time.Time
		)
		if err := rows.Scan(&id, &vid, &plate, &mk, &mdl,
			&title, &desc, &mtype, &st,
			&sdate, &cdate, &odo, &nsk, &nsdate,
			&tech, &garage, &laborC, &partsC, &totalC,
			&workPerf, &notes, &createdAt); err != nil {
			continue
		}
		statusSum[st]++
		isOverdue := false
		if sdate != nil && (st == "scheduled" || st == "in_progress") {
			if t, err2 := time.Parse("2006-01-02", *sdate); err2 == nil {
				isOverdue = t.Before(time.Now().Truncate(24 * time.Hour))
			}
		}
		list = append(list, map[string]interface{}{
			"id": id, "vehicle_id": vid,
			"plate_number": plate, "vehicle_name": mk + " " + mdl,
			"title": title, "description": desc,
			"maintenance_type": mtype, "status": st, "overdue": isOverdue,
			"scheduled_date": sdate, "completed_date": cdate,
			"mileage_at_fill": odo, "next_service_km": nsk, "next_service_date": nsdate,
			"technician": tech, "garage_name": garage,
			"labor_cost": laborC, "parts_cost": partsC, "total_cost": totalC,
			"work_performed": workPerf, "notes": notes, "created_at": createdAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"maintenance":    list,
		"total":          total,
		"page":           pg,
		"limit":          lim,
		"status_summary": statusSum,
	})
}

// CreateFleetMaintenance POST /fleet/maintenance
func (h *FleetHandler) CreateFleetMaintenance(c *gin.Context) {
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
	vid := flStr(body, "vehicle_id")
	title := flStr(body, "title")
	if vid == "" || title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vehicle_id, title required"})
		return
	}
	laborC := flF64D(body, "labor_cost", 0)
	partsC := flF64D(body, "parts_cost", 0)
	otherC := flF64D(body, "other_cost", 0)
	totalC := laborC + partsC + otherC

	var id string
	err := h.db.QueryRow(context.Background(), `
		INSERT INTO fleet_maintenance (
			company_id, vehicle_id, title, description, maintenance_type, status,
			scheduled_date, completed_date, mileage_at_fill,
			next_service_km, next_service_date,
			technician, garage_name,
			labor_cost, parts_cost, total_cost, work_performed, notes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		RETURNING id`,
		cid, vid, title,
		flStrN(body, "description"),
		flStrD(body, "maintenance_type", "routine"),
		flStrD(body, "status", "scheduled"),
		flDateN(body, "scheduled_date"),
		flDateN(body, "completed_date"),
		flIntD(body, "mileage_at_fill", 0),
		flIntN(body, "next_service_km"),
		flDateN(body, "next_service_date"),
		flStrN(body, "technician"), flStrN(body, "garage_name"),
		laborC, partsC, totalC,
		flStrN(body, "work_performed"), flStrN(body, "notes"),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "maintenance record created"})
}

// UpdateFleetMaintenance PUT /fleet/maintenance/:id
func (h *FleetHandler) UpdateFleetMaintenance(c *gin.Context) {
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
	laborC := flF64D(body, "labor_cost", 0)
	partsC := flF64D(body, "parts_cost", 0)
	otherC := flF64D(body, "other_cost", 0)
	totalC := laborC + partsC + otherC

	_, err := h.db.Exec(context.Background(), `
		UPDATE fleet_maintenance SET
			title=$3, description=$4, maintenance_type=$5, status=$6,
			scheduled_date=$7, completed_date=$8, mileage_at_fill=$9,
			next_service_km=$10, next_service_date=$11,
			technician=$12, garage_name=$13,
			labor_cost=$14, parts_cost=$15, total_cost=$16,
			work_performed=$17, notes=$18, updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, cid,
		flStr(body, "title"),
		flStrN(body, "description"),
		flStrD(body, "maintenance_type", "routine"),
		flStrD(body, "status", "scheduled"),
		flDateN(body, "scheduled_date"),
		flDateN(body, "completed_date"),
		flIntD(body, "mileage_at_fill", 0),
		flIntN(body, "next_service_km"),
		flDateN(body, "next_service_date"),
		flStrN(body, "technician"), flStrN(body, "garage_name"),
		laborC, partsC, totalC,
		flStrN(body, "work_performed"), flStrN(body, "notes"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "maintenance record updated"})
}

// DeleteFleetMaintenance DELETE /fleet/maintenance/:id
func (h *FleetHandler) DeleteFleetMaintenance(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	_, err := h.db.Exec(context.Background(),
		"DELETE FROM fleet_maintenance WHERE id=$1 AND company_id=$2",
		id, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "maintenance record deleted"})
}

// ── Expenses ─────────────────────────────────────────────────────────────────

// ListExpenses GET /fleet/expenses
func (h *FleetHandler) ListExpenses(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	vehicleID := c.Query("vehicle_id")
	expType := c.Query("type")
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
	idx := 2
	wheres := []string{"e.company_id=$1"}

	if vehicleID != "" {
		wheres = append(wheres, fmt.Sprintf("e.vehicle_id=$%d", idx))
		args = append(args, vehicleID)
		idx++
	}
	if expType != "" {
		wheres = append(wheres, fmt.Sprintf("e.expense_type=$%d", idx))
		args = append(args, expType)
		idx++
	}
	if dateFrom != "" {
		wheres = append(wheres, fmt.Sprintf("e.expense_date>=$%d", idx))
		args = append(args, dateFrom)
		idx++
	}
	if dateTo != "" {
		wheres = append(wheres, fmt.Sprintf("e.expense_date<=$%d", idx))
		args = append(args, dateTo)
		idx++
	}
	where := strings.Join(wheres, " AND ")

	var total int
	var totalAmount float64
	_ = h.db.QueryRow(context.Background(),
		"SELECT COUNT(*), COALESCE(SUM(amount),0) FROM fleet_expenses e WHERE "+where, args...).
		Scan(&total, &totalAmount)

	args = append(args, lim, offset)
	rows, err := h.db.Query(context.Background(), `
		SELECT e.id, e.vehicle_id, e.driver_id,
		       v.plate_number, v.make, v.model,
		       COALESCE(CONCAT(d.first_name,' ',d.last_name),'') AS driver_name,
		       e.expense_type, e.expense_date, e.amount,
		       e.description, e.reference_number, e.notes, e.created_at
		FROM fleet_expenses e
		JOIN fleet_vehicles v ON e.vehicle_id = v.id
		LEFT JOIN fleet_drivers d ON e.driver_id = d.id
		WHERE `+where+`
		ORDER BY e.expense_date DESC
		LIMIT $`+strconv.Itoa(idx)+` OFFSET $`+strconv.Itoa(idx+1),
		args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, vid, plate, mk, mdl, dname, etype, edate string
			amount                                        float64
			did, desc, refNum, notes                     *string
			createdAt                                     time.Time
		)
		if err := rows.Scan(&id, &vid, &did, &plate, &mk, &mdl, &dname,
			&etype, &edate, &amount, &desc, &refNum, &notes, &createdAt); err != nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"id": id, "vehicle_id": vid, "driver_id": did,
			"plate_number": plate, "vehicle_name": mk + " " + mdl,
			"driver_name":      dname,
			"expense_type":     etype, "expense_date": edate,
			"amount":           amount,
			"description":      desc, "reference_number": refNum,
			"notes":            notes, "created_at": createdAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"expenses":     list,
		"total":        total,
		"total_amount": totalAmount,
		"page": pg, "limit": lim,
	})
}

// CreateExpense POST /fleet/expenses
func (h *FleetHandler) CreateExpense(c *gin.Context) {
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
	vid := flStr(body, "vehicle_id")
	if vid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vehicle_id required"})
		return
	}
	edate := flStrD(body, "expense_date", time.Now().Format("2006-01-02"))

	var id string
	err := h.db.QueryRow(context.Background(), `
		INSERT INTO fleet_expenses (
			company_id, vehicle_id, driver_id, expense_type,
			expense_date, amount, description, reference_number, notes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		cid, vid, flStrN(body, "driver_id"),
		flStrD(body, "expense_type", "other"),
		edate, flF64D(body, "amount", 0),
		flStrN(body, "description"), flStrN(body, "reference_number"), flStrN(body, "notes"),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "expense created"})
}

// UpdateExpense PUT /fleet/expenses/:id
func (h *FleetHandler) UpdateExpense(c *gin.Context) {
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
		UPDATE fleet_expenses SET
			expense_type=$3, expense_date=$4, amount=$5,
			description=$6, reference_number=$7, notes=$8, updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, cid,
		flStrD(body, "expense_type", "other"),
		flStrD(body, "expense_date", time.Now().Format("2006-01-02")),
		flF64D(body, "amount", 0),
		flStrN(body, "description"), flStrN(body, "reference_number"), flStrN(body, "notes"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "expense updated"})
}

// DeleteExpense DELETE /fleet/expenses/:id
func (h *FleetHandler) DeleteExpense(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	id := c.Param("id")
	_, err := h.db.Exec(context.Background(),
		"DELETE FROM fleet_expenses WHERE id=$1 AND company_id=$2",
		id, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "expense deleted"})
}

// ── Reports / Dashboard ───────────────────────────────────────────────────────

// GetFleetDashboard GET /fleet/dashboard
func (h *FleetHandler) GetFleetDashboard(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}

	// Vehicle summary
	vrows, _ := h.db.Query(context.Background(), `
		SELECT status, COUNT(*) FROM fleet_vehicles
		WHERE company_id=$1 GROUP BY status`, cid)
	defer vrows.Close()
	vehicleSum := map[string]int{}
	totalVehicles := 0
	for vrows.Next() {
		var st string
		var cnt int
		if err := vrows.Scan(&st, &cnt); err == nil {
			vehicleSum[st] = cnt
			totalVehicles += cnt
		}
	}

	// Driver summary
	drows, _ := h.db.Query(context.Background(), `
		SELECT status, COUNT(*) FROM fleet_drivers
		WHERE company_id=$1 GROUP BY status`, cid)
	defer drows.Close()
	driverSum := map[string]int{}
	totalDrivers := 0
	for drows.Next() {
		var st string
		var cnt int
		if err := drows.Scan(&st, &cnt); err == nil {
			driverSum[st] = cnt
			totalDrivers += cnt
		}
	}

	// This month fuel cost
	var monthFuelCost, monthFuelLiters float64
	_ = h.db.QueryRow(context.Background(), `
		SELECT COALESCE(SUM(total_cost),0), COALESCE(SUM(liters),0)
		FROM fleet_fuel_logs
		WHERE company_id=$1
		  AND DATE_TRUNC('month', fill_date) = DATE_TRUNC('month', CURRENT_DATE)`, cid).
		Scan(&monthFuelCost, &monthFuelLiters)

	// This month expenses
	var monthExpenses float64
	_ = h.db.QueryRow(context.Background(), `
		SELECT COALESCE(SUM(amount),0) FROM fleet_expenses
		WHERE company_id=$1
		  AND DATE_TRUNC('month', expense_date) = DATE_TRUNC('month', CURRENT_DATE)`, cid).
		Scan(&monthExpenses)

	// Overdue maintenance count
	var overdueCount int
	_ = h.db.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM fleet_maintenance
		WHERE company_id=$1
		  AND status IN ('scheduled','in_progress')
		  AND scheduled_date < CURRENT_DATE`, cid).Scan(&overdueCount)

	// Expiring documents (next 30 days)
	var expiringInsurance, expiringRegistration int
	_ = h.db.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM fleet_vehicles
		WHERE company_id=$1
		  AND insurance_expiry BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days'`, cid).
		Scan(&expiringInsurance)
	_ = h.db.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM fleet_vehicles
		WHERE company_id=$1
		  AND registration_expiry BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days'`, cid).
		Scan(&expiringRegistration)

	// Monthly cost trend (last 12 months)
	mrows, _ := h.db.Query(context.Background(), `
		SELECT TO_CHAR(month, 'YYYY-MM') AS m,
		       COALESCE(SUM(fuel_cost),0) AS fuel,
		       COALESCE(SUM(exp_cost),0) AS expenses
		FROM (
		  SELECT DATE_TRUNC('month', fill_date) AS month,
		         SUM(total_cost) AS fuel_cost, 0 AS exp_cost
		  FROM fleet_fuel_logs
		  WHERE company_id=$1
		    AND fill_date >= CURRENT_DATE - INTERVAL '12 months'
		  GROUP BY DATE_TRUNC('month', fill_date)
		  UNION ALL
		  SELECT DATE_TRUNC('month', expense_date),
		         0, SUM(amount)
		  FROM fleet_expenses
		  WHERE company_id=$1
		    AND expense_date >= CURRENT_DATE - INTERVAL '12 months'
		  GROUP BY DATE_TRUNC('month', expense_date)
		) combined
		GROUP BY month ORDER BY month`, cid, cid)
	defer mrows.Close()
	monthlyCosts := []map[string]interface{}{}
	for mrows.Next() {
		var month string
		var fuel, exp float64
		if err := mrows.Scan(&month, &fuel, &exp); err == nil {
			monthlyCosts = append(monthlyCosts, map[string]interface{}{
				"month": month, "fuel_cost": fuel,
				"expense_cost": exp, "total": fuel + exp,
			})
		}
	}

	// Top vehicles by cost
	tvrows, _ := h.db.Query(context.Background(), `
		SELECT v.plate_number, v.make||' '||v.model AS name,
		       COALESCE(f.fuel_total,0) AS fuel_total,
		       COALESCE(e.exp_total,0) AS exp_total,
		       COALESCE(f.fuel_total,0) + COALESCE(e.exp_total,0) AS grand_total
		FROM fleet_vehicles v
		LEFT JOIN (
		  SELECT vehicle_id, SUM(total_cost) AS fuel_total
		  FROM fleet_fuel_logs WHERE company_id=$1 GROUP BY vehicle_id
		) f ON f.vehicle_id = v.id
		LEFT JOIN (
		  SELECT vehicle_id, SUM(amount) AS exp_total
		  FROM fleet_expenses WHERE company_id=$1 GROUP BY vehicle_id
		) e ON e.vehicle_id = v.id
		WHERE v.company_id=$1
		ORDER BY grand_total DESC LIMIT 8`, cid, cid, cid)
	defer tvrows.Close()
	topVehicles := []map[string]interface{}{}
	for tvrows.Next() {
		var plate, name string
		var fuelT, expT, grandT float64
		if err := tvrows.Scan(&plate, &name, &fuelT, &expT, &grandT); err == nil {
			topVehicles = append(topVehicles, map[string]interface{}{
				"plate_number": plate, "name": name,
				"fuel_cost": fuelT, "expense_cost": expT, "total_cost": grandT,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"vehicle_summary":     vehicleSum,
		"total_vehicles":      totalVehicles,
		"driver_summary":      driverSum,
		"total_drivers":       totalDrivers,
		"month_fuel_cost":     monthFuelCost,
		"month_fuel_liters":   monthFuelLiters,
		"month_expenses":      monthExpenses,
		"overdue_maintenance": overdueCount,
		"expiring_insurance":  expiringInsurance,
		"expiring_registration": expiringRegistration,
		"monthly_costs":       monthlyCosts,
		"top_vehicles":        topVehicles,
	})
}

// GetFleetReports GET /fleet/reports
func (h *FleetHandler) GetFleetReports(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company id"})
		return
	}
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)
	if year == 0 {
		year = time.Now().Year()
	}

	// Annual totals
	var totalFuelCost, totalFuelLiters, totalExpenses, totalMaintCost float64
	var totalFuelLogs, totalExpenseRecords, totalMaintRecords int
	_ = h.db.QueryRow(context.Background(), `
		SELECT COUNT(*), COALESCE(SUM(total_cost),0), COALESCE(SUM(liters),0)
		FROM fleet_fuel_logs
		WHERE company_id=$1
		  AND EXTRACT(YEAR FROM fill_date) = $2`, cid, year).
		Scan(&totalFuelLogs, &totalFuelCost, &totalFuelLiters)
	_ = h.db.QueryRow(context.Background(), `
		SELECT COUNT(*), COALESCE(SUM(amount),0)
		FROM fleet_expenses
		WHERE company_id=$1
		  AND EXTRACT(YEAR FROM expense_date) = $2`, cid, year).
		Scan(&totalExpenseRecords, &totalExpenses)
	_ = h.db.QueryRow(context.Background(), `
		SELECT COUNT(*), COALESCE(SUM(total_cost),0)
		FROM fleet_maintenance
		WHERE company_id=$1
		  AND EXTRACT(YEAR FROM COALESCE(completed_date, scheduled_date)) = $2`, cid, year).
		Scan(&totalMaintRecords, &totalMaintCost)

	// Expense breakdown by type
	etrows, _ := h.db.Query(context.Background(), `
		SELECT expense_type, COUNT(*), COALESCE(SUM(amount),0)
		FROM fleet_expenses
		WHERE company_id=$1
		  AND EXTRACT(YEAR FROM expense_date) = $2
		GROUP BY expense_type ORDER BY SUM(amount) DESC`, cid, year)
	defer etrows.Close()
	expByType := []map[string]interface{}{}
	for etrows.Next() {
		var etype string
		var cnt int
		var amt float64
		if err := etrows.Scan(&etype, &cnt, &amt); err == nil {
			expByType = append(expByType, map[string]interface{}{
				"type": etype, "count": cnt, "amount": amt,
			})
		}
	}

	// Monthly breakdown
	mmrows, _ := h.db.Query(context.Background(), `
		SELECT TO_CHAR(month,'YYYY-MM') AS m, fuel, expenses, maintenance
		FROM (
		  SELECT generate_series(
		    DATE_TRUNC('year', TO_DATE($2::text,'YYYY')),
		    DATE_TRUNC('year', TO_DATE($2::text,'YYYY')) + INTERVAL '11 months',
		    '1 month') AS month
		) months
		LEFT JOIN LATERAL (
		  SELECT COALESCE(SUM(total_cost),0) AS fuel
		  FROM fleet_fuel_logs
		  WHERE company_id=$1
		    AND DATE_TRUNC('month',fill_date) = months.month
		) fl ON TRUE
		LEFT JOIN LATERAL (
		  SELECT COALESCE(SUM(amount),0) AS expenses
		  FROM fleet_expenses
		  WHERE company_id=$1
		    AND DATE_TRUNC('month',expense_date) = months.month
		) ex ON TRUE
		LEFT JOIN LATERAL (
		  SELECT COALESCE(SUM(total_cost),0) AS maintenance
		  FROM fleet_maintenance
		  WHERE company_id=$1
		    AND DATE_TRUNC('month',COALESCE(completed_date,scheduled_date)) = months.month
		) mn ON TRUE
		ORDER BY month`, cid, year, cid, cid, cid)
	defer mmrows.Close()
	monthly := []map[string]interface{}{}
	for mmrows.Next() {
		var month string
		var fuel, expenses, maintenance float64
		if err := mmrows.Scan(&month, &fuel, &expenses, &maintenance); err == nil {
			monthly = append(monthly, map[string]interface{}{
				"month": month, "fuel": fuel,
				"expenses": expenses, "maintenance": maintenance,
				"total": fuel + expenses + maintenance,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"year":              year,
		"total_fuel_cost":   totalFuelCost,
		"total_fuel_liters": totalFuelLiters,
		"total_fuel_logs":   totalFuelLogs,
		"total_expenses":    totalExpenses,
		"total_exp_records": totalExpenseRecords,
		"total_maint_cost":  totalMaintCost,
		"total_maint_records": totalMaintRecords,
		"grand_total":       totalFuelCost + totalExpenses + totalMaintCost,
		"expense_by_type":   expByType,
		"monthly":           monthly,
	})
}
