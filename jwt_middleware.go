package goauth

import (
"net/http"
"strings"

"github.com/gin-gonic/gin"
)

const (
// JWT 相关常量
HeaderAuthorization = "Authorization"
BearerPrefix        = "Bearer "

// Context Keys - JWT
ContextKeyUserID   = "goauth_user_id"
ContextKeyUsername = "goauth_username"
ContextKeyRole     = "goauth_role"
ContextKeyClaims   = "goauth_claims"
)

// JWTAuth JWT 认证中间件
func (m *AuthMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.jwtService == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
"code":    500,
"message": "JWT service not configured",
})
			c.Abort()
			return
		}

		authHeader := c.GetHeader(HeaderAuthorization)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
"code":    401,
"message": "Missing authorization header",
})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{
"code":    401,
"message": "Invalid authorization header format",
})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
"code":    401,
"message": "Missing token",
})
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			statusCode := http.StatusUnauthorized
			message := "Invalid token"
			
			switch err {
			case ErrExpiredToken:
				message = "Token expired"
			case ErrTokenRevoked:
				message = "Token revoked"
			case ErrInvalidTokenType:
				message = "Invalid token type"
			}
			
			c.JSON(statusCode, gin.H{
"code":    statusCode,
"message": message,
})
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRole, claims.Role)
		c.Set(ContextKeyClaims, claims)
		if claims.AppID != "" {
			c.Set(ContextKeyAppID, claims.AppID)
		}

		c.Next()
	}
}

// AdminAuth 管理员认证中间件
func (m *AuthMiddleware) AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsValue, exists := c.Get(ContextKeyClaims)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
"code":    401,
"message": "Unauthorized",
})
			c.Abort()
			return
		}

		claims, ok := claimsValue.(*Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
"code":    500,
"message": "Internal error",
})
			c.Abort()
			return
		}

		if claims.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
"code":    403,
"message": "Forbidden",
})
			c.Abort()
			return
		}

		c.Next()
	}
}

// WithJWT 设置 JWT 服务
func (m *AuthMiddleware) WithJWT(jwtService *JWTService) *AuthMiddleware {
	m.jwtService = jwtService
	return m
}

// GetUserID 从 Context 获取用户 ID
func GetUserID(c *gin.Context) (uint64, bool) {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, false
	}
	id, ok := userID.(uint64)
	return id, ok
}

// GetUsername 从 Context 获取用户名
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get(ContextKeyUsername)
	if !exists {
		return "", false
	}
	name, ok := username.(string)
	return name, ok
}

// GetClaims 从 Context 获取 Claims
func GetClaims(c *gin.Context) (*Claims, bool) {
	claims, exists := c.Get(ContextKeyClaims)
	if !exists {
		return nil, false
	}
	c2, ok := claims.(*Claims)
	return c2, ok
}
