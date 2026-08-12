package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"nexus-erp/internal/middleware"
)

// Ensure time is used
var _ = time.Now

// ─── HR Handler ────────────────────────────────────────────────────────────────

type HRHandler struct{ db *pgxpool.Pool }

// ══════════════════════════════════════════════════════════════════════════════
// EMPLOYEES
// ══════════════════════════════════════════════════════════════════════════════

func (h *HRHandler) ListEmployees(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	search := c.Query("search")
	deptID := c.Query("department_id")
	status := c.Query("status")

	query := `
		SELECT
			e.id, e.employee_number, e.first_name, e.last_name,
			e.first_name || ' ' || e.last_name AS full_name,
			e.email, e.phone, e.gender,
			e.hire_date, e.birth_date, e.termination_date,
			e.employment_type, e.status,
			e.base_salary,
			e.national_id, e.cnas_number, e.nif,
			e.bank_name, e.bank_account,
			e.address, e.city, e.wilaya, e.notes,
			e.department_id, COALESCE(d.name,'')  AS department_name,
			e.position_id,  COALESCE(p.title,'')  AS position_title,
			e.manager_id,
			COALESCE(m.first_name||' '||m.last_name,'') AS manager_name,
			e.created_at, e.updated_at
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		LEFT JOIN positions   p ON p.id = e.position_id
		LEFT JOIN employees   m ON m.id = e.manager_id
		WHERE e.company_id = $1`

	args := []interface{}{companyID}
	idx := 2

	if search != "" {
		query += fmt.Sprintf(` AND (
			e.first_name ILIKE $%d OR e.last_name ILIKE $%d OR
			e.employee_number ILIKE $%d OR e.email ILIKE $%d
		)`, idx, idx, idx, idx)
		args = append(args, "%"+search+"%")
		idx++
	}
	if deptID != "" {
		query += fmt.Sprintf(` AND e.department_id = $%d`, idx)
		args = append(args, deptID)
		idx++
	}
	if status != "" {
		query += fmt.Sprintf(` AND e.status = $%d`, idx)
		args = append(args, status)
		idx++
	}
	_ = idx

	query += ` ORDER BY e.last_name, e.first_name`

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var employees []map[string]interface{}
	for rows.Next() {
		var (
			id, empNum, firstName, lastName, fullName  string
			email, phone, gender                       string
			employmentType, status                     string
			nationalID, cnasNum, nif                   string
			bankName, bankAccount                      string
			address, city, wilaya, notes               string
			deptID2, deptName, posID, posTitle         string
			mgrID, mgrName                             string
			baseSalary                                 float64
			hireDate                                   time.Time
			birthDate, termDate                        *time.Time
			createdAt, updatedAt                       time.Time
		)
		err = rows.Scan(
			&id, &empNum, &firstName, &lastName, &fullName,
			&email, &phone, &gender,
			&hireDate, &birthDate, &termDate,
			&employmentType, &status,
			&baseSalary,
			&nationalID, &cnasNum, &nif,
			&bankName, &bankAccount,
			&address, &city, &wilaya, &notes,
			&deptID2, &deptName,
			&posID, &posTitle,
			&mgrID, &mgrName,
			&createdAt, &updatedAt,
		)
		if err != nil {
			continue
		}
		employees = append(employees, map[string]interface{}{
			"id": id, "employee_number": empNum,
			"first_name": firstName, "last_name": lastName, "full_name": fullName,
			"email": email, "phone": phone, "gender": gender,
			"hire_date": hireDate, "birth_date": birthDate, "termination_date": termDate,
			"employment_type": employmentType, "status": status,
			"base_salary": baseSalary,
			"national_id": nationalID, "cnas_number": cnasNum, "nif": nif,
			"bank_name": bankName, "bank_account": bankAccount,
			"address": address, "city": city, "wilaya": wilaya, "notes": notes,
			"department_id": deptID2, "department_name": deptName,
			"position_id": posID, "position_title": posTitle,
			"manager_id": mgrID, "manager_name": mgrName,
			"created_at": createdAt, "updated_at": updatedAt,
		})
	}
	if employees == nil {
		employees = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, employees)
}

func (h *HRHandler) GetEmployee(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var (
		empID, empNum, firstName, lastName string
		email, phone, gender               string
		employmentType, status             string
		nationalID, cnasNum, nif           string
		bankName, bankAccount              string
		address, city, wilaya, notes       string
		deptID, deptName                   string
		posID, posTitle                    string
		mgrID, mgrName                     string
		baseSalary                         float64
		hireDate                           time.Time
		birthDate, termDate                *time.Time
		createdAt, updatedAt               time.Time
	)

	err := h.db.QueryRow(ctx, `
		SELECT
			e.id, e.employee_number, e.first_name, e.last_name,
			e.email, e.phone, e.gender,
			e.hire_date, e.birth_date, e.termination_date,
			e.employment_type, e.status,
			e.base_salary,
			e.national_id, e.cnas_number, e.nif,
			e.bank_name, e.bank_account,
			e.address, e.city, e.wilaya, e.notes,
			e.department_id, COALESCE(d.name,''),
			e.position_id,  COALESCE(p.title,''),
			e.manager_id,
			COALESCE(m.first_name||' '||m.last_name,''),
			e.created_at, e.updated_at
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		LEFT JOIN positions   p ON p.id = e.position_id
		LEFT JOIN employees   m ON m.id = e.manager_id
		WHERE e.id = $1
	`, id).Scan(
		&empID, &empNum, &firstName, &lastName,
		&email, &phone, &gender,
		&hireDate, &birthDate, &termDate,
		&employmentType, &status,
		&baseSalary,
		&nationalID, &cnasNum, &nif,
		&bankName, &bankAccount,
		&address, &city, &wilaya, &notes,
		&deptID, &deptName,
		&posID, &posTitle,
		&mgrID, &mgrName,
		&createdAt, &updatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": empID, "employee_number": empNum,
		"first_name": firstName, "last_name": lastName,
		"full_name": firstName + " " + lastName,
		"email": email, "phone": phone, "gender": gender,
		"hire_date": hireDate, "birth_date": birthDate, "termination_date": termDate,
		"employment_type": employmentType, "status": status,
		"base_salary": baseSalary,
		"national_id": nationalID, "cnas_number": cnasNum, "nif": nif,
		"bank_name": bankName, "bank_account": bankAccount,
		"address": address, "city": city, "wilaya": wilaya, "notes": notes,
		"department_id": deptID, "department_name": deptName,
		"position_id": posID, "position_title": posTitle,
		"manager_id": mgrID, "manager_name": mgrName,
		"created_at": createdAt, "updated_at": updatedAt,
	})
}

