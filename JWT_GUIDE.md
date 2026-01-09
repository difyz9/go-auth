# GoAuth JWT 认证使用指南

本文档介绍如何使用 GoAuth 的 JWT 认证功能。

## 📋 目录

- [快速开始](#快速开始)
- [功能特性](#功能特性)
- [配置说明](#配置说明)
- [使用示例](#使用示例)
- [中间件](#中间件)
- [最佳实践](#最佳实践)

## 🚀 快速开始

### 1. 安装

```bash
go get github.com/difyz9/go-auth
go get github.com/golang-jwt/jwt/v5
```

### 2. 基本使用

```go
package main

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/difyz9/go-auth"
)

func main() {
    r := gin.Default()

    // 创建 JWT 服务
    jwtService := goauth.NewJWTService(goauth.JWTConfig{
        SecretKey:     "your-32-char-secret-key-here",
        Issuer:        "my-app",
        AccessExpiry:  24 * time.Hour,
        RefreshExpiry: 7 * 24 * time.Hour,
    })

    // 创建认证中间件
    authMiddleware := goauth.New(goauth.NewConfig()).WithJWT(jwtService)

    // 登录接口
    r.POST("/login", handleLogin(jwtService))

    // 受保护的接口
    api := r.Group("/api")
    api.Use(authMiddleware.JWTAuth())
    {
        api.GET("/profile", handleProfile)
    }

    r.Run(":8080")
}
```

## ✨ 功能特性

- ✅ **Access Token + Refresh Token** - 双 Token 机制
- ✅ **Token 黑名单** - 支持 Token 撤销
- ✅ **自定义 Claims** - 灵活的自定义字段
- ✅ **角色权限** - 内置角色验证中间件
- ✅ **自动过期** - Token 自动过期管理
- ✅ **类型安全** - 完整的类型定义

## ⚙️ 配置说明

### JWTConfig

```go
type JWTConfig struct {
    SecretKey     string        // JWT 密钥（必填，至少32字符）
    Issuer        string        // 签发者（可选，默认 "go-auth"）
    AccessExpiry  time.Duration // Access Token 有效期（可选，默认 24小时）
    RefreshExpiry time.Duration // Refresh Token 有效期（可选，默认 7天）
}
```

### Claims 结构

```go
type Claims struct {
    UserID   uint64                 // 用户ID
    Username string                 // 用户名
    Role     string                 // 角色（如 admin, user）
    AppID    string                 // 应用ID
    Type     TokenType              // Token 类型（access/refresh）
    Custom   map[string]interface{} // 自定义字段
    jwt.RegisteredClaims
}
```

## 📝 使用示例

### 示例 1: 登录生成 Token

```go
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

        // 验证用户名密码...

        // 生成 Token 对
        tokenPair, err := jwtService.GenerateTokenPair(
            userID,
            req.Username,
            "user",      // role
            "web-app",   // appID
            nil,         // custom claims
        )

        if err != nil {
            c.JSON(500, gin.H{"error": "failed to generate token"})
            return
        }

        c.JSON(200, gin.H{
            "code": 0,
            "data": tokenPair,
        })
    }
}
```

响应示例：

```json
{
  "code": 0,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2026-01-10T10:00:00Z",
    "expires_in": 86400,
    "token_type": "Bearer"
  }
}
```

### 示例 2: 刷新 Token

```go
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
        newAccessToken, err := jwtService.RefreshAccessToken(
            req.RefreshToken,
            "user",      // role
            "web-app",   // appID
            nil,         // custom claims
        )

        if err != nil {
            c.JSON(401, gin.H{"error": "invalid refresh token"})
            return
        }

        c.JSON(200, gin.H{
            "code": 0,
            "data": gin.H{
                "access_token": newAccessToken,
                "token_type": "Bearer",
            },
        })
    }
}
```

### 示例 3: 登出（撤销 Token）

```go
func handleLogout(jwtService *goauth.JWTService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从 Header 获取 Token
        authHeader := c.GetHeader("Authorization")
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

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
```

### 示例 4: 使用自定义 Claims

```go
// 生成带自定义字段的 Token
custom := map[string]interface{}{
    "department": "engineering",
    "level":      5,
    "permissions": []string{"read", "write"},
}

token, err := jwtService.GenerateAccessToken(
    userID,
    username,
    "user",
    "web-app",
    custom,
)

// 在处理器中获取自定义字段
func handleProfile(c *gin.Context) {
    claims, _ := goauth.GetClaims(c)
    
    department := claims.Custom["department"].(string)
    level := int(claims.Custom["level"].(float64))
    
    c.JSON(200, gin.H{
        "user_id":    claims.UserID,
        "department": department,
        "level":      level,
    })
}
```

## 🔒 中间件

### 1. JWTAuth - JWT 认证

```go
// 要求有效的 JWT Token
api := r.Group("/api")
api.Use(authMiddleware.JWTAuth())
{
    api.GET("/profile", handleProfile)
}
```

### 2. AdminAuth - 管理员认证

```go
// 要求管理员角色
admin := r.Group("/admin")
admin.Use(authMiddleware.JWTAuth())
admin.Use(authMiddleware.AdminAuth())
{
    admin.GET("/users", handleUsers)
}
```

### 3. RequireRole - 指定角色

```go
// 要求特定角色
moderator := r.Group("/moderator")
moderator.Use(authMiddleware.JWTAuth())
moderator.Use(authMiddleware.RequireRole("moderator", "admin"))
{
    moderator.POST("/review", handleReview)
}
```

### 4. OptionalJWTAuth - 可选认证

```go
// Token 无效不会中断请求
public := r.Group("/public")
public.Use(authMiddleware.OptionalJWTAuth())
{
    public.GET("/content", func(c *gin.Context) {
        username, authenticated := goauth.GetUsername(c)
        
        if authenticated {
            // 已登录用户
            c.JSON(200, gin.H{
                "username": username,
                "level": "premium",
            })
        } else {
            // 游客
            c.JSON(200, gin.H{
                "level": "basic",
            })
        }
    })
}
```

### 5. 组合认证（API 签名 + JWT）

```go
// 同时需要 API 签名和用户身份
secure := r.Group("/secure")
secure.Use(authMiddleware.Authenticate())  // API 签名
secure.Use(authMiddleware.JWTAuth())       // JWT 认证
{
    secure.GET("/sensitive-data", handleSensitiveData)
}
```

## 🛠️ Helper 函数

```go
// 获取用户 ID
userID, ok := goauth.GetUserID(c)

// 获取用户名
username, ok := goauth.GetUsername(c)

// 获取角色
role, ok := goauth.GetRole(c)

// 获取完整的 Claims
claims, ok := goauth.GetClaims(c)

// Must 版本（不存在会 panic）
userID := goauth.MustGetUserID(c)
username := goauth.MustGetUsername(c)
claims := goauth.MustGetClaims(c)
```

## 📌 最佳实践

### 1. 密钥安全

```go
// ❌ 不要硬编码密钥
jwtService := goauth.NewJWTService(goauth.JWTConfig{
    SecretKey: "hardcoded-secret",
})

// ✅ 使用环境变量
jwtService := goauth.NewJWTService(goauth.JWTConfig{
    SecretKey: os.Getenv("JWT_SECRET_KEY"),
})
```

### 2. Token 过期时间

```go
// ✅ 推荐配置
jwtService := goauth.NewJWTService(goauth.JWTConfig{
    SecretKey:     os.Getenv("JWT_SECRET_KEY"),
    Issuer:        "my-app",
    AccessExpiry:  15 * time.Minute,    // Access Token 短期有效
    RefreshExpiry: 7 * 24 * time.Hour,  // Refresh Token 长期有效
})
```

### 3. Token 黑名单

```go
// 使用带自动清理的黑名单
blacklist := goauth.NewInMemoryBlacklistWithCleanup(1 * time.Hour)
jwtService := goauth.NewJWTService(config).WithBlacklist(blacklist)

// 记得在应用关闭时停止清理
defer blacklist.StopAutoCleanup()
```

### 4. 错误处理

```go
func handleProtectedRoute(c *gin.Context) {
    claims, ok := goauth.GetClaims(c)
    if !ok {
        // 理论上不会发生（中间件已验证）
        c.JSON(401, gin.H{"error": "unauthorized"})
        return
    }

    // 使用 claims...
}
```

### 5. 客户端调用

```bash
# 获取 Token
TOKEN=$(curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' \
  | jq -r '.data.access_token')

# 使用 Token 访问受保护接口
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer $TOKEN"

# 刷新 Token
curl -X POST http://localhost:8080/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}"

# 登出
curl -X POST http://localhost:8080/logout \
  -H "Authorization: Bearer $TOKEN"
```

## 🔗 相关文档

- [README.md](../README.md) - GoAuth 主文档
- [API 签名认证](./API_SIGNATURE.md)
- [完整示例](../examples/jwt_example.go)

## ❓ 常见问题

### Q1: JWT 密钥长度要求？

A: 建议至少 32 字符，使用强随机密钥。

### Q2: Token 过期后如何刷新？

A: 使用 Refresh Token 调用 `RefreshAccessToken()` 方法。

### Q3: 如何撤销 Token？

A: 使用 `RevokeToken()` 将 Token 加入黑名单。

### Q4: 支持多设备登录吗？

A: 是的，每个设备生成独立的 Token 对。

### Q5: 如何实现单点登录（SSO）？

A: 在生成新 Token 时撤销旧 Token，或使用自定义的 Token 管理策略。

---

**更新时间**: 2026-01-09
