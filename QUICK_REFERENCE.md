# GoAuth 快速参考

## 安装

```bash
go get github.com/difyz9/go-auth
```

## 最简单的使用（2行代码）

```go
config := goauth.QuickConfig("my-app", "my-secret-key-123456")
r.Use(goauth.New(config).Authenticate())
```

## 服务端常用方法

### 配置创建

```go
// 方式1：快速配置
config := goauth.QuickConfig("app-id", "app-secret")

// 方式2：构建器模式
config := goauth.NewConfigBuilder().
    SetTimestampTolerance(600).
    AddSimpleApp("app-id", "app-secret", "应用名").
    MustBuild()

// 方式3：加载文件
config := goauth.MustLoadYAML("config.yaml")
config := goauth.MustLoadJSON("config.json")

// 方式4：函数式选项
config := goauth.NewConfig(
    goauth.WithTimestampTolerance(600),
    goauth.WithDefaultRateLimit(2000),
    goauth.WithIPCheck(false),
)
```

### 中间件使用

```go
// 基本使用
r.Use(goauth.New(config).Authenticate())

// 链式调用
auth := goauth.New(config).
    WithLogger(myLogger)
r.Use(auth.Authenticate())

// 自定义错误处理
r.Use(goauth.New(config).Authenticate(customErrorHandler))
```

### 获取应用信息

```go
// 在处理函数中
app, exists := goauth.GetAppFromContext(c)
appID, exists := goauth.GetAppIDFromContext(c)
```

### 配置管理

```go
// 添加应用
config.AddApp(&goauth.AppConfig{
    AppID:       "app-id",
    AppSecret:   "secret",
    AppName:     "应用名",
    RequireSign: true,
    Enabled:     true,
    RateLimit:   100,
    IPWhitelist: []string{"*"},
})

// 获取应用
app, exists := config.GetApp("app-id")

// 获取所有应用
apps := config.GetApps()

// 获取启用的应用数量
count := config.GetEnabledAppCount()

// 删除应用
config.RemoveApp("app-id")

// 验证配置
err := config.Validate()

// 保存配置
config.SaveToYAML("config.yaml")
config.SaveToJSON("config.json")
```

## 客户端常用方法

### 创建客户端

```go
// 基本创建
client := goauth.NewClient(baseURL, appID, appSecret)

// 带选项创建
client := goauth.NewClient(
    baseURL, appID, appSecret,
    goauth.WithDebug(true),
    goauth.WithTimeout(30*time.Second),
)
```

### 发送请求

```go
// GET 请求
var result map[string]interface{}
err := client.GetJSON("/api/users", &result)

// POST 请求
data := map[string]interface{}{"key": "value"}
err := client.PostJSON("/api/data", data, &result)

// PUT 请求
err := client.PutJSON("/api/data/1", data, &result)

// DELETE 请求
err := client.DeleteJSON("/api/data/1", &result)

// 获取原始响应
resp, err := client.Get("/api/users")
resp, err := client.Post("/api/data", data)
```

### 调试签名

```go
// 输出签名调试信息
client.DebugSign(requestBody)
```

## 签名相关

### 生成签名

```go
// 快速生成签名
params, sign, err := goauth.QuickSign(appID, appSecret, body)

// 构建签名参数
params, err := goauth.BuildSignParams(appID, body)

// 生成签名
sign := goauth.GenerateSign(params, appSecret)

// 验证签名
valid := goauth.VerifySign(params, appSecret, sign)

// 验证签名并获取调试信息
valid, debugInfo := goauth.VerifySignWithDebug(params, appSecret, sign)
```

### 生成随机字符串

```go
// 简单随机字符串
nonce := goauth.GenerateNonce(16)

// 加密安全的随机字符串
nonce, err := goauth.GenerateSecureNonce(16)
```

## 工具函数

```go
// 生成应用密钥
secret, err := goauth.GenerateAppSecret(32)

// 遮蔽密钥（用于日志）
masked := goauth.MaskSecret(secret)

// 验证应用ID
err := goauth.ValidateAppID("my-app-123")

// 生成请求ID
requestID := goauth.BuildRequestID()

// 清理IP地址
ip := goauth.SanitizeIP("192.168.1.1:8080")
```

## 错误处理