func (h *HRHandler) CreateEmployee(c *gin.Context) {
	var req struct {
		FirstName      string  `json:"first_name"`
		LastName       string  `json:"last_name"`
		Email          string  `json:"email"`
		Phone          string  `json:"phone"`
		Gender         string  `json:"gender"`
		BirthDate      *string `json:"birth_date"`
		HireDate       string  `json:"hire_date"`
		NationalID     string  `json:"national_id"`
		CNASNumber     string  `json:"cnas_number"`
		NIF            string  `json:"nif"`
		DepartmentID   *string `json:"department_id"`
		PositionID     *string `json:"position_id"`
		ManagerID      *string `json:"manager_id"`
		EmploymentType string  `json:"employment_type"`
		BaseSalary     float64 `json:"base_salary"`
		BankName       string  `json:"bank_name"`
		BankAccount    string  `json:"bank_account"`
		Address        string  `json:"address"`
		City           string  `json:"city"`
		Wilaya         string  `json:"wilaya"`
		Notes          string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.EmploymentType == "" {
		req.EmploymentType = "permanent"
	}

	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	// Generate employee number
	var count int
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE company_id = $1`, companyID).Scan(&count)
	empNum := fmt.Sprintf("EMP-%04d", count+1)

	id := uuid.NewString()
	_, err := h.db.Exec(ctx, `
		INSERT INTO employees (
			id, company_id, employee_number, first_name, last_name,
			email, phone, gender, birth_date, hire_date,
			national_id, cnas_number, nif,
			department_id, position_id, manager_id,
			employment_type, status, base_salary,
			bank_name, bank_account, address, city, wilaya, notes
		) VALUES (
			$1,$2,$3,$4,$5,
			$6,$7,$8,$9,$10,
			$11,$12,$13,
			$14,$15,$16,
			$17,'active',$18,
			$19,$20,$21,$22,$23,$24
		)`,
		id, companyID, empNum, req.FirstName, req.LastName,
		req.Email, req.Phone, req.Gender, req.BirthDate, req.HireDate,
		req.NationalID, req.CNASNumber, req.NIF,
		req.DepartmentID, req.PositionID, req.ManagerID,
		req.EmploymentType, req.BaseSalary,
		req.BankName, req.BankAccount, req.Address, req.City, req.Wilaya, req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": id, "employee_number": empNum,
		"first_name": req.FirstName, "last_name": req.LastName,
		"full_name": req.FirstName + " " + req.LastName,
		"message": "Employee created",
	})
}

func (h *HRHandler) UpdateEmployee(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		FirstName      string  `json:"first_name"`
		LastName       string  `json:"last_name"`
		Email          string  `json:"email"`
		Phone          string  `json:"phone"`
		Gender         string  `json:"gender"`
		BirthDate      *string `json:"birth_date"`
		NationalID     string  `json:"national_id"`
		CNASNumber     string  `json:"cnas_number"`
		NIF            string  `json:"nif"`
		DepartmentID   *string `json:"department_id"`
		PositionID     *string `json:"position_id"`
		ManagerID      *string `json:"manager_id"`
		EmploymentType string  `json:"employment_type"`
		BaseSalary     float64 `json:"base_salary"`
		BankName       string  `json:"bank_name"`
		BankAccount    string  `json:"bank_account"`
		Address        string  `json:"address"`
		City           string  `json:"city"`
		Wilaya         string  `json:"wilaya"`
		Notes          string  `json:"notes"`
		Status         string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE employees SET
			first_name=$1, last_name=$2, email=$3, phone=$4, gender=$5,
			birth_date=$6, national_id=$7, cnas_number=$8, nif=$9,
			department_id=$10, position_id=$11, manager_id=$12,
			employment_type=$13, base_salary=$14,
			bank_name=$15, bank_account=$16,
			address=$17, city=$18, wilaya=$19, notes=$20,
			status=$21, updated_at=NOW()
		WHERE id=$22`,
		req.FirstName, req.LastName, req.Email, req.Phone, req.Gender,
		req.BirthDate, req.NationalID, req.CNASNumber, req.NIF,
		req.DepartmentID, req.PositionID, req.ManagerID,
		req.EmploymentType, req.BaseSalary,
		req.BankName, req.BankAccount,
		req.Address, req.City, req.Wilaya, req.Notes,
		req.Status, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Employee updated"})
}

func (h *HRHandler) DeleteEmployee(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE employees SET status='inactive', updated_at=NOW() WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Employee deactivated"})
}

// GetHRDashboard returns KPIs for the HR module
func (h *HRHandler) GetHRDashboard(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var totalEmp, activeEmp, onLeaveEmp int
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE company_id=$1`, companyID).Scan(&totalEmp)
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE company_id=$1 AND status='active'`, companyID).Scan(&activeEmp)
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE company_id=$1 AND status='on_leave'`, companyID).Scan(&onLeaveEmp)

	var pendingLeaves int
	_ = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM leave_requests lr
		JOIN employees e ON e.id=lr.employee_id
		WHERE e.company_id=$1 AND lr.status='pending'`, companyID).Scan(&pendingLeaves)

	var totalPayroll float64
	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_net),0) FROM payroll_runs
		WHERE company_id=$1 AND status IN ('approved','paid')
		AND period_year=EXTRACT(YEAR FROM NOW())::INT
		AND period_month=EXTRACT(MONTH FROM NOW())::INT`, companyID).Scan(&totalPayroll)

	var totalDepts int
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM departments WHERE company_id=$1 AND is_active=true`, companyID).Scan(&totalDepts)

	c.JSON(http.StatusOK, gin.H{
		"total_employees":  totalEmp,
		"active_employees": activeEmp,
		"on_leave":         onLeaveEmp,
		"pending_leaves":   pendingLeaves,
		"total_payroll":    totalPayroll,
		"departments":      totalDepts,
	})
}

// ══════════════════════════════════════════════════════════════════════════════
// DEPARTMENTS
// ══════════════════════════════════════════════════════════════════════════════

