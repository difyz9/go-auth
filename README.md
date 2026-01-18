# GoAuth - 简单易用的 Go API 认证中间件

GoAuth 是一个轻量级、易于集成的 Go API 认证中间件，专为 Gin 框架设计。无需数据库依赖，通过配置文件即可快速实现 API 认证、签名验证、IP 白名单和速率限制等功能。

## 特性

- ✅ **零数据库依赖** - 使用配置文件（YAML/JSON）管理应用
- ✅ **简单易用** - 仅需几行代码即可集成
- ✅ **安全可靠** - HMAC-SHA256 签名 + 密码学安全随机数
- ✅ **灵活配置** - 支持内存配置和文件配置
- ✅ **IP 白名单** - 支持 CIDR、通配符和精确匹配
- ✅ **速率限制** - 内置请求频率限制
- ✅ **时间戳验证** - 防重放攻击
- ✅ **高性能** - 配置缓存优化，减少锁竞争
- ✅ **可扩展** - 支持自定义 Logger 和错误处理
- ✅ **标准化错误** - 结构化错误响应，便于调试

## 安装

```bash
go get github.com/difyz9/go-auth
```

## 快速开始

### 最简单的方式（2行代码）

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/difyz9/go-auth"
)

func main() {
    r := gin.Default()
    
    // 只需两行代码即可启用认证！
    config := goauth.QuickConfig("my-app", "my-secret-key-123456")
    r.Use(goauth.New(config).Authenticate())
    
    r.GET("/api/users", func(c *gin.Context) {
        c.JSON(200, gin.H{"users": []string{"Alice", "Bob"}})
    })
    
    r.Run(":8080")
}
```

### 使用配置构建器（链式调用）

```go
// 使用链式调用优雅地构建配置
config := goauth.NewConfigBuilder().
    SetTimestampTolerance(600).
    SetDefaultRateLimit(2000).
    AddSimpleApp("app-001", "secret-001", "应用1").
    AddSimpleApp("app-002", "secret-002", "应用2").
    MustBuild()

r := gin.Default()
r.Use(goauth.New(config).Authenticate())
r.Run(":8080")
```

### 从配置文件加载

#### 1. 创建 `goauth_config.yaml`

```yaml
timestamp_tolerance: 300
default_rate_limit: 1000
enable_ip_check: true

apps:
  test-app-001:
    app_id: test-app-001
    app_secret: your-secret-key-here
    app_name: 测试应用
    require_sign: true
    enabled: true
    rate_limit: 100
    ip_whitelist:
      - "*"
```

#### 2. 一行代码加载配置

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/difyz9/go-auth"
)

func main() {
    r := gin.Default()
    
    // 自动加载并验证配置
    config := goauth.MustLoadYAML("goauth_config.yaml")
    r.Use(goauth.New(config).Authenticate())
    
    r.GET("/api/data", handleData)
    r.Run(":8080")
}

func handleData(c *gin.Context) {
    app, _ := goauth.GetAppFromContext(c)
    c.JSON(200, gin.H{
        "app": app.AppName,
        "data": []int{1, 2, 3},
    })
}
```

### 客户端调用（超简单）

```go
package main

import (
    "fmt"
    "github.com/difyz9/go-auth"
)

func main() {
    // 创建客户端
    client := goauth.NewClient(
        "http://localhost:8080",
        "my-app",
        "my-secret-key-123456",
        goauth.WithDebug(true), // 可选：启用调试
    )
    
    // GET 请求
    var users map[string]interface{}
    if err := client.GetJSON("/api/users", &users); err != nil {
        fmt.Println("请求失败:", err)
        return
    }
    fmt.Printf("用户列表: %+v\n", users)
    
    // POST 请求
    orderData := map[string]interface{}{
        "user_id": 123,
        "amount":  99.99,
    }
    
    var result map[string]interface{}
    if err := client.PostJSON("/api/orders", orderData, &result); err != nil {
        fmt.Println("请求失败:", err)
        return
    }
    fmt.Printf("订单结果: %+v\n", result)
}
```

## 配置说明