```go
// 预定义错误
goauth.ErrInvalidParams
goauth.ErrAppNotFound
goauth.ErrAppDisabled
goauth.ErrInvalidTimestamp
goauth.ErrInvalidSign
goauth.ErrIPNotAllowed
goauth.ErrRateLimitExceeded

// 创建错误
err := goauth.NewAuthError(
    goauth.ErrCodeInvalidSign,
    "签名验证失败",
    "详细信息",
    http.StatusUnauthorized,
)

// 创建错误响应
response := goauth.NewErrorResponse(err, requestID)

// 创建成功响应
response := goauth.NewSuccessResponse(data, requestID)
```

## 配置文件格式

### YAML

```yaml
timestamp_tolerance: 300
default_rate_limit: 1000
enable_ip_check: true

apps:
  app-id:
    app_id: app-id
    app_secret: app-secret
    app_name: 应用名
    require_sign: true
    enabled: true
    rate_limit: 100
    ip_whitelist:
      - "*"
```

### JSON

```json
{
  "timestamp_tolerance": 300,
  "default_rate_limit": 1000,
  "enable_ip_check": true,
  "apps": {
    "app-id": {
      "app_id": "app-id",
      "app_secret": "app-secret",
      "app_name": "应用名",
      "require_sign": true,
      "enabled": true,
      "rate_limit": 100,
      "ip_whitelist": ["*"]
    }
  }
}
```

## 完整示例

### 服务端

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/difyz9/go-auth"
)

func main() {
    r := gin.Default()
    
    // 创建配置并应用中间件
    config := goauth.QuickConfig("my-app", "my-secret")
    r.Use(goauth.New(config).Authenticate())
    
    // 定义路由
    r.GET("/api/users", func(c *gin.Context) {
        app, _ := goauth.GetAppFromContext(c)
        c.JSON(200, gin.H{
            "app": app.AppName,
            "users": []string{"Alice", "Bob"},
        })
    })
    
    r.Run(":8080")
}
```

### 客户端

```go
package main

import (
    "fmt"
    "github.com/difyz9/go-auth"
)

func main() {
    client := goauth.NewClient(
        "http://localhost:8080",
        "my-app",
        "my-secret",
    )
    
    var result map[string]interface{}
    if err := client.GetJSON("/api/users", &result); err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    fmt.Printf("Result: %+v\n", result)
}
```

## 常见场景

### 场景1：开发环境快速测试

```go
config := goauth.QuickConfig("test", "test-secret")
r.Use(goauth.New(config).Authenticate())
```

### 场景2：多应用管理

```go
config := goauth.NewConfigBuilder().
    AddSimpleApp("app1", "secret1", "应用1").
    AddSimpleApp("app2", "secret2", "应用2").
    AddSimpleApp("app3", "secret3", "应用3").
    MustBuild()
```

### 场景3：生产环境

```go
config := goauth.MustLoadYAML("config.yaml")
auth := goauth.New(config).WithLogger(productionLogger)
r.Use(auth.Authenticate(productionErrorHandler))
```

### 场景4：动态更新配置

```go
// 添加新应用
config.AddApp(newApp)

// 保存到文件
config.SaveToYAML("config.yaml")

// 重新加载（需要重启服务）
config = goauth.MustLoadYAML("config.yaml")
```

## 性能提示

- ✅ 配置对象可以重用，无需每次请求创建
- ✅ 中间件是线程安全的
- ✅ 速率限制器自动清理过期数据
- ✅ 签名验证性能优秀（~30μs/op）

## 安全建议

- 🔐 使用至少32位的强随机密钥
- 🔐 生产环境配置具体的IP白名单
- 🔐 启用HTTPS传输
- 🔐 定期轮换密钥
- 🔐 不要在版本控制中提交密钥

## 故障排查

### 签名验证失败

```go
// 使用调试模式
client := goauth.NewClient(url, id, secret, goauth.WithDebug(true))

// 或手动调试签名
client.DebugSign(body)

// 服务端验证并输出调试信息
valid, info := goauth.VerifySignWithDebug(params, secret, sign)
fmt.Println(info)
```

### 时间戳错误

- 检查服务器时间是否同步
- 调整 `timestamp_tolerance` 参数
- 使用 NTP 同步时间

### IP白名单问题

- 检查客户端IP是否正确
- 使用 `*` 临时允许所有IP调试
- 支持通配符：`192.168.1.*`

## 更多资源

- 📖 完整文档：[README.md](README.md)
- 🚀 快速开始：[QUICK_START.md](QUICK_START.md)
- 📊 项目总结：[PROJECT_SUMMARY.md](PROJECT_SUMMARY.md)
- ✨ 优化说明：[OPTIMIZATION_SUMMARY.md](OPTIMIZATION_SUMMARY.md)
- 💡 示例代码：[examples/](examples/)