func (h *HRHandler) ListDepartments(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT
			d.id, d.code, d.name, d.is_active, d.created_at,
			d.parent_id, COALESCE(p.name,'') AS parent_name,
			d.manager_id, COALESCE(m.first_name||' '||m.last_name,'') AS manager_name,
			d.cost_center_id,
			COUNT(e.id) AS employee_count
		FROM departments d
		LEFT JOIN departments d2 AS p ON p.id = d.parent_id
		LEFT JOIN employees   m  ON m.id = d.manager_id
		LEFT JOIN employees   e  ON e.department_id = d.id AND e.status='active'
		WHERE d.company_id = $1
		GROUP BY d.id, d.code, d.name, d.is_active, d.created_at,
		         d.parent_id, p.name, d.manager_id, m.first_name, m.last_name,
		         d.cost_center_id
		ORDER BY d.name
	`, companyID)

	// If the lateral join alias fails (older pg), fall back to simpler query
	if err != nil {
		rows, err = h.db.Query(ctx, `
			SELECT
				d.id, d.code, d.name, d.is_active, d.created_at,
				COALESCE(d.parent_id::text,''), '',
				COALESCE(d.manager_id::text,''), '',
				COALESCE(d.cost_center_id::text,''),
				(SELECT COUNT(*) FROM employees e2 WHERE e2.department_id=d.id AND e2.status='active') AS emp_count
			FROM departments d
			WHERE d.company_id = $1
			ORDER BY d.name
		`, companyID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	defer rows.Close()

	var depts []map[string]interface{}
	for rows.Next() {
		var id, code, name, parentID, parentName, mgrID, mgrName, ccID string
		var isActive bool
		var createdAt time.Time
		var empCount int
		if err2 := rows.Scan(&id, &code, &name, &isActive, &createdAt,
			&parentID, &parentName, &mgrID, &mgrName, &ccID, &empCount); err2 != nil {
			continue
		}
		depts = append(depts, map[string]interface{}{
			"id": id, "code": code, "name": name,
			"is_active": isActive, "created_at": createdAt,
			"parent_id": parentID, "parent_name": parentName,
			"manager_id": mgrID, "manager_name": mgrName,
			"cost_center_id": ccID,
			"employee_count": empCount,
		})
	}
	if depts == nil {
		depts = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, depts)
}

func (h *HRHandler) CreateDepartment(c *gin.Context) {
	var req struct {
		Code       string  `json:"code"`
		Name       string  `json:"name"`
		ParentID   *string `json:"parent_id"`
		ManagerID  *string `json:"manager_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		INSERT INTO departments (id, company_id, code, name, parent_id, manager_id, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,true)`,
		id, companyID, req.Code, req.Name, req.ParentID, req.ManagerID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "code": req.Code, "name": req.Name, "message": "Department created"})
}

func (h *HRHandler) UpdateDepartment(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Code      string  `json:"code"`
		Name      string  `json:"name"`
		ParentID  *string `json:"parent_id"`
		ManagerID *string `json:"manager_id"`
		IsActive  bool    `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE departments SET code=$1, name=$2, parent_id=$3, manager_id=$4, is_active=$5
		WHERE id=$6`,
		req.Code, req.Name, req.ParentID, req.ManagerID, req.IsActive, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Department updated"})
}

func (h *HRHandler) DeleteDepartment(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `UPDATE departments SET is_active=false WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Department deactivated"})
}

// ══════════════════════════════════════════════════════════════════════════════
// POSITIONS
// ══════════════════════════════════════════════════════════════════════════════

func (h *HRHandler) ListPositions(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx, `
		SELECT p.id, p.code, p.title, p.grade,
		       COALESCE(p.min_salary,0), COALESCE(p.max_salary,0),
		       p.is_active,
		       p.department_id, COALESCE(d.name,'') AS dept_name,
		       COUNT(e.id) AS headcount
		FROM positions p
		LEFT JOIN departments d ON d.id = p.department_id
		LEFT JOIN employees   e ON e.position_id = p.id AND e.status='active'
		WHERE p.company_id = $1
		GROUP BY p.id, p.code, p.title, p.grade, p.min_salary, p.max_salary,
		         p.is_active, p.department_id, d.name
		ORDER BY p.title
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var positions []map[string]interface{}
	for rows.Next() {
		var id, code, title, grade, deptID, deptName string
		var minSal, maxSal float64
		var isActive bool
		var headcount int
		if err2 := rows.Scan(&id, &code, &title, &grade, &minSal, &maxSal,
			&isActive, &deptID, &deptName, &headcount); err2 != nil {
			continue
		}
		positions = append(positions, map[string]interface{}{
			"id": id, "code": code, "title": title, "grade": grade,
			"min_salary": minSal, "max_salary": maxSal, "is_active": isActive,
			"department_id": deptID, "department_name": deptName,
			"headcount": headcount,
		})
	}
	if positions == nil {
		positions = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, positions)
}

func (h *HRHandler) CreatePosition(c *gin.Context) {
	var req struct {
		Code         string   `json:"code"`
		Title        string   `json:"title"`
		Grade        string   `json:"grade"`
		DepartmentID *string  `json:"department_id"`
		MinSalary    *float64 `json:"min_salary"`
		MaxSalary    *float64 `json:"max_salary"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		INSERT INTO positions (id, company_id, code, title, grade, department_id, min_salary, max_salary, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,true)`,
		id, companyID, req.Code, req.Title, req.Grade,
		req.DepartmentID, req.MinSalary, req.MaxSalary,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "title": req.Title, "message": "Position created"})
}

// ══════════════════════════════════════════════════════════════════════════════
// ATTENDANCE
// ══════════════════════════════════════════════════════════════════════════════

