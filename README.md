# GoAuth - 简单易用的 Go API 认证中间件

GoAuth 是一个轻量级、易于集成的 Go API 认证中间件，专为 Gin 框架设计。无需数据库依赖，通过配置文件即可快速实现 API 认证、签名验证、IP 白名单和速率限制等功能。

## 特性

- ✅ **零数据库依赖** - 使用配置文件（YAML/JSON）管理应用
- ✅ **简单易用** - 仅需几行代码即可集成
- ✅ **安全可靠** - HMAC-SHA256 签名算法
- ✅ **灵活配置** - 支持内存配置和文件配置
- ✅ **IP 白名单** - 支持通配符匹配
- ✅ **速率限制** - 内置请求频率限制
- ✅ **时间戳验证** - 防重放攻击
- ✅ **可扩展** - 支持自定义日志和错误处理

## 安装

```bash
go get github.com/difyz9/go-auth
```

## 快速开始

### 1. 创建配置文件

创建 `goauth_config.yaml`:

```yaml
timestamp_tolerance: 300  # 时间戳容差（秒）
default_rate_limit: 1000  # 默认速率限制（次/分钟）
enable_ip_check: true     # 是否启用IP检查

apps:
  test-app-001:
    app_id: test-app-001
    app_secret: your-secret-key-here
    app_name: 测试应用
    require_sign: true
    enabled: true
    rate_limit: 100
    ip_whitelist:
      - "*"  # 允许所有IP，生产环境请配置具体IP
    
  production-app:
    app_id: prod-app-001
    app_secret: prod-secret-key-here
    app_name: 生产应用
    require_sign: true
    enabled: true
    rate_limit: 1000
    ip_whitelist:
      - "192.168.1.*"
      - "10.0.0.1"
```

### 2. 集成到 Gin 应用

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/yourusername/payment_service/pkg/goauth"
)

func main() {
    // 创建配置
    config := goauth.NewConfig()
    
    // 从文件加载配置
    if err := config.LoadFromYAML("goauth_config.yaml"); err != nil {
        panic(err)
    }
    
    // 创建认证中间件
    authMiddleware := goauth.NewAuthMiddleware(goauth.Options{
        Config: config,
    })
    
    // 创建 Gin 路由
    r := gin.Default()
    
    // 应用认证中间件
    api := r.Group("/api")
    api.Use(authMiddleware.Authenticate())
    {
        api.GET("/users", getUsers)
        api.POST("/orders", createOrder)
    }
    
    r.Run(":8080")
}

func getUsers(c *gin.Context) {
    // 获取当前应用信息
    app, _ := goauth.GetAppFromContext(c)
    c.JSON(200, gin.H{
        "app": app.AppName,
        "users": []string{"user1", "user2"},
    })
}

func createOrder(c *gin.Context) {
    c.JSON(200, gin.H{"status": "success"})
}
```

### 3. 客户端调用示例

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"
    
    "github.com/yourusername/payment_service/pkg/goauth"
)

func main() {
    appID := "test-app-001"
    appSecret := "your-secret-key-here"
    
    // 准备请求参数
    timestamp := strconv.FormatInt(time.Now().Unix(), 10)
    nonce := goauth.GenerateNonce(16)
    
    // 请求体
    requestBody := map[string]interface{}{
        "user_id": 123,
        "amount":  100.00,
    }
    bodyBytes, _ := json.Marshal(requestBody)
    
    // 生成签名
    params := map[string]string{
        "appId":       appID,
        "timestamp":   timestamp,
        "nonce":       nonce,
        "requestBody": string(bodyBytes),
    }
    sign := goauth.GenerateSign(params, appSecret)
    
    // 发送请求
    req, _ := http.NewRequest("POST", "http://localhost:8080/api/orders", bytes.NewBuffer(bodyBytes))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-App-Id", appID)
    req.Header.Set("X-Timestamp", timestamp)
    req.Header.Set("X-Nonce", nonce)
    req.Header.Set("X-Sign", sign)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("请求失败:", err)
        return
    }
    defer resp.Body.Close()
    
    fmt.Println("响应状态:", resp.Status)
}
```

## 配置说明

### 全局配置

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `timestamp_tolerance` | int64 | 时间戳容差（秒） | 300 |
| `default_rate_limit` | int | 默认速率限制（次/分钟） | 1000 |
| `enable_ip_check` | bool | 是否启用IP检查 | true |

### 应用配置

| 参数 | 类型 | 说明 | 必填 |
|------|------|------|------|
| `app_id` | string | 应用ID | 是 |
| `app_secret` | string | 应用密钥 | 是 |
| `app_name` | string | 应用名称 | 否 |
| `require_sign` | bool | 是否需要签名验证 | 否（默认true） |
| `enabled` | bool | 是否启用 | 否（默认false） |
| `rate_limit` | int | 速率限制（次/分钟） | 否 |
| `ip_whitelist` | []string | IP白名单 | 否 |