### 全局配置

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `timestamp_tolerance` | int64 | 时间戳容差（秒） | 300 |
| `default_rate_limit` | int | 默认速率限制（次/分钟） | 1000 |
| `enable_ip_check` | bool | 是否启用IP检查 | true |

> 💡 **v2.0 更新**：移除了 `sign_include_body` 配置项，签名统一不包含请求体，简化了签名流程。

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

### 1. 使用函数式选项创建配置

```go
// 使用函数式选项优雅地自定义配置
config := goauth.NewConfig(
    goauth.WithTimestampTolerance(600),      // 10分钟容差
    goauth.WithDefaultRateLimit(2000),        // 2000次/分钟
    goauth.WithIPCheck(false),                // 禁用IP检查
)

config.AddApp(&goauth.AppConfig{
    AppID:       "app-001",
    AppSecret:   "secret-key",
    AppName:     "我的应用",
    RequireSign: true,
    Enabled:     true,
    RateLimit:   100,
    IPWhitelist: []string{"192.168.1.*"},
})
```

### 2. 统一的错误响应格式

```go
// 使用内置的统一错误响应格式
customErrorHandler := func(c *gin.Context, code int, message string, detail string) {
    requestID := goauth.BuildRequestID()
    
    err := goauth.NewAuthError(
        goauth.ErrorCode(message),
        message,
        detail,
        code,
    )
    
    response := goauth.NewErrorResponse(err, requestID)
    c.JSON(code, response)
}

r.Use(goauth.New(config).Authenticate(customErrorHandler))
```

### 3. 自定义 Logger 🆕

```go
// 实现 Logger 接口
type MyLogger struct{}

func (l *MyLogger) Info(msg string, fields ...interface{}) {
    log.Printf("[INFO] %s %v", msg, fields)
}

func (l *MyLogger) Error(msg string, fields ...interface{}) {
    log.Printf("[ERROR] %s %v", msg, fields)
}

func (l *MyLogger) Debug(msg string, fields ...interface{}) {
    log.Printf("[DEBUG] %s %v", msg, fields)
}

// 使用自定义 Logger
middleware := goauth.NewAuthMiddleware(goauth.Options{
    Config: config,
    Logger: &MyLogger{},
})

// 客户端也支持自定义 Logger
client := goauth.NewClient(
    baseURL, appID, appSecret,
    goauth.WithLogger(&MyLogger{}),
    goauth.WithDebug(true),
)
```

### 4. 动态管理应用

```go
// 添加应用
config.AddApp(&goauth.AppConfig{
    AppID:     "new-app",
    AppSecret: "secret",
    Enabled:   true,
})

// 刷新中间件缓存（v2.0 新增）
middleware.RefreshAppCache()  // 动态更新配置后需要调用

// 获取应用
app, exists := config.GetApp("new-app")

// 获取所有应用
apps := config.GetApps()

// 获取启用的应用数量
count := config.GetEnabledAppCount()

// 删除应用
config.RemoveApp("new-app")

// 验证配置
if err := config.Validate(); err != nil {
    log.Fatal("配置验证失败:", err)
}

// 保存配置到文件
config.SaveToYAML("goauth_config.yaml")
```

### 5. 便捷的签名工具

```go
// 快速生成签名
params, sign, err := goauth.QuickSign(appID, appSecret, requestBody)
if err != nil {
    log.Fatal(err)
}

// 调试签名问题
client := goauth.NewClient(baseURL, appID, appSecret)
client.DebugSign(requestBody)  // 输出详细的签名信息

// 验证签名并获取调试信息
valid, debugInfo := goauth.VerifySignWithDebug(params, appSecret, sign)
fmt.Println(debugInfo)
```

### 6. 实用工具函数

```go
// 生成安全的应用密钥
secret, err := goauth.GenerateAppSecret(32)

// 遮蔽密钥用于日志输出
masked := goauth.MaskSecret(secret)  // 输出: abcd****xyz

// 验证应用ID格式
if err := goauth.ValidateAppID("my-app-123"); err != nil {
    log.Fatal("应用ID无效")
}

// 生成请求ID
requestID := goauth.BuildRequestID()

// 生成加密安全的随机字符串
nonce, err := goauth.GenerateSecureNonce(16)
```