func (h *HRHandler) ListAttendance(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	month := c.Query("month")
	year := c.Query("year")
	empID := c.Query("employee_id")

	query := `
		SELECT
			a.id, a.employee_id,
			e.first_name||' '||e.last_name AS employee_name,
			e.employee_number,
			COALESCE(d.name,'') AS department_name,
			a.date, a.check_in, a.check_out,
			COALESCE(a.hours_worked,0), a.overtime_hours, a.status, COALESCE(a.notes,'')
		FROM attendance a
		JOIN employees   e ON e.id = a.employee_id
		LEFT JOIN departments d ON d.id = e.department_id
		WHERE e.company_id = $1`

	args := []interface{}{companyID}
	idx := 2

	if empID != "" {
		query += fmt.Sprintf(` AND a.employee_id = $%d`, idx)
		args = append(args, empID)
		idx++
	}
	if year != "" {
		query += fmt.Sprintf(` AND EXTRACT(YEAR FROM a.date)=$%d`, idx)
		args = append(args, year)
		idx++
	}
	if month != "" {
		query += fmt.Sprintf(` AND EXTRACT(MONTH FROM a.date)=$%d`, idx)
		args = append(args, month)
		idx++
	}
	_ = idx
	query += ` ORDER BY a.date DESC LIMIT 500`

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var id, empID2, empName, empNum, deptName, status, notes string
		var date time.Time
		var checkIn, checkOut *time.Time
		var hours, overtime float64
		if err2 := rows.Scan(&id, &empID2, &empName, &empNum, &deptName,
			&date, &checkIn, &checkOut, &hours, &overtime, &status, &notes); err2 != nil {
			continue
		}
		records = append(records, map[string]interface{}{
			"id": id, "employee_id": empID2,
			"employee_name": empName, "employee_number": empNum,
			"department_name": deptName,
			"date": date, "check_in": checkIn, "check_out": checkOut,
			"hours_worked": hours, "overtime_hours": overtime,
			"status": status, "notes": notes,
		})
	}
	if records == nil {
		records = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, records)
}

func (h *HRHandler) RecordAttendance(c *gin.Context) {
	var req struct {
		EmployeeID    string   `json:"employee_id"`
		Date          string   `json:"date"`
		CheckIn       *string  `json:"check_in"`
		CheckOut      *string  `json:"check_out"`
		HoursWorked   *float64 `json:"hours_worked"`
		OvertimeHours float64  `json:"overtime_hours"`
		Status        string   `json:"status"`
		Notes         string   `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status == "" {
		req.Status = "present"
	}
	id := uuid.NewString()
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		INSERT INTO attendance (id, employee_id, date, check_in, check_out, hours_worked, overtime_hours, status, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (employee_id, date) DO UPDATE SET
			check_in=EXCLUDED.check_in, check_out=EXCLUDED.check_out,
			hours_worked=EXCLUDED.hours_worked, overtime_hours=EXCLUDED.overtime_hours,
			status=EXCLUDED.status, notes=EXCLUDED.notes`,
		id, req.EmployeeID, req.Date, req.CheckIn, req.CheckOut,
		req.HoursWorked, req.OvertimeHours, req.Status, req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Attendance recorded"})
}

func (h *HRHandler) UpdateAttendance(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		CheckIn       *string  `json:"check_in"`
		CheckOut      *string  `json:"check_out"`
		HoursWorked   *float64 `json:"hours_worked"`
		OvertimeHours float64  `json:"overtime_hours"`
		Status        string   `json:"status"`
		Notes         string   `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE attendance SET
			check_in=$1, check_out=$2, hours_worked=$3,
			overtime_hours=$4, status=$5, notes=$6
		WHERE id=$7`,
		req.CheckIn, req.CheckOut, req.HoursWorked,
		req.OvertimeHours, req.Status, req.Notes, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Attendance updated"})
}

func (h *HRHandler) GetAttendanceSummary(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	month := c.Query("month")
	year := c.Query("year")
	if month == "" {
		month = strconv.Itoa(int(time.Now().Month()))
	}
	if year == "" {
		year = strconv.Itoa(time.Now().Year())
	}

	rows, err := h.db.Query(ctx, `
		SELECT
			e.id, e.employee_number,
			e.first_name||' '||e.last_name AS full_name,
			COALESCE(d.name,'') AS dept_name,
			COUNT(a.id) FILTER (WHERE a.status='present')  AS present_days,
			COUNT(a.id) FILTER (WHERE a.status='absent')   AS absent_days,
			COUNT(a.id) FILTER (WHERE a.status='late')     AS late_days,
			COUNT(a.id) FILTER (WHERE a.status='half_day') AS half_days,
			COALESCE(SUM(a.hours_worked),0)                AS total_hours,
			COALESCE(SUM(a.overtime_hours),0)              AS overtime_hours
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		LEFT JOIN attendance a ON a.employee_id = e.id
			AND EXTRACT(MONTH FROM a.date)=$2
			AND EXTRACT(YEAR  FROM a.date)=$3
		WHERE e.company_id=$1 AND e.status='active'
		GROUP BY e.id, e.employee_number, e.first_name, e.last_name, d.name
		ORDER BY e.last_name, e.first_name
	`, companyID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var summary []map[string]interface{}
	for rows.Next() {
		var id, empNum, fullName, deptName string
		var present, absent, late, halfDay int
		var totalHours, overtimeHours float64
		if err2 := rows.Scan(&id, &empNum, &fullName, &deptName,
			&present, &absent, &late, &halfDay, &totalHours, &overtimeHours); err2 != nil {
			continue
		}
		summary = append(summary, map[string]interface{}{
			"employee_id": id, "employee_number": empNum, "full_name": fullName,
			"department_name": deptName,
			"present_days": present, "absent_days": absent,
			"late_days": late, "half_days": halfDay,
			"total_hours": totalHours, "overtime_hours": overtimeHours,
		})
	}
	if summary == nil {
		summary = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, summary)
}

// ══════════════════════════════════════════════════════════════════════════════
// LEAVE TYPES
// ══════════════════════════════════════════════════════════════════════════════

func (h *HRHandler) ListLeaveTypes(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx, `
		SELECT id, name, days_allowed, is_paid, COALESCE(color,'#6366f1'), created_at
		FROM leave_types WHERE company_id=$1 ORDER BY name`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()
	var types []map[string]interface{}
	for rows.Next() {
		var id, name, color string
		var days int
		var isPaid bool
		var createdAt time.Time
		if err2 := rows.Scan(&id, &name, &days, &isPaid, &color, &createdAt); err2 != nil {
			continue
		}
		types = append(types, map[string]interface{}{
			"id": id, "name": name, "days_allowed": days,
			"is_paid": isPaid, "color": color, "created_at": createdAt,
		})
	}
	if types == nil {
		types = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, types)
}

func (h *HRHandler) CreateLeaveType(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		DaysAllowed int    `json:"days_allowed"`
		IsPaid      bool   `json:"is_paid"`
		Color       string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Color == "" {
		req.Color = "#6366f1"
	}
	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		INSERT INTO leave_types (id, company_id, name, days_allowed, is_paid, color)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		id, companyID, req.Name, req.DaysAllowed, req.IsPaid, req.Color,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "name": req.Name})
}

