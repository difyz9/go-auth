# 从现有系统迁移到 GoAuth

本指南帮助你将现有的基于数据库的 API 认证系统迁移到 GoAuth 配置文件方案。

## 迁移优势

- ✅ **无数据库依赖**：减少系统复杂度和维护成本
- ✅ **配置即代码**：可以版本控制，便于追踪变更
- ✅ **快速启动**：无需初始化数据库
- ✅ **易于测试**：可以快速切换不同的配置
- ✅ **性能更好**：无数据库查询开销

## 迁移步骤

### 步骤1：导出现有应用数据

如果你当前使用数据库存储应用配置，首先导出数据：

```sql
-- 导出应用配置
SELECT 
    app_id,
    app_secret,
    app_name,
    status,
    ip_whitelist,
    rate_limit,
    require_sign
FROM tb_api_app
WHERE status = 1;  -- 只导出启用的应用
```

### 步骤2：转换为配置文件

创建 `goauth_config.yaml`：

```yaml
timestamp_tolerance: 300
default_rate_limit: 1000
enable_ip_check: true

apps:
  # 将数据库中的每条记录转换为一个应用配置
  app-001:
    app_id: app-001
    app_secret: secret-from-database
    app_name: 应用名称
    require_sign: true
    enabled: true
    rate_limit: 1000
    ip_whitelist:
      - "192.168.1.*"
```

### 步骤3：编写迁移脚本（可选）

如果应用较多，可以编写脚本自动转换：

```go
package main

import (
    "database/sql"
    "fmt"
    "strings"
    
    _ "github.com/go-sql-driver/mysql"
    "payment_service/pkg/goauth"
)

func migrateFromDatabase(db *sql.DB) error {
    // 查询所有启用的应用
    rows, err := db.Query(`
        SELECT app_id, app_secret, app_name, ip_whitelist, 
               rate_limit, require_sign 
        FROM tb_api_app 
        WHERE status = 1
    `)
    if err != nil {
        return err
    }
    defer rows.Close()
    
    // 创建配置
    config := goauth.NewConfig()
    
    // 遍历每个应用
    for rows.Next() {
        var appID, appSecret, appName, ipWhitelist string
        var rateLimit int
        var requireSign bool
        
        err := rows.Scan(&appID, &appSecret, &appName, 
                        &ipWhitelist, &rateLimit, &requireSign)
        if err != nil {
            return err
        }
        
        // 解析IP白名单
        var ips []string
        if ipWhitelist != "" {
            ips = strings.Split(ipWhitelist, ",")
            for i := range ips {
                ips[i] = strings.TrimSpace(ips[i])
            }
        }
        
        // 添加到配置
        config.AddApp(&goauth.AppConfig{
            AppID:       appID,
            AppSecret:   appSecret,
            AppName:     appName,
            RequireSign: requireSign,
            Enabled:     true,
            RateLimit:   rateLimit,
            IPWhitelist: ips,
        })
    }
    
    // 保存配置文件
    return config.SaveToYAML("goauth_config.yaml")
}

func main() {
    db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/dbname")
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    if err := migrateFromDatabase(db); err != nil {
        panic(err)
    }
    
    fmt.Println("迁移完成！配置已保存到 goauth_config.yaml")
}
```

### 步骤4：更新代码

#### 旧代码（使用数据库）

```go
import (
    "payment_service/internal/middleware"
    "gorm.io/gorm"
)

func setupRouter(db *gorm.DB) *gin.Engine {
    r := gin.Default()
    
    // 创建认证中间件（依赖数据库）
    authMiddleware := middleware.NewApiAuthMiddleware(db)
    
    api := r.Group("/api")
    api.Use(authMiddleware.AuthenticateAPI())
    {
        // ... 路由定义
    }
    
    return r
}
```

#### 新代码（使用配置文件）

```go
import (
    "payment_service/pkg/goauth"
)

func setupRouter() *gin.Engine {
    r := gin.Default()
    
    // 加载配置
    config := goauth.NewConfig()
    config.LoadFromYAML("goauth_config.yaml")
    
    // 创建认证中间件（无需数据库）
    authMiddleware := goauth.NewAuthMiddleware(goauth.Options{
        Config: config,
    })
    
    api := r.Group("/api")
    api.Use(authMiddleware.Authenticate())
    {
        // ... 路由定义
    }
    
    return r
}
```

### 步骤5：更新依赖

在 `go.mod` 中移除数据库相关依赖（如果只用于认证）：

```bash
# 可以移除这些依赖（如果只用于认证）
# gorm.io/gorm
# gorm.io/driver/mysql
```

