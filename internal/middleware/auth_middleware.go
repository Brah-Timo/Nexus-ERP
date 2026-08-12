package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"os"
)

type Claims struct {
	UserID    string `json:"sub"`
	Role      string `json:"role"`
	CompanyID string `json:"company_id"`
	BranchID  string `json:"branch_id"`
	TenantID  string `json:"tenant_id"`
	jwt.RegisteredClaims
}

// JWTAuth middleware validates bearer tokens
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret(), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("company_id", claims.CompanyID)
		c.Set("branch_id", claims.BranchID)
		c.Set("tenant_id", claims.TenantID)
		c.Next()
	}
}

// RequireRole enforces role-based access
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		roleStr, _ := role.(string)

		for _, r := range roles {
			if r == roleStr || roleStr == "admin" || roleStr == "superadmin" {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
	}
}

// AuditLog middleware records every mutating request
func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		// TODO: record to audit_log table after handler executes
	}
}

func jwtSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "nexus-erp-default-secret-change-in-production"
	}
	return []byte(secret)
}

// GetUserID extracts the user ID from context
func GetUserID(c *gin.Context) string {
	v, _ := c.Get("user_id")
	s, _ := v.(string)
	return s
}

// GetCompanyID extracts the company ID from context
func GetCompanyID(c *gin.Context) string {
	v, _ := c.Get("company_id")
	s, _ := v.(string)
	return s
}