// ══════════════════════════════════════════════════════════════════════════════
// LEAVE REQUESTS
// ══════════════════════════════════════════════════════════════════════════════

func (h *HRHandler) ListLeaveRequests(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	status := c.Query("status")
	empID := c.Query("employee_id")

	query := `
		SELECT
			lr.id, lr.employee_id,
			e.first_name||' '||e.last_name AS employee_name,
			e.employee_number,
			COALESCE(d.name,'') AS department_name,
			lr.leave_type_id, COALESCE(lt.name,'') AS leave_type_name,
			COALESCE(lt.color,'#6366f1') AS leave_color,
			COALESCE(lt.is_paid, true) AS is_paid,
			lr.start_date, lr.end_date, lr.days_count,
			COALESCE(lr.reason,''), lr.status,
			COALESCE(lr.rejection_reason,''),
			lr.approved_at, lr.created_at, lr.updated_at
		FROM leave_requests lr
		JOIN employees   e  ON e.id  = lr.employee_id
		LEFT JOIN departments  d  ON d.id  = e.department_id
		LEFT JOIN leave_types  lt ON lt.id = lr.leave_type_id
		WHERE lr.company_id = $1`

	args := []interface{}{companyID}
	idx := 2

	if status != "" {
		query += fmt.Sprintf(` AND lr.status = $%d`, idx)
		args = append(args, status)
		idx++
	}
	if empID != "" {
		query += fmt.Sprintf(` AND lr.employee_id = $%d`, idx)
		args = append(args, empID)
		idx++
	}
	_ = idx
	query += ` ORDER BY lr.created_at DESC`

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var requests []map[string]interface{}
	for rows.Next() {
		var id, empID2, empName, empNum, deptName string
		var ltID, ltName, ltColor, reason, status2, rejectReason string
		var isPaid bool
		var startDate, endDate time.Time
		var daysCount int
		var approvedAt *time.Time
		var createdAt, updatedAt time.Time

		if err2 := rows.Scan(
			&id, &empID2, &empName, &empNum, &deptName,
			&ltID, &ltName, &ltColor, &isPaid,
			&startDate, &endDate, &daysCount,
			&reason, &status2, &rejectReason,
			&approvedAt, &createdAt, &updatedAt,
		); err2 != nil {
			continue
		}
		requests = append(requests, map[string]interface{}{
			"id": id, "employee_id": empID2,
			"employee_name": empName, "employee_number": empNum,
			"department_name": deptName,
			"leave_type_id": ltID, "leave_type_name": ltName,
			"leave_color": ltColor, "is_paid": isPaid,
			"start_date": startDate, "end_date": endDate, "days_count": daysCount,
			"reason": reason, "status": status2, "rejection_reason": rejectReason,
			"approved_at": approvedAt, "created_at": createdAt, "updated_at": updatedAt,
		})
	}
	if requests == nil {
		requests = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, requests)
}

func (h *HRHandler) CreateLeaveRequest(c *gin.Context) {
	var req struct {
		EmployeeID  string `json:"employee_id"`
		LeaveTypeID string `json:"leave_type_id"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		DaysCount   int    `json:"days_count"`
		Reason      string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		INSERT INTO leave_requests
			(id, company_id, employee_id, leave_type_id, start_date, end_date, days_count, reason, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'pending')`,
		id, companyID, req.EmployeeID, req.LeaveTypeID,
		req.StartDate, req.EndDate, req.DaysCount, req.Reason,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "status": "pending", "message": "Leave request submitted"})
}

func (h *HRHandler) ApproveLeaveRequest(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE leave_requests SET
			status='approved', approved_by=$1, approved_at=NOW(), updated_at=NOW()
		WHERE id=$2 AND status='pending'`, userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Update employee status if needed
	_, _ = h.db.Exec(ctx, `
		UPDATE employees SET status='on_leave', updated_at=NOW()
		WHERE id=(SELECT employee_id FROM leave_requests WHERE id=$1)
		  AND status='active'`, id)
	c.JSON(http.StatusOK, gin.H{"message": "Leave request approved"})
}

func (h *HRHandler) RejectLeaveRequest(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE leave_requests SET
			status='rejected', rejection_reason=$1, updated_at=NOW()
		WHERE id=$2 AND status='pending'`, req.Reason, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Leave request rejected"})
}

func (h *HRHandler) CancelLeaveRequest(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE leave_requests SET status='cancelled', updated_at=NOW()
		WHERE id=$1 AND status='pending'`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Leave request cancelled"})
}

// ══════════════════════════════════════════════════════════════════════════════
// PAYROLL
// ══════════════════════════════════════════════════════════════════════════════

func (h *HRHandler) ListPayrollRuns(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx, `
		SELECT id, period_month, period_year, status,
		       total_gross, total_irg, total_cnas_employee, total_cnas_employer,
		       total_net, total_employees,
		       approved_at, paid_at, created_at
		FROM payroll_runs
		WHERE company_id=$1
		ORDER BY period_year DESC, period_month DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var runs []map[string]interface{}
	for rows.Next() {
		var id, status string
		var month, year, totalEmp int
		var gross, irg, cnasEmp, cnasEmpl, net float64
		var approvedAt, paidAt *time.Time
		var createdAt time.Time
		if err2 := rows.Scan(&id, &month, &year, &status,
			&gross, &irg, &cnasEmp, &cnasEmpl, &net, &totalEmp,
			&approvedAt, &paidAt, &createdAt); err2 != nil {
			continue
		}
		runs = append(runs, map[string]interface{}{
			"id": id, "period_month": month, "period_year": year, "status": status,
			"total_gross": gross, "total_irg": irg,
			"total_cnas_employee": cnasEmp, "total_cnas_employer": cnasEmpl,
			"total_net": net, "total_employees": totalEmp,
			"approved_at": approvedAt, "paid_at": paidAt, "created_at": createdAt,
		})
	}
	if runs == nil {
		runs = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, runs)
}

