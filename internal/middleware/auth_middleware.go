package middleware

import (
	"slices"
	"strings"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/DevKayoS/fintech-kodify/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")

		if authHeader == "" {
			ctx.Error(errors.Unauthorized("authorization header required"))
			ctx.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.Error(errors.Unauthorized("invalid authorization format"))
			ctx.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.NewValidationError("unexpected signing method", jwt.ValidationErrorSignatureInvalid)
			}
			return models.SecretKey, nil
		})
		if err != nil {
			ctx.Error(errors.Unauthorized("access denied"))
			ctx.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if userID, ok := claims["user_id"].(float64); ok {
				ctx.Set("user_id", int64(userID))
			}
			ctx.Set("email", claims["email"])
			ctx.Set("role", claims["role"])
			ctx.Set("permissions", claims["permissions"])
			ctx.Set("claims", claims)
		}

		ctx.Next()
	}
}

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, exists := ctx.Get("role")
		if !exists {
			ctx.Error(errors.Forbidden("role not found"))
			ctx.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			ctx.Error(errors.Forbidden("invalid role format"))
			ctx.Abort()
			return
		}

		if slices.Contains(allowedRoles, userRole) {
			ctx.Next()
			return
		}

		ctx.Error(errors.Forbidden("insufficient permissions"))
		ctx.Abort()
	}
}

func RequirePermissions(requiredPermissions ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		permissionsInterface, exists := ctx.Get("permissions")
		if !exists {
			ctx.Error(errors.Forbidden("permission not found"))
			ctx.Abort()
			return
		}

		var userPermissions []string
		switch v := permissionsInterface.(type) {
		case []interface{}:
			for _, perm := range v {
				if str, ok := perm.(string); ok {
					userPermissions = append(userPermissions, str)
				}
			}
		case []string:
			userPermissions = v
		default:
			ctx.Error(errors.Forbidden("invalid permissions format"))
			ctx.Abort()
			return
		}

		for _, required := range requiredPermissions {
			hasPermission := false
			for _, userPerm := range userPermissions {
				if userPerm == required || userPerm == "manage:all" {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				ctx.Error(errors.Forbidden("insufficient permissions"))
				ctx.Abort()
				return
			}
		}

		ctx.Next()
	}
}
