package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"nexus-erp/internal/middleware"
	"nexus-erp/internal/models"
)

// Ensure time and models are used (referenced elsewhere in file)
var _ = time.Now
var _ = models.Base{}

// ─── Handler Registry ──────────────────────────────────────────────────────────

type Handler struct {
	db           *pgxpool.Pool
	Auth         *AuthHandler
	Dashboard    *DashboardHandler
	Settings     *SettingsHandler
	Accounting   *AccountingHandler
	HR           *HRHandler
	Sales        *SalesHandler
	Purchase     *PurchaseHandler
	Inventory    *InventoryHandler
	Manufacturing *ManufacturingHandler
	Projects     *ProjectsHandler
	Treasury     *TreasuryHandler
	Tax          *TaxHandler
	Workflow     *WorkflowHandler
	Reports      *ReportsHandler
	ReportsBI    *ReportsBIHandler
	Diagnostics  *DiagnosticsHandler
	Maintenance  *MaintenanceHandler
	Fleet        *FleetHandler
	Quality      *QualityHandler
}

// NewHandler constructs all handlers with a shared DB pool
func NewHandler(db *pgxpool.Pool) *Handler {
	h := &Handler{db: db}
	h.Auth = &AuthHandler{db: db}
	h.Dashboard = &DashboardHandler{db: db}
	h.Settings = &SettingsHandler{db: db}
	h.Accounting = &AccountingHandler{db: db}
	h.HR = &HRHandler{db: db}
	h.Sales = &SalesHandler{db: db}
	h.Purchase = &PurchaseHandler{db: db}
	h.Inventory = &InventoryHandler{db: db}
	h.Manufacturing = &ManufacturingHandler{db: db}
	h.Projects = &ProjectsHandler{db: db}
	h.Treasury = &TreasuryHandler{db: db}
	h.Tax = &TaxHandler{db: db}
	h.Workflow = &WorkflowHandler{db: db}
	h.Reports = &ReportsHandler{db: db}
	h.ReportsBI = &ReportsBIHandler{db: db}
	h.Diagnostics = &DiagnosticsHandler{db: db}
	h.Maintenance = &MaintenanceHandler{db: db}
	h.Fleet = &FleetHandler{db: db}
	h.Quality = &QualityHandler{db: db}
	return h
}

// ─── Auth Handler ──────────────────────────────────────────────────────────────

type AuthHandler struct{ db *pgxpool.Pool }

type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	CompanyID string `json:"company_id"`
}