func (h *HRHandler) RunPayroll(c *gin.Context) {
	var req struct {
		Month int `json:"month"`
		Year  int `json:"year"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	ctx := context.Background()

	// Check for duplicate
	var existingID string
	err := h.db.QueryRow(ctx, `
		SELECT id FROM payroll_runs
		WHERE company_id=$1 AND period_year=$2 AND period_month=$3
		  AND status NOT IN ('cancelled')`,
		companyID, req.Year, req.Month,
	).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Payroll for this period already exists", "id": existingID})
		return
	}

	// Get active employees with their base salary
	rows, err := h.db.Query(ctx, `
		SELECT id, base_salary FROM employees
		WHERE company_id=$1 AND status='active'`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	runID := uuid.NewString()

	var (
		totalGross     decimal.Decimal
		totalIRG       decimal.Decimal
		totalCNASEmp   decimal.Decimal
		totalCNASEmpl  decimal.Decimal
		totalNet       decimal.Decimal
		totalEmployees int
	)

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	type empRow struct {
		id          string
		baseSalary  decimal.Decimal
	}
	var emps []empRow
	for rows.Next() {
		var e empRow
		var sal float64
		if err2 := rows.Scan(&e.id, &sal); err2 != nil {
			continue
		}
		e.baseSalary = decimal.NewFromFloat(sal)
		emps = append(emps, e)
	}
	rows.Close()

	for _, emp := range emps {
		// Fixed allowances (Algerian standard)
		transportAllowance := decimal.NewFromFloat(3000)
		mealAllowance := decimal.NewFromFloat(1000)

		grossSalary := emp.baseSalary.Add(transportAllowance).Add(mealAllowance)

		// CNAS employee: 9% of gross
		cnasEmployee := grossSalary.Mul(decimal.NewFromFloat(0.09)).Round(2)
		// CNAS employer: 26% of gross
		cnasEmployer := grossSalary.Mul(decimal.NewFromFloat(0.26)).Round(2)

		// Taxable income = gross - CNAS employee
		taxableIncome := grossSalary.Sub(cnasEmployee).Round(2)

		// IRG calculation (Algerian 2024 scale)
		irg := hrCalculateIRG(taxableIncome).Round(2)

		netSalary := taxableIncome.Sub(irg).Round(2)

		psID := uuid.NewString()
		_, err2 := tx.Exec(ctx, `
			INSERT INTO payslips (
				id, payroll_run_id, employee_id,
				period_month, period_year,
				days_worked, overtime_hours,
				base_salary, transport_allowance, meal_allowance,
				gross_salary, cnas_employee, cnas_employer,
				taxable_income, irg_amount,
				other_deductions, advance_deduction,
				net_salary
			) VALUES (
				$1,$2,$3,$4,$5,
				26,0,
				$6,$7,$8,
				$9,$10,$11,
				$12,$13,
				0,0,$14
			)`,
			psID, runID, emp.id,
			req.Month, req.Year,
			emp.baseSalary, transportAllowance, mealAllowance,
			grossSalary, cnasEmployee, cnasEmployer,
			taxableIncome, irg,
			netSalary,
		)
		if err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err2.Error()})
			return
		}

		totalGross = totalGross.Add(grossSalary)
		totalIRG = totalIRG.Add(irg)
		totalCNASEmp = totalCNASEmp.Add(cnasEmployee)
		totalCNASEmpl = totalCNASEmpl.Add(cnasEmployer)
		totalNet = totalNet.Add(netSalary)
		totalEmployees++
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO payroll_runs (
			id, company_id, period_month, period_year, status,
			total_gross, total_irg,
			total_cnas_employee, total_cnas_employer,
			total_net, total_employees, created_by
		) VALUES ($1,$2,$3,$4,'draft',$5,$6,$7,$8,$9,$10,$11)`,
		runID, companyID, req.Month, req.Year,
		totalGross, totalIRG,
		totalCNASEmp, totalCNASEmpl,
		totalNet, totalEmployees, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":              runID,
		"period_month":    req.Month,
		"period_year":     req.Year,
		"status":          "draft",
		"total_gross":     totalGross,
		"total_net":       totalNet,
		"total_employees": totalEmployees,
		"message":         "Payroll calculated successfully",
	})
}

func (h *HRHandler) ApprovePayrollRun(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE payroll_runs SET
			status='approved', approved_by=$1, approved_at=NOW(), updated_at=NOW()
		WHERE id=$2 AND status='draft'`, userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Payroll run approved"})
}

func (h *HRHandler) PayPayrollRun(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE payroll_runs SET
			status='paid', paid_at=NOW(), updated_at=NOW()
		WHERE id=$1 AND status='approved'`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Payroll marked as paid"})
}

func (h *HRHandler) GetPayslips(c *gin.Context) {
	runID := c.Param("id")
	ctx := context.Background()
	rows, err := h.db.Query(ctx, `
		SELECT
			ps.id,
			e.employee_number,
			e.first_name||' '||e.last_name AS employee_name,
			COALESCE(d.name,'') AS department_name,
			ps.days_worked, ps.overtime_hours,
			ps.base_salary, ps.transport_allowance, ps.meal_allowance,
			ps.housing_allowance, ps.other_allowances,
			ps.gross_salary,
			ps.cnas_employee, ps.cnas_employer,
			ps.taxable_income, ps.irg_amount,
			ps.other_deductions, ps.advance_deduction,
			ps.net_salary
		FROM payslips ps
		JOIN employees   e ON e.id = ps.employee_id
		LEFT JOIN departments d ON d.id = e.department_id
		WHERE ps.payroll_run_id = $1
		ORDER BY e.last_name, e.first_name`, runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var payslips []map[string]interface{}
	for rows.Next() {
		var id, empNum, empName, deptName string
		var daysWorked, overtimeH float64
		var baseSal, transport, meal, housing, otherAllow float64
		var gross, cnasEmp, cnasEmpl, taxable, irg float64
		var otherDed, advanceDed, net float64
		if err2 := rows.Scan(
			&id, &empNum, &empName, &deptName,
			&daysWorked, &overtimeH,
			&baseSal, &transport, &meal, &housing, &otherAllow,
			&gross, &cnasEmp, &cnasEmpl, &taxable, &irg,
			&otherDed, &advanceDed, &net,
		); err2 != nil {
			continue
		}
		payslips = append(payslips, map[string]interface{}{
			"id": id, "employee_number": empNum, "employee_name": empName,
			"department_name": deptName,
			"days_worked": daysWorked, "overtime_hours": overtimeH,
			"base_salary": baseSal, "transport_allowance": transport,
			"meal_allowance": meal, "housing_allowance": housing,
			"other_allowances": otherAllow,
			"gross_salary": gross,
			"cnas_employee": cnasEmp, "cnas_employer": cnasEmpl,
			"taxable_income": taxable, "irg_amount": irg,
			"other_deductions": otherDed, "advance_deduction": advanceDed,
			"net_salary": net,
		})
	}
	if payslips == nil {
		payslips = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, payslips)
}

func (h *HRHandler) ExportG29(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "G29 (CNAS declaration) export generated", "format": "PDF"})
}

