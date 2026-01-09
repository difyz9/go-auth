package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/difyz9/go-auth"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// ========== 1. 创建 JWT 服务 ==========
	jwtService := goauth.NewJWTService(goauth.JWTConfig{
		SecretKey:     "your-super-secret-jwt-key-must-be-at-least-32-chars-long",
		Issuer:        "my-app",
		AccessExpiry:  24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
	})

	// ========== 2. 创建 API 签名认证配置 ==========
	config := goauth.NewConfigBuilder().
		SetTimestampTolerance(600).
		SetDefaultRateLimit(1000).
		AddSimpleApp("app-001", "secret-001", "应用1").
		AddSimpleApp("app-002", "secret-002", "应用2").
		MustBuild()

	// ========== 3. 创建认证中间件 ==========
	authMiddleware := goauth.New(config).WithJWT(jwtService)

	// ========== 4. 公开接口（无需认证） ==========
	public := r.Group("/api/v1/public")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	// ========== 5. 认证接口（JWT） ==========
	// 登录接口（生成 Token）
	r.POST("/api/v1/auth/login", handleLogin(jwtService))
	
	// 刷新 Token 接口
	r.POST("/api/v1/auth/refresh", handleRefresh(jwtService))
	
	// 登出接口
	r.POST("/api/v1/auth/logout", authMiddleware.JWTAuth(), handleLogout(jwtService))

	// ========== 6. 需要 JWT 认证的接口 ==========
	user := r.Group("/api/v1/user")
	user.Use(authMiddleware.JWTAuth())
	{
		user.GET("/profile", handleUserProfile)
		user.PUT("/profile", handleUpdateProfile)
	}

	// ========== 7. 需要管理员权限的接口 ==========
	admin := r.Group("/api/v1/admin")
	admin.Use(authMiddleware.JWTAuth())
	admin.Use(authMiddleware.AdminAuth())
	{
		admin.GET("/users", handleListUsers)
		admin.POST("/users", handleCreateUser)
	}

	// ========== 8. 需要 API 签名认证的接口 ==========
	appAPI := r.Group("/api/v1/app")
	appAPI.Use(authMiddleware.Authenticate())
	{
		appAPI.GET("/data", handleAppData)
	}

	// ========== 9. 组合认证：API 签名 + JWT ==========
	combined := r.Group("/api/v1/secure")
	combined.Use(authMiddleware.Authenticate())  // 先验证应用身份
	combined.Use(authMiddleware.JWTAuth())       // 再验证用户身份
	{
		combined.GET("/sensitive-data", handleSensitiveData)
	}

	// ========== 10. 可选 JWT 认证 ==========
	optional := r.Group("/api/v1/optional")
	optional.Use(authMiddleware.OptionalJWTAuth())
	{
		optional.GET("/content", handleOptionalContent)
	}

	fmt.Println("Server running on :8080")
	fmt.Println("\n测试命令:")
	fmt.Println("1. 登录获取 Token:")
	fmt.Println("   curl -X POST http://localhost:8080/api/v1/auth/login \\")
	fmt.Println("     -H 'Content-Type: application/json' \\")
	fmt.Println("     -d '{\"username\":\"admin\",\"password\":\"password\"}'")
	fmt.Println("\n2. 使用 Token 访问受保护接口:")
	fmt.Println("   curl http://localhost:8080/api/v1/user/profile \\")
	fmt.Println("     -H 'Authorization: Bearer YOUR_TOKEN'")
	fmt.Println("\n3. 使用 API 签名访问接口:")
	fmt.Println("   ./generate_sign.sh app-001 secret-001 /api/v1/app/data")
	fmt.Println("")

	r.Run(":8080")
}

// handleLogin 登录处理
func handleLogin(jwtService *goauth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		// 这里应该验证用户名密码（示例简化处理）
		if req.Username == "" || req.Password == "" {
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}

		// 确定用户角色（示例）
		role := "user"
		if req.Username == "admin" {
			role = "admin"
		}

		// 生成 Token 对
		tokenPair, err := jwtService.GenerateTokenPair(
			1,           // userID
			req.Username,
			role,
			"web-app",
			nil,
		)

		if err != nil {
			c.JSON(500, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(200, gin.H{
			"code": 0,
			"message": "login successful",
			"data": tokenPair,
		})
	}
}

// handleRefresh 刷新 Token
func handleRefresh(jwtService *goauth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		// 刷新 Access Token
		newAccessToken, err := jwtService.RefreshAccessToken(req.RefreshToken, "user", "web-app", nil)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid refresh token"})
			return
		}

		c.JSON(200, gin.H{
			"code": 0,
			"message": "token refreshed",
			"data": gin.H{
				"access_token": newAccessToken,
				"token_type": "Bearer",
			},
		})
	}
}

// handleLogout 登出（撤销 Token）
func handleLogout(jwtService *goauth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(400, gin.H{"error": "missing token"})
			return
		}

		tokenString := authHeader[7:] // 去掉 "Bearer "

		// 撤销 Token
		if err := jwtService.RevokeToken(tokenString); err != nil {
			c.JSON(500, gin.H{"error": "failed to revoke token"})
			return
		}

		c.JSON(200, gin.H{
			"code": 0,
			"message": "logout successful",
		})
	}
}

// handleUserProfile 获取用户信息
func handleUserProfile(c *gin.Context) {
	claims, _ := goauth.GetClaims(c)

	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"user_id":  claims.UserID,
			"username": claims.Username,
			"role":     claims.Role,
			"app_id":   claims.AppID,
		},
	})
}

// handleUpdateProfile 更新用户信息
func handleUpdateProfile(c *gin.Context) {
	userID, _ := goauth.GetUserID(c)
	
	var req struct {
		Email string `json:"email"`
		Phone string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"message": "profile updated",
		"data": gin.H{
			"user_id": userID,
			"email":   req.Email,
			"phone":   req.Phone,
		},
	})
}

// handleListUsers 列出所有用户（管理员接口）
func handleListUsers(c *gin.Context) {
	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"users": []gin.H{
				{"id": 1, "username": "user1"},
				{"id": 2, "username": "user2"},
			},
		},
	})
}

// handleCreateUser 创建用户（管理员接口）
func handleCreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"message": "user created",
		"data": gin.H{
			"id":       3,
			"username": req.Username,
		},
	})
}

// handleAppData 应用数据接口（需要 API 签名）
func handleAppData(c *gin.Context) {
	appID, _ := c.Get(goauth.ContextKeyAppID)

	c.JSON(200, gin.H{
		"code": 0,
		"message": "app data",
		"app_id": appID,
	})
}

// handleSensitiveData 敏感数据接口（需要 API 签名 + JWT）
func handleSensitiveData(c *gin.Context) {
	appID, _ := c.Get(goauth.ContextKeyAppID)
	claims, _ := goauth.GetClaims(c)

	c.JSON(200, gin.H{
		"code": 0,
		"message": "sensitive data",
		"app_id": appID,
		"user": gin.H{
			"id":       claims.UserID,
			"username": claims.Username,
			"role":     claims.Role,
		},
	})
}

// handleOptionalContent 可选认证内容
func handleOptionalContent(c *gin.Context) {
	username, authenticated := goauth.GetUsername(c)

	response := gin.H{
		"code": 0,
		"message": "content",
		"authenticated": authenticated,
	}

	if authenticated {
		response["username"] = username
		response["level"] = "premium"
	} else {
		response["level"] = "basic"
	}

	c.JSON(200, response)
}