### 7. 配置构建器模式

```go
// 使用构建器模式创建复杂配置
config, err := goauth.NewConfigBuilder().
    SetTimestampTolerance(600).
    SetDefaultRateLimit(2000).
    SetEnableIPCheck(true).
    AddSimpleApp("app-001", "secret-001", "应用1").
    AddSimpleApp("app-002", "secret-002", "应用2").
    AddApp(&goauth.AppConfig{
        AppID:       "app-003",
        AppSecret:   "secret-003",
        AppName:     "应用3",
        RequireSign: false,  // 不需要签名
        Enabled:     true,
        RateLimit:   500,
        IPWhitelist: []string{"192.168.1.*"},
    }).
    Build()

if err != nil {
    log.Fatal("配置构建失败:", err)
}

// 或使用 MustBuild（失败则panic）
config := goauth.NewConfigBuilder().
    AddSimpleApp("app-001", "secret-001", "应用1").
    MustBuild()
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

### 签名参数说明

**v2.0 优化**：签名统一只包含 `appId`、`timestamp`、`nonce` 三个参数，不包含请求体。

**优势**：
- ✅ 更简单：减少签名复杂度
- ✅ 更灵活：请求体可独立修改
- ✅ 更快速：无需序列化请求体
- ✅ 更易调试：签名参数少，问题更容易定位

### 示例

```
参数（不包含请求体，推荐）：
  appId: test-app-001
  timestamp: 1700000000
  nonce: abc123

排序后拼接：
  appId=test-app-001&nonce=abc123&timestamp=1700000000

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

### Q: 签名验证失败怎么办？

A: 检查以下几点：
1. **参数完整性**：确保 `appId`、`timestamp`、`nonce` 都已传递
2. **时间同步**：客户端和服务器时间差不超过容差（默认 5 分钟）
3. **密钥正确**：验证 `appSecret` 是否匹配
4. **参数排序**：确保参数按 ASCII 码排序
5. **调试模式**：启用 `WithDebug(true)` 查看详细签名过程

```go
client := goauth.NewClient(baseURL, appID, appSecret,
    goauth.WithDebug(true))  // 启用调试
client.DebugSign(body)      // 输出签名详情
```

### Q: 支持哪些 IP 白名单格式？

A: **v2.0 增强** - 支持：
- 单个 IP: `192.168.1.1`
- CIDR 格式: `192.168.1.0/24` 🆕
- IP 段（通配符）: `192.168.1.*` （兼容旧版）
- 本地地址: `localhost`, `127.0.0.1`, `::1`

推荐使用 CIDR 格式，更标准、更精确。

### Q: 如何调试签名问题？

A: 
1. 检查参数是否完整
2. 确认参数排序是否正确
3. 验证拼接字符串格式
4. 确保 `appSecret` 正确
5. 启用客户端调试模式：`goauth.WithDebug(true)`
6. 使用 `client.DebugSign(body)` 查看签名过程

## v2.0 新特性 🎉

### 安全性增强
- 🔒 **密码学安全随机数**：使用 `crypto/rand` 生成 Nonce，防止预测攻击
- 🎯 **标准化错误处理**：支持 `errors.Is` 判断，便于精细化错误处理

### 性能优化
- ⚡ **配置缓存**：预加载应用配置，减少锁竞争，高并发性能提升 20%
- 📊 **更好的错误信息**：HTTP 错误返回结构化信息，便于调试

### 功能增强
- 🌐 **CIDR 支持**：IP 白名单支持标准 CIDR 格式（如 `192.168.1.0/24`）
- 🔌 **可插拔 Logger**：支持自定义日志实现，集成第三方日志库
- 🧹 **简化签名**：移除请求体签名选项，统一签名流程

详细改进请查看 [CODE_REVIEW_REPORT.md](CODE_REVIEW_REPORT.md)

## 完整示例

查看 `examples/` 目录获取完整的服务端和客户端示例代码。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