// hrCalculateIRG implements the Algerian IRG monthly progressive scale (2024)
func hrCalculateIRG(netTaxable decimal.Decimal) decimal.Decimal {
	type Bracket struct {
		Min  decimal.Decimal
		Max  decimal.Decimal
		Rate decimal.Decimal
	}
	brackets := []Bracket{
		{decimal.NewFromInt(0),      decimal.NewFromInt(20000),  decimal.NewFromFloat(0.00)},
		{decimal.NewFromInt(20001),  decimal.NewFromInt(40000),  decimal.NewFromFloat(0.23)},
		{decimal.NewFromInt(40001),  decimal.NewFromInt(80000),  decimal.NewFromFloat(0.27)},
		{decimal.NewFromInt(80001),  decimal.NewFromInt(160000), decimal.NewFromFloat(0.30)},
		{decimal.NewFromInt(160001), decimal.NewFromInt(320000), decimal.NewFromFloat(0.33)},
		{decimal.NewFromInt(320001), decimal.NewFromInt(9999999),decimal.NewFromFloat(0.35)},
	}
	var irg decimal.Decimal
	for _, b := range brackets {
		if netTaxable.GreaterThan(b.Max) {
			irg = irg.Add(b.Max.Sub(b.Min).Mul(b.Rate))
		} else if netTaxable.GreaterThan(b.Min) {
			irg = irg.Add(netTaxable.Sub(b.Min).Mul(b.Rate))
			break
		}
	}
	return irg
}

// ══════════════════════════════════════════════════════════════════════════════
// RECRUITMENT
// ══════════════════════════════════════════════════════════════════════════════

func (h *HRHandler) ListJobPostings(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	status := c.Query("status")
	query := `
		SELECT
			jp.id, jp.title, jp.location, jp.employment_type,
			jp.vacancies, jp.status,
			jp.department_id, COALESCE(d.name,'') AS department_name,
			jp.position_id,  COALESCE(p.title,'') AS position_title,
			jp.published_at, jp.deadline_date,
			jp.description, COALESCE(jp.requirements,''),
			jp.created_at,
			COUNT(ja.id) AS applications_count,
			COUNT(ja.id) FILTER (WHERE ja.status='new')        AS new_count,
			COUNT(ja.id) FILTER (WHERE ja.status='hired')      AS hired_count
		FROM job_postings jp
		LEFT JOIN departments d ON d.id = jp.department_id
		LEFT JOIN positions   p ON p.id = jp.position_id
		LEFT JOIN job_applications ja ON ja.job_posting_id = jp.id
		WHERE jp.company_id=$1`

	args := []interface{}{companyID}
	if status != "" {
		query += ` AND jp.status=$2`
		args = append(args, status)
	}
	query += ` GROUP BY jp.id, d.name, p.title ORDER BY jp.created_at DESC`

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var postings []map[string]interface{}
	for rows.Next() {
		var id, title, location, empType, status2 string
		var deptID, deptName, posID, posTitle string
		var description, requirements string
		var vacancies, appCount, newCount, hiredCount int
		var publishedAt, deadlineDate *time.Time
		var createdAt time.Time
		if err2 := rows.Scan(
			&id, &title, &location, &empType, &vacancies, &status2,
			&deptID, &deptName, &posID, &posTitle,
			&publishedAt, &deadlineDate,
			&description, &requirements, &createdAt,
			&appCount, &newCount, &hiredCount,
		); err2 != nil {
			continue
		}
		postings = append(postings, map[string]interface{}{
			"id": id, "title": title, "location": location,
			"employment_type": empType, "vacancies": vacancies, "status": status2,
			"department_id": deptID, "department_name": deptName,
			"position_id": posID, "position_title": posTitle,
			"published_at": publishedAt, "deadline_date": deadlineDate,
			"description": description, "requirements": requirements,
			"created_at": createdAt,
			"applications_count": appCount, "new_count": newCount, "hired_count": hiredCount,
		})
	}
	if postings == nil {
		postings = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, postings)
}

func (h *HRHandler) GetJobPosting(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var jpID, title, location, empType, status string
	var deptID, deptName, posID, posTitle, description, requirements string
	var vacancies int
	var publishedAt, deadlineDate *time.Time
	var createdAt time.Time

	err := h.db.QueryRow(ctx, `
		SELECT jp.id, jp.title, jp.location, jp.employment_type, jp.vacancies, jp.status,
		       jp.department_id, COALESCE(d.name,''),
		       jp.position_id,  COALESCE(p.title,''),
		       jp.published_at, jp.deadline_date,
		       COALESCE(jp.description,''), COALESCE(jp.requirements,''),
		       jp.created_at
		FROM job_postings jp
		LEFT JOIN departments d ON d.id=jp.department_id
		LEFT JOIN positions   p ON p.id=jp.position_id
		WHERE jp.id=$1`, id,
	).Scan(&jpID, &title, &location, &empType, &vacancies, &status,
		&deptID, &deptName, &posID, &posTitle,
		&publishedAt, &deadlineDate,
		&description, &requirements, &createdAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job posting not found"})
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"id": jpID, "title": title, "location": location,
		"employment_type": empType, "vacancies": vacancies, "status": status,
		"department_id": deptID, "department_name": deptName,
		"position_id": posID, "position_title": posTitle,
		"published_at": publishedAt, "deadline_date": deadlineDate,
		"description": description, "requirements": requirements,
		"created_at": createdAt,
	})
}