## 高级用法

### 1. 内存配置（无需配置文件）

```go
config := goauth.NewConfig()

// 添加应用配置
config.AddApp(&goauth.AppConfig{
    AppID:       "app-001",
    AppSecret:   "secret-key",
    AppName:     "我的应用",
    RequireSign: true,
    Enabled:     true,
    RateLimit:   100,
    IPWhitelist: []string{"192.168.1.*"},
})

authMiddleware := goauth.NewAuthMiddleware(goauth.Options{
    Config: config,
})
```

### 2. 自定义错误处理

```go
customErrorHandler := func(c *gin.Context, code int, message string, detail string) {
    c.JSON(code, gin.H{
        "success": false,
        "error": gin.H{
            "code":    code,
            "message": message,
            "detail":  detail,
        },
        "timestamp": time.Now().Unix(),
    })
}

r.Use(authMiddleware.Authenticate(customErrorHandler))
```

### 3. 自定义日志

```go
type MyLogger struct{}

func (l *MyLogger) Info(msg string, fields ...interface{})  { 
    log.Info(msg, fields...) 
}
func (l *MyLogger) Error(msg string, fields ...interface{}) { 
    log.Error(msg, fields...) 
}
func (l *MyLogger) Debug(msg string, fields ...interface{}) { 
    log.Debug(msg, fields...) 
}

authMiddleware := goauth.NewAuthMiddleware(goauth.Options{
    Config: config,
    Logger: &MyLogger{},
})
```

### 4. 动态管理应用

```go
// 添加应用
config.AddApp(&goauth.AppConfig{
    AppID:     "new-app",
    AppSecret: "secret",
    Enabled:   true,
})

// 获取应用
app, exists := config.GetApp("new-app")

// 删除应用
config.RemoveApp("new-app")

// 保存配置到文件
config.SaveToYAML("goauth_config.yaml")
```

## API 认证流程

1. 客户端准备请求参数：`appId`, `timestamp`, `nonce`
2. 客户端生成签名：
   - 将所有参数（包括请求体）按 key 排序
   - 拼接成 `key1=value1&key2=value2` 格式
   - 使用 HMAC-SHA256 算法生成签名
3. 客户端发送请求，将认证参数放在 Header 中
4. 服务端验证：
   - 检查应用是否存在且启用
   - 验证时间戳是否在有效期内
   - 检查 IP 是否在白名单中
   - 验证速率限制
   - 验证签名是否正确

## 签名算法

### 签名生成步骤

1. **收集参数**：包括 `appId`, `timestamp`, `nonce` 和所有业务参数
2. **参数排序**：按 ASCII 码从小到大排序
3. **拼接字符串**：格式为 `key1=value1&key2=value2`
4. **HMAC-SHA256**：使用 `appSecret` 作为密钥
5. **十六进制编码**：将结果转换为小写十六进制字符串

### 示例

```
参数：
  appId: test-app-001
  timestamp: 1700000000
  nonce: abc123
  amount: 100.00

排序后拼接：
  amount=100.00&appId=test-app-001&nonce=abc123&timestamp=1700000000

HMAC-SHA256(上述字符串, appSecret)
=> 签名结果
```

## 错误码说明

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 请求参数错误或格式错误 |
| 401 | 认证失败（应用不存在、签名错误） |
| 403 | 禁止访问（IP 不在白名单） |
| 429 | 请求过于频繁（超过速率限制） |

## 最佳实践

1. **生产环境配置**
   - 使用强密钥（至少 32 位随机字符）
   - 配置具体的 IP 白名单，避免使用 `*`
   - 根据实际情况设置合理的速率限制
   - 启用签名验证

2. **安全建议**
   - 密钥定期轮换
   - 使用 HTTPS 加密传输
   - 记录所有认证失败的请求
   - 监控异常访问模式

3. **性能优化**
   - 使用配置文件缓存，避免频繁读取
   - 合理设置速率限制，避免过度限制
   - 定期清理过期的速率限制计数器

## 常见问题

### Q: 如何在不需要签名的情况下使用？

A: 将应用配置中的 `require_sign` 设置为 `false`。

### Q: 支持哪些 IP 白名单格式？

A: 支持：
- 单个 IP: `192.168.1.1`
- IP 段（通配符）: `192.168.1.*`
- 允许所有: `*`
- 本地地址: `localhost`, `127.0.0.1`, `::1`

### Q: 如何调试签名问题？

A: 
1. 检查参数是否完整
2. 确认参数排序是否正确
3. 验证拼接字符串格式
4. 确保 `appSecret` 正确
5. 使用提供的 `SignatureExample` 函数生成测试签名

## 完整示例

查看 `example.go` 文件获取完整的服务端和客户端示例代码。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
