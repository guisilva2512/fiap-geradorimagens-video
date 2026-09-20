package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AuthenticatedUsernameKey = "authenticated_username"
	AuthenticatedRoleKey     = "authenticated_role"
)

type authClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := strings.Fields(c.GetHeader("Authorization"))
		if len(authorization) != 2 || !strings.EqualFold(authorization[0], "Bearer") {
			unauthorized(c)
			return
		}

		secretKey := os.Getenv("JWT_SECRET")
		issuer := os.Getenv("JWT_ISSUER")
		if secretKey == "" || issuer == "" {
			unauthorized(c)
			return
		}

		claims := &authClaims{}
		token, err := jwt.ParseWithClaims(authorization[1], claims, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secretKey), nil
		}, jwt.WithIssuer(issuer))
		if err != nil || !token.Valid {
			unauthorized(c)
			return
		}

		c.Set(AuthenticatedUsernameKey, claims.Username)
		c.Set(AuthenticatedRoleKey, claims.Role)
		c.Next()
	}
}

func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "não autorizado"})
}