添加新依赖：

```bash
go get gopkg.in/yaml.v3
```

### 步骤6：测试

1. 启动服务
2. 运行测试脚本验证功能
```bash
./test_api.sh http://localhost:8080 test-app-001 test-secret
```

## 保留数据库方案（混合模式）

如果你仍需要动态管理应用，可以保留数据库，但在启动时加载到内存：

```go
func setupAuthFromDatabase(db *gorm.DB) *goauth.AuthMiddleware {
    // 创建配置
    config := goauth.NewConfig()
    
    // 从数据库加载应用
    var apps []model.TbApiApp
    db.Where("status = ?", 1).Find(&apps)
    
    for _, app := range apps {
        // 解析IP白名单
        var ips []string
        if app.IpWhitelist != "" {
            ips = strings.Split(app.IpWhitelist, ",")
        }
        
        config.AddApp(&goauth.AppConfig{
            AppID:       app.AppId,
            AppSecret:   app.AppSecret,
            AppName:     app.AppName,
            RequireSign: app.RequireSign,
            Enabled:     app.Status == 1,
            RateLimit:   app.RateLimit,
            IPWhitelist: ips,
        })
    }
    
    return goauth.NewAuthMiddleware(goauth.Options{
        Config: config,
    })
}

// 使用
func main() {
    db := setupDatabase()
    auth := setupAuthFromDatabase(db)
    
    r := gin.Default()
    r.Use(auth.Authenticate())
    
    // 定期重新加载（可选）
    go func() {
        ticker := time.NewTicker(5 * time.Minute)
        for range ticker.C {
            auth = setupAuthFromDatabase(db)
        }
    }()
    
    r.Run(":8080")
}
```

## 对比表

| 特性 | 数据库方案 | GoAuth配置文件方案 |
|------|-----------|------------------|
| 依赖 | 需要数据库 | 无需数据库 |
| 启动速度 | 较慢（需连接DB） | 快速 |
| 配置管理 | SQL/管理界面 | YAML/JSON文件 |
| 动态更新 | 支持（需重新查询） | 需重启或重新加载 |
| 版本控制 | 困难 | 容易（Git） |
| 性能 | 每次验证查询DB | 内存操作，更快 |
| 适用场景 | 应用频繁变更 | 应用相对固定 |

## 常见问题

### Q: 如何动态添加应用？

**方案1：重启服务**
- 修改配置文件
- 重启应用

**方案2：热重载**
```go
// 提供一个管理接口
admin.POST("/reload-config", func(c *gin.Context) {
    config.LoadFromYAML("goauth_config.yaml")
    c.JSON(200, gin.H{"message": "配置已重新加载"})
})
```

**方案3：保留数据库用于管理**
- 使用数据库管理应用
- 启动时加载到配置
- 定期同步或提供手动同步接口

### Q: 如何管理大量应用？

**建议：**
1. 使用多个配置文件按环境或业务分类
2. 使用脚本生成配置文件
3. 考虑使用配置中心（如 etcd、consul）
4. 对于超大规模，仍建议保留数据库方案

### Q: 配置文件如何保证安全？

**建议：**
1. 使用环境变量存储敏感信息
2. 配置文件加密存储
3. 使用密钥管理服务（KMS）
4. 限制配置文件访问权限
5. 不要将生产配置提交到版本控制

示例（使用环境变量）：
```go
config := goauth.NewConfig()
config.AddApp(&goauth.AppConfig{
    AppID:     "prod-app",
    AppSecret: os.Getenv("PROD_APP_SECRET"), // 从环境变量读取
    Enabled:   true,
})
```

### Q: 如何回滚到旧系统？

保留旧代码，使用功能开关：

```go
func setupAuth(db *gorm.DB) gin.HandlerFunc {
    useNewAuth := os.Getenv("USE_NEW_AUTH") == "true"
    
    if useNewAuth {
        // 使用新的 GoAuth
        config := goauth.NewConfig()
        config.LoadFromYAML("goauth_config.yaml")
        auth := goauth.NewAuthMiddleware(goauth.Options{Config: config})
        return auth.Authenticate()
    } else {
        // 使用旧的认证方式
        oldAuth := middleware.NewApiAuthMiddleware(db)
        return oldAuth.AuthenticateAPI()
    }
}
```

## 下一步

- 阅读 [README.md](README.md) 了解完整功能
- 查看 [QUICK_START.md](QUICK_START.md) 快速开始
- 运行测试验证功能
- 逐步迁移生产环境

## 需要帮助？

如有迁移问题，请查阅文档或提交 Issue。