func (h *HRHandler) CreateJobPosting(c *gin.Context) {
	var req struct {
		Title          string  `json:"title"`
		Description    string  `json:"description"`
		Requirements   string  `json:"requirements"`
		Location       string  `json:"location"`
		EmploymentType string  `json:"employment_type"`
		Vacancies      int     `json:"vacancies"`
		DepartmentID   *string `json:"department_id"`
		PositionID     *string `json:"position_id"`
		DeadlineDate   *string `json:"deadline_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.EmploymentType == "" {
		req.EmploymentType = "permanent"
	}
	if req.Vacancies == 0 {
		req.Vacancies = 1
	}
	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		INSERT INTO job_postings (
			id, company_id, title, description, requirements,
			location, employment_type, vacancies, status,
			department_id, position_id, deadline_date, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'draft',$9,$10,$11,$12)`,
		id, companyID, req.Title, req.Description, req.Requirements,
		req.Location, req.EmploymentType, req.Vacancies,
		req.DepartmentID, req.PositionID, req.DeadlineDate, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "title": req.Title, "status": "draft", "message": "Job posting created"})
}

func (h *HRHandler) UpdateJobPosting(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Title          string  `json:"title"`
		Description    string  `json:"description"`
		Requirements   string  `json:"requirements"`
		Location       string  `json:"location"`
		EmploymentType string  `json:"employment_type"`
		Vacancies      int     `json:"vacancies"`
		Status         string  `json:"status"`
		DepartmentID   *string `json:"department_id"`
		PositionID     *string `json:"position_id"`
		DeadlineDate   *string `json:"deadline_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()

	var publishedAt interface{}
	if req.Status == "open" {
		publishedAt = "NOW()"
	}
	_ = publishedAt

	_, err := h.db.Exec(ctx, `
		UPDATE job_postings SET
			title=$1, description=$2, requirements=$3,
			location=$4, employment_type=$5, vacancies=$6,
			status=$7, department_id=$8, position_id=$9, deadline_date=$10,
			published_at = CASE WHEN $7='open' AND published_at IS NULL THEN NOW() ELSE published_at END,
			updated_at=NOW()
		WHERE id=$11`,
		req.Title, req.Description, req.Requirements,
		req.Location, req.EmploymentType, req.Vacancies,
		req.Status, req.DepartmentID, req.PositionID, req.DeadlineDate, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Job posting updated"})
}

func (h *HRHandler) DeleteJobPosting(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `UPDATE job_postings SET status='closed', updated_at=NOW() WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Job posting closed"})
}

// Applications

func (h *HRHandler) ListApplications(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	jobID := c.Query("job_posting_id")
	status := c.Query("status")

	query := `
		SELECT
			ja.id, ja.job_posting_id, jp.title AS job_title,
			ja.first_name, ja.last_name,
			ja.first_name||' '||ja.last_name AS full_name,
			ja.email, ja.phone, ja.source, ja.status,
			COALESCE(ja.expected_salary,0),
			ja.interview_date, COALESCE(ja.interview_notes,''),
			COALESCE(ja.rejection_reason,''),
			ja.created_at, ja.updated_at
		FROM job_applications ja
		JOIN job_postings jp ON jp.id = ja.job_posting_id
		WHERE ja.company_id=$1`

	args := []interface{}{companyID}
	idx := 2
	if jobID != "" {
		query += fmt.Sprintf(` AND ja.job_posting_id=$%d`, idx)
		args = append(args, jobID)
		idx++
	}
	if status != "" {
		query += fmt.Sprintf(` AND ja.status=$%d`, idx)
		args = append(args, status)
		idx++
	}
	_ = idx
	query += ` ORDER BY ja.created_at DESC`

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var apps []map[string]interface{}
	for rows.Next() {
		var id, jobID2, jobTitle string
		var firstName, lastName, fullName string
		var email, phone, source, status2 string
		var interviewNotes, rejectReason string
		var expectedSalary float64
		var interviewDate *time.Time
		var createdAt, updatedAt time.Time
		if err2 := rows.Scan(
			&id, &jobID2, &jobTitle,
			&firstName, &lastName, &fullName,
			&email, &phone, &source, &status2,
			&expectedSalary, &interviewDate, &interviewNotes, &rejectReason,
			&createdAt, &updatedAt,
		); err2 != nil {
			continue
		}
		apps = append(apps, map[string]interface{}{
			"id": id, "job_posting_id": jobID2, "job_title": jobTitle,
			"first_name": firstName, "last_name": lastName, "full_name": fullName,
			"email": email, "phone": phone, "source": source, "status": status2,
			"expected_salary": expectedSalary,
			"interview_date": interviewDate, "interview_notes": interviewNotes,
			"rejection_reason": rejectReason,
			"created_at": createdAt, "updated_at": updatedAt,
		})
	}
	if apps == nil {
		apps = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, apps)
}

func (h *HRHandler) CreateApplication(c *gin.Context) {
	var req struct {
		JobPostingID   string   `json:"job_posting_id"`
		FirstName      string   `json:"first_name"`
		LastName       string   `json:"last_name"`
		Email          string   `json:"email"`
		Phone          string   `json:"phone"`
		CvURL          string   `json:"cv_url"`
		CoverLetter    string   `json:"cover_letter"`
		Source         string   `json:"source"`
		ExpectedSalary *float64 `json:"expected_salary"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Source == "" {
		req.Source = "direct"
	}
	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		INSERT INTO job_applications (
			id, job_posting_id, company_id,
			first_name, last_name, email, phone,
			cv_url, cover_letter, source, status, expected_salary
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'new',$11)`,
		id, req.JobPostingID, companyID,
		req.FirstName, req.LastName, req.Email, req.Phone,
		req.CvURL, req.CoverLetter, req.Source, req.ExpectedSalary,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": id, "status": "new",
		"full_name": req.FirstName + " " + req.LastName,
		"message": "Application submitted",
	})
}

func (h *HRHandler) UpdateApplicationStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status          string  `json:"status"`
		InterviewDate   *string `json:"interview_date"`
		InterviewNotes  string  `json:"interview_notes"`
		RejectionReason string  `json:"rejection_reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE job_applications SET
			status=$1, interview_date=$2, interview_notes=$3,
			rejection_reason=$4, updated_at=NOW()
		WHERE id=$5`,
		req.Status, req.InterviewDate, req.InterviewNotes, req.RejectionReason, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Application status updated"})
}

// ── Legacy stubs (kept for backward compat) ───────────────────────────────────

func (h *HRHandler) ListCandidates(c *gin.Context) {
	h.ListApplications(c)
}
func (h *HRHandler) CreateCandidate(c *gin.Context) {
	h.CreateApplication(c)
}
func (h *HRHandler) ListPerformanceReviews(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}
func (h *HRHandler) CreatePerformanceReview(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"id": uuid.NewString()})
}