type LoginResponse struct {
	Token        string        `json:"token"`
	RefreshToken string        `json:"refresh_token"`
	User         *models.User  `json:"user"`
	ExpiresAt    time.Time     `json:"expires_at"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := context.Background()
	var user models.User
	err := h.db.QueryRow(ctx,
		`SELECT id, username, email, password_hash, full_name, role,
		        COALESCE(company_id::text, ''),
		        COALESCE(branch_id::text, ''),
		        COALESCE(tenant_id::text, ''),
		        is_active
		 FROM users WHERE username = $1 AND is_active = true`,
		req.Username,
	).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FullName, &user.Role, &user.CompanyID, &user.BranchID,
		&user.TenantID, &user.IsActive,
	)
	if err != nil {
		log.Printf("[LOGIN ERROR] DB scan failed for user '%s': %v", req.Username, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	expiresAt := time.Now().Add(8 * time.Hour)
	claims := &middleware.Claims{
		UserID:    user.ID,
		Role:      user.Role,
		CompanyID: user.CompanyID,
		BranchID:  user.BranchID,
		TenantID:  user.TenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwtKey())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Refresh token (longer expiry)
	refreshClaims := &middleware.Claims{
		UserID:    user.ID,
		Role:      user.Role,
		CompanyID: user.CompanyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			ID:        uuid.NewString(),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefresh, _ := refreshToken.SignedString(jwtKey())

	// Update last login
	_, _ = h.db.Exec(ctx, `UPDATE users SET last_login_at = NOW() WHERE id = $1`, user.ID)

	user.PasswordHash = ""
	c.JSON(http.StatusOK, LoginResponse{
		Token:        signed,
		RefreshToken: signedRefresh,
		User:         &user,
		ExpiresAt:    expiresAt,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// In a production system, blacklist the token here
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey(), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	expiresAt := time.Now().Add(8 * time.Hour)
	newClaims := &middleware.Claims{
		UserID:    claims.UserID,
		Role:      claims.Role,
		CompanyID: claims.CompanyID,
		BranchID:  claims.BranchID,
		TenantID:  claims.TenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	signed, _ := newToken.SignedString(jwtKey())

	c.JSON(http.StatusOK, gin.H{"token": signed, "expires_at": expiresAt})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid email required"})
		return
	}
	// TODO: send reset email
	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	// TODO: validate token and update password
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func jwtKey() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "nexus-erp-default-secret-change-in-production"
	}
	return []byte(s)
}

// DashboardHandler is defined in sales.go

// ─── Settings Handler ──────────────────────────────────────────────────────────

type SettingsHandler struct{ db *pgxpool.Pool }

// ── Companies ─────────────────────────────────────────────────────────────────

func (h *SettingsHandler) ListCompanies(c *gin.Context) {
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, name, COALESCE(nif,''), COALESCE(address,''), COALESCE(city,''),
		        COALESCE(phone,''), COALESCE(email,''), COALESCE(website,''),
		        COALESCE(logo_url,''), COALESCE(country,'DZ'), COALESCE(timezone,'Africa/Algiers'),
		        COALESCE(default_currency, currency, 'DZD'), is_active, created_at,
		        COALESCE(legal_name,''), COALESCE(nis,''), COALESCE(rc,''), COALESCE(art,''),
		        COALESCE(wilaya,''), COALESCE(postal_code,'')
		 FROM companies ORDER BY name`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, name, nif, address, city, phone, email, website, logoURL, country, tz, currency string
		var isActive bool
		var createdAt interface{}
		var legalName, nis, rc, art, wilaya, postalCode string
		_ = rows.Scan(&id, &name, &nif, &address, &city, &phone, &email,
			&website, &logoURL, &country, &tz, &currency, &isActive, &createdAt,
			&legalName, &nis, &rc, &art, &wilaya, &postalCode)
		list = append(list, map[string]interface{}{
			"id": id, "name": name, "nif": nif, "address": address,
			"city": city, "phone": phone, "email": email, "website": website,
			"logo_url": logoURL, "country": country, "timezone": tz,
			"default_currency": currency, "is_active": isActive, "created_at": createdAt,
			"legal_name": legalName, "nis": nis, "rc": rc, "art": art,
			"wilaya": wilaya, "postal_code": postalCode,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *SettingsHandler) CreateCompany(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := uuid.NewString()
	ctx := context.Background()
	// Get or create a default tenant_id
	var tenantID string
	_ = h.db.QueryRow(ctx, `SELECT id FROM tenants LIMIT 1`).Scan(&tenantID)
	if tenantID == "" {
		tenantID = uuid.NewString()
		_, _ = h.db.Exec(ctx,
			`INSERT INTO tenants (id, code, name) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
			tenantID, "default", "Default Tenant")
	}
	code := strVal(req, "code")
	if code == "" {
		code = fmt.Sprintf("CO%s", id[:8])
	}
	_, err := h.db.Exec(ctx,
		`INSERT INTO companies (id, tenant_id, code, name, nif, address, city, phone, email, website,
		  logo_url, country, timezone, currency, default_currency, fiscal_year_start, is_active)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,1,$16)`,
		id, tenantID, code,
		strVal(req, "name"), strVal(req, "nif"),
		strVal(req, "address"), strVal(req, "city"),
		strVal(req, "phone"), strVal(req, "email"),
		strVal(req, "website"), strVal(req, "logo_url"),
		strValDefault(req, "country", "DZ"),
		strValDefault(req, "timezone", "Africa/Algiers"),
		strValDefault(req, "currency", "DZD"),
		strValDefault(req, "default_currency", "DZD"),
		true,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req["id"] = id
	c.JSON(http.StatusCreated, req)
}

func (h *SettingsHandler) UpdateCompany(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx,
		`UPDATE companies SET name=$1, nif=$2, address=$3, city=$4, phone=$5, email=$6,
		  website=$7, logo_url=$8, country=$9, timezone=$10, default_currency=$11, updated_at=NOW()
		 WHERE id=$12`,
		strVal(req, "name"), strVal(req, "nif"), strVal(req, "address"),
		strVal(req, "city"), strVal(req, "phone"), strVal(req, "email"),
		strVal(req, "website"), strVal(req, "logo_url"), strVal(req, "country"),
		strVal(req, "timezone"), strVal(req, "default_currency"), id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req["id"] = id
	c.JSON(http.StatusOK, req)
}

// ── Users ─────────────────────────────────────────────────────────────────────

func (h *SettingsHandler) ListUsers(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT u.id, u.username, u.email,
		        COALESCE(u.full_name,''), COALESCE(u.role::text,''),
		        COALESCE(u.role_id::text,''), COALESCE(r.name,''),
		        COALESCE(u.phone,''), COALESCE(u.avatar_url,''),
		        COALESCE(u.is_active, true), u.last_login_at, u.created_at
		 FROM users u
		 LEFT JOIN roles r ON r.id = u.role_id
		 WHERE u.company_id = $1
		 ORDER BY u.full_name NULLS LAST`,
		companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, username, email, fullName, roleStr, roleID, roleName, phone, avatarURL string
		var isActive bool
		var lastLogin, createdAt interface{}
		_ = rows.Scan(&id, &username, &email, &fullName, &roleStr, &roleID, &roleName,
			&phone, &avatarURL, &isActive, &lastLogin, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "username": username, "email": email,
			"full_name": fullName, "role": roleStr,
			"role_id": roleID, "role_name": roleName,
			"phone": phone, "avatar_url": avatarURL,
			"is_active": isActive, "last_login": lastLogin, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *SettingsHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FullName  string `json:"full_name"`
		Role      string `json:"role"`
		RoleID    string `json:"role_id"`
		Phone     string `json:"phone"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var roleIDVal interface{}
	if req.RoleID != "" {
		roleIDVal = req.RoleID
	}

	// Get tenant_id from company
	var tenantID2 string
	_ = h.db.QueryRow(ctx, `SELECT tenant_id FROM companies WHERE id=$1`, companyID).Scan(&tenantID2)
	if tenantID2 == "" {
		_ = h.db.QueryRow(ctx, `SELECT id FROM tenants LIMIT 1`).Scan(&tenantID2)
	}
	_, err = h.db.Exec(ctx,
		`INSERT INTO users (id, tenant_id, company_id, username, email, password_hash, full_name, role, role_id, phone, avatar_url, is_active)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		id, tenantID2, companyID, req.Username, req.Email, string(hash),
		req.FullName, req.Role, roleIDVal, req.Phone, req.AvatarURL, true,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": id, "username": req.Username, "email": req.Email,
		"full_name": req.FullName, "role": req.Role, "role_id": req.RoleID,
		"phone": req.Phone, "is_active": true,
	})
}

func (h *SettingsHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Email     string `json:"email"`
		FullName  string `json:"full_name"`
		Role      string `json:"role"`
		RoleID    string `json:"role_id"`
		Phone     string `json:"phone"`
		AvatarURL string `json:"avatar_url"`
		IsActive  *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	var roleIDVal interface{}
	if req.RoleID != "" {
		roleIDVal = req.RoleID
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	_, err := h.db.Exec(ctx,
		`UPDATE users SET email=$1, full_name=$2, role=$3, role_id=$4, phone=$5, avatar_url=$6,
		  is_active=$7, updated_at=NOW() WHERE id=$8`,
		req.Email, req.FullName, req.Role, roleIDVal, req.Phone, req.AvatarURL, isActive, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "User updated"})
}

func (h *SettingsHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `UPDATE users SET is_active=false, updated_at=NOW() WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deactivated"})
}

// ── Roles ─────────────────────────────────────────────────────────────────────

func (h *SettingsHandler) ListRoles(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, name, COALESCE(description,''), permissions, is_system, is_active, created_at
		 FROM roles WHERE company_id=$1 ORDER BY name`,
		companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, name, desc string
		var perms interface{}
		var isSystem, isActive bool
		var createdAt interface{}
		_ = rows.Scan(&id, &name, &desc, &perms, &isSystem, &isActive, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "name": name, "description": desc,
			"permissions": perms, "is_system": isSystem,
			"is_active": isActive, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *SettingsHandler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Permissions []interface{} `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	if req.Permissions == nil {
		req.Permissions = []interface{}{}
	}
	_, err := h.db.Exec(ctx,
		`INSERT INTO roles (id, company_id, name, description, permissions, is_system, is_active)
		 VALUES ($1,$2,$3,$4,$5,false,true)`,
		id, companyID, req.Name, req.Description, req.Permissions,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": id, "name": req.Name, "description": req.Description,
		"permissions": req.Permissions, "is_system": false, "is_active": true,
	})
}

func (h *SettingsHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Permissions []interface{} `json:"permissions"`
		IsActive    *bool         `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.Permissions == nil {
		req.Permissions = []interface{}{}
	}
	_, err := h.db.Exec(ctx,
		`UPDATE roles SET name=$1, description=$2, permissions=$3, is_active=$4, updated_at=NOW()
		 WHERE id=$5 AND is_system=false`,
		req.Name, req.Description, req.Permissions, isActive, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Role updated"})
}

func (h *SettingsHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `DELETE FROM roles WHERE id=$1 AND is_system=false`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role deleted"})
}

// ── Fiscal Years ──────────────────────────────────────────────────────────────

func (h *SettingsHandler) ListFiscalYears(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, name, start_date, end_date, status, is_current,
		        closed_at, created_at
		 FROM fiscal_years WHERE company_id=$1 ORDER BY start_date DESC`,
		companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, name, status string
		var isCurrent bool
		var startDate, endDate, closedAt, createdAt interface{}
		_ = rows.Scan(&id, &name, &startDate, &endDate, &status, &isCurrent, &closedAt, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "name": name, "start_date": startDate, "end_date": endDate,
			"status": status, "is_current": isCurrent,
			"closed_at": closedAt, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *SettingsHandler) CreateFiscalYear(c *gin.Context) {
	var req struct {
		Name      string `json:"name"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		IsCurrent bool   `json:"is_current"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	// If setting as current, unset others first
	if req.IsCurrent {
		_, _ = h.db.Exec(ctx,
			`UPDATE fiscal_years SET is_current=false WHERE company_id=$1`, companyID)
	}
	_, err := h.db.Exec(ctx,
		`INSERT INTO fiscal_years (id, company_id, name, start_date, end_date, is_closed, status, is_current)
		 VALUES ($1,$2,$3,$4,$5,false,'open',$6)`,
		id, companyID, req.Name, req.StartDate, req.EndDate, req.IsCurrent,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": id, "name": req.Name, "start_date": req.StartDate,
		"end_date": req.EndDate, "status": "open", "is_current": req.IsCurrent,
	})
}

func (h *SettingsHandler) CloseFiscalYear(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx,
		`UPDATE fiscal_years SET status='closed', is_closed=true, is_current=false, closed_at=NOW()
		 WHERE id=$1 AND company_id=$2`,
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Fiscal year closed"})
}

// ── Currencies ────────────────────────────────────────────────────────────────

func (h *SettingsHandler) ListCurrencies(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, code, name, symbol, exchange_rate, is_base, is_active, updated_at
		 FROM company_currencies WHERE company_id=$1 ORDER BY is_base DESC, code`,
		companyID,
	)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, code, name, symbol string
		var rate float64
		var isBase, isActive bool
		var updatedAt interface{}
		_ = rows.Scan(&id, &code, &name, &symbol, &rate, &isBase, &isActive, &updatedAt)
		list = append(list, map[string]interface{}{
			"id": id, "code": code, "name": name, "symbol": symbol,
			"exchange_rate": rate, "is_base": isBase, "is_active": isActive,
			"updated_at": updatedAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *SettingsHandler) CreateCurrency(c *gin.Context) {
	var req struct {
		Code         string  `json:"code"`
		Name         string  `json:"name"`
		Symbol       string  `json:"symbol"`
		ExchangeRate float64 `json:"exchange_rate"`
		IsBase       bool    `json:"is_base"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	id := uuid.NewString()
	_, err := h.db.Exec(ctx,
		`INSERT INTO company_currencies (id, company_id, code, name, symbol, exchange_rate, is_base, is_active)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,true)`,
		id, companyID, req.Code, req.Name, req.Symbol, req.ExchangeRate, req.IsBase,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": id, "code": req.Code, "name": req.Name, "symbol": req.Symbol,
		"exchange_rate": req.ExchangeRate, "is_base": req.IsBase, "is_active": true,
	})
}

func (h *SettingsHandler) UpdateExchangeRate(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	var req struct {
		ExchangeRate float64 `json:"exchange_rate"`
		Name         string  `json:"name"`
		Symbol       string  `json:"symbol"`
		IsActive     *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	_, err := h.db.Exec(ctx,
		`UPDATE company_currencies SET exchange_rate=$1, name=$2, symbol=$3, is_active=$4, updated_at=NOW()
		 WHERE id=$5 AND company_id=$6`,
		req.ExchangeRate, req.Name, req.Symbol, isActive, id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Currency updated"})
}

func (h *SettingsHandler) DeleteCurrency(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx,
		`DELETE FROM company_currencies WHERE id=$1 AND company_id=$2 AND is_base=false`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Currency deleted"})
}

// ── Numbering Config ──────────────────────────────────────────────────────────

func (h *SettingsHandler) GetNumberingConfig(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, doc_type, prefix, suffix, next_number, padding, reset_yearly
		 FROM numbering_config WHERE company_id=$1 ORDER BY doc_type`,
		companyID,
	)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, docType, prefix, suffix string
		var nextNum, padding int
		var resetYearly bool
		_ = rows.Scan(&id, &docType, &prefix, &suffix, &nextNum, &padding, &resetYearly)
		list = append(list, map[string]interface{}{
			"id": id, "doc_type": docType, "prefix": prefix, "suffix": suffix,
			"next_number": nextNum, "padding": padding, "reset_yearly": resetYearly,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *SettingsHandler) UpdateNumberingConfig(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var items []struct {
		DocType     string `json:"doc_type"`
		Prefix      string `json:"prefix"`
		Suffix      string `json:"suffix"`
		NextNumber  int    `json:"next_number"`
		Padding     int    `json:"padding"`
		ResetYearly bool   `json:"reset_yearly"`
	}
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	for _, item := range items {
		if item.Padding == 0 {
			item.Padding = 4
		}
		_, err := h.db.Exec(ctx,
			`INSERT INTO numbering_config (id, company_id, doc_type, prefix, suffix, next_number, padding, reset_yearly)
			 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
			 ON CONFLICT (company_id, doc_type)
			 DO UPDATE SET prefix=$3, suffix=$4, next_number=$5, padding=$6, reset_yearly=$7`,
			companyID, item.DocType, item.Prefix, item.Suffix,
			item.NextNumber, item.Padding, item.ResetYearly,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Numbering config saved"})
}

// ── Taxes ─────────────────────────────────────────────────────────────────────

func (h *SettingsHandler) ListTaxes(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, name, code, tax_type, rate, is_active, created_at
		 FROM taxes WHERE company_id=$1 ORDER BY name`,
		companyID,
	)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, name, code, taxType string
		var rate float64
		var isActive bool
		var createdAt interface{}
		_ = rows.Scan(&id, &name, &code, &taxType, &rate, &isActive, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "name": name, "code": code, "tax_type": taxType,
			"rate": rate, "is_active": isActive, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *SettingsHandler) CreateTax(c *gin.Context) {
	var req struct {
		Name    string  `json:"name"`
		Code    string  `json:"code"`
		TaxType string  `json:"tax_type"`
		Rate    float64 `json:"rate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.TaxType == "" {
		req.TaxType = "percentage"
	}
	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx,
		`INSERT INTO taxes (id, company_id, name, code, tax_type, rate, is_active)
		 VALUES ($1,$2,$3,$4,$5,$6,true)`,
		id, companyID, req.Name, req.Code, req.TaxType, req.Rate,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": id, "name": req.Name, "code": req.Code,
		"tax_type": req.TaxType, "rate": req.Rate, "is_active": true,
	})
}

func (h *SettingsHandler) UpdateTax(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	var req struct {
		Name     string  `json:"name"`
		Code     string  `json:"code"`
		TaxType  string  `json:"tax_type"`
		Rate     float64 `json:"rate"`
		IsActive *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	_, err := h.db.Exec(ctx,
		`UPDATE taxes SET name=$1, code=$2, tax_type=$3, rate=$4, is_active=$5
		 WHERE id=$6 AND company_id=$7`,
		req.Name, req.Code, req.TaxType, req.Rate, isActive, id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Tax updated"})
}

func (h *SettingsHandler) DeleteTax(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `DELETE FROM taxes WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tax deleted"})
}

// ── Workflow Rules ────────────────────────────────────────────────────────────

func (h *SettingsHandler) ListWorkflowRules(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, name, doc_type, trigger_event, conditions, actions, is_active, priority, created_at
		 FROM workflow_rules WHERE company_id=$1 ORDER BY priority, name`,
		companyID,
	)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, name, docType, triggerEvent string
		var conditions, actions interface{}
		var isActive bool
		var priority int
		var createdAt interface{}
		_ = rows.Scan(&id, &name, &docType, &triggerEvent, &conditions, &actions, &isActive, &priority, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "name": name, "doc_type": docType, "trigger_event": triggerEvent,
			"conditions": conditions, "actions": actions,
			"is_active": isActive, "priority": priority, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *SettingsHandler) CreateWorkflowRule(c *gin.Context) {
	var req struct {
		Name         string        `json:"name"`
		DocType      string        `json:"doc_type"`
		TriggerEvent string        `json:"trigger_event"`
		Conditions   []interface{} `json:"conditions"`
		Actions      []interface{} `json:"actions"`
		Priority     int           `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.TriggerEvent == "" {
		req.TriggerEvent = "on_create"
	}
	if req.Priority == 0 {
		req.Priority = 10
	}
	if req.Conditions == nil {
		req.Conditions = []interface{}{}
	}
	if req.Actions == nil {
		req.Actions = []interface{}{}
	}
	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx,
		`INSERT INTO workflow_rules (id, company_id, name, doc_type, trigger_event, conditions, actions, is_active, priority)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,true,$8)`,
		id, companyID, req.Name, req.DocType, req.TriggerEvent,
		req.Conditions, req.Actions, req.Priority,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": id, "name": req.Name, "doc_type": req.DocType,
		"trigger_event": req.TriggerEvent, "priority": req.Priority,
		"conditions": req.Conditions, "actions": req.Actions, "is_active": true,
	})
}

func (h *SettingsHandler) UpdateWorkflowRule(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	var req struct {
		Name         string        `json:"name"`
		DocType      string        `json:"doc_type"`
		TriggerEvent string        `json:"trigger_event"`
		Conditions   []interface{} `json:"conditions"`
		Actions      []interface{} `json:"actions"`
		Priority     int           `json:"priority"`
		IsActive     *bool         `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.Conditions == nil {
		req.Conditions = []interface{}{}
	}
	if req.Actions == nil {
		req.Actions = []interface{}{}
	}
	_, err := h.db.Exec(ctx,
		`UPDATE workflow_rules SET name=$1, doc_type=$2, trigger_event=$3,
		  conditions=$4, actions=$5, priority=$6, is_active=$7, updated_at=NOW()
		 WHERE id=$8 AND company_id=$9`,
		req.Name, req.DocType, req.TriggerEvent, req.Conditions,
		req.Actions, req.Priority, isActive, id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Workflow rule updated"})
}

func (h *SettingsHandler) DeleteWorkflowRule(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx,
		`DELETE FROM workflow_rules WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Workflow rule deleted"})
}

// ── Audit Log ─────────────────────────────────────────────────────────────────

func (h *SettingsHandler) GetAuditLog(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")
	actionFilter := c.Query("action")
	entityFilter := c.Query("entity_type")
	search := c.Query("search")

	page := 1
	limit := 50
	if v, err := parseInt(pageStr); err == nil && v > 0 {
		page = v
	}
	if v, err := parseInt(limitStr); err == nil && v > 0 && v <= 200 {
		limit = v
	}
	offset := (page - 1) * limit

	// Count total
	countQuery := `SELECT COUNT(*) FROM audit_logs al
	  LEFT JOIN users u ON u.id = al.user_id
	  WHERE al.company_id=$1`
	countArgs := []interface{}{companyID}
	argIdx := 2

	if actionFilter != "" {
		countQuery += " AND al.action=$" + intToStr(argIdx)
		countArgs = append(countArgs, actionFilter)
		argIdx++
	}
	if entityFilter != "" {
		countQuery += " AND al.entity_type=$" + intToStr(argIdx)
		countArgs = append(countArgs, entityFilter)
		argIdx++
	}
	if search != "" {
		countQuery += " AND (al.entity_type ILIKE $" + intToStr(argIdx) + " OR al.action ILIKE $" + intToStr(argIdx) + ")"
		countArgs = append(countArgs, "%"+search+"%")
		argIdx++
	}

	var total int
	_ = h.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)

	// Data query
	dataQuery := `SELECT al.id, COALESCE(al.user_id::text,''), COALESCE(u.username,''), COALESCE(u.full_name,''),
	       al.action, al.entity_type, COALESCE(al.entity_id,''),
	       COALESCE(al.ip_address::text,''), COALESCE(al.user_agent,''), al.created_at
	  FROM audit_logs al
	  LEFT JOIN users u ON u.id = al.user_id
	  WHERE al.company_id=$1`

	dataArgs := []interface{}{companyID}
	argIdx = 2

	if actionFilter != "" {
		dataQuery += " AND al.action=$" + intToStr(argIdx)
		dataArgs = append(dataArgs, actionFilter)
		argIdx++
	}
	if entityFilter != "" {
		dataQuery += " AND al.entity_type=$" + intToStr(argIdx)
		dataArgs = append(dataArgs, entityFilter)
		argIdx++
	}
	if search != "" {
		dataQuery += " AND (al.entity_type ILIKE $" + intToStr(argIdx) + " OR al.action ILIKE $" + intToStr(argIdx) + ")"
		dataArgs = append(dataArgs, "%"+search+"%")
		argIdx++
	}
	dataQuery += " ORDER BY al.created_at DESC LIMIT $" + intToStr(argIdx) + " OFFSET $" + intToStr(argIdx+1)
	dataArgs = append(dataArgs, limit, offset)

	rows, err := h.db.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "total": 0, "page": page, "limit": limit})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var id, userID, username, fullName, action, entityType, entityID, ipAddr, userAgent string
		var createdAt interface{}
		_ = rows.Scan(&id, &userID, &username, &fullName, &action, &entityType,
			&entityID, &ipAddr, &userAgent, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "user_id": userID, "username": username, "full_name": fullName,
			"action": action, "entity_type": entityType, "entity_id": entityID,
			"ip_address": ipAddr, "user_agent": userAgent, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, gin.H{
		"data": list, "total": total, "page": page, "limit": limit,
	})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func strVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func strValDefault(m map[string]interface{}, key, def string) string {
	if v := strVal(m, key); v != "" {
		return v
	}
	return def
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func intToStr(n int) string {
	return fmt.Sprintf("%d", n)
}
