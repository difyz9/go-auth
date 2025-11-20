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
- ✅ **POST签名优化** - 可配置是否包含请求体（v0.0.3新增）🆕
- ✅ **请求体加解密** - 支持 AES 加密传输
- ✅ **响应体加密** - 可选的响应数据加密
- ✅ **可扩展** - 支持自定义日志和错误处理

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

### 3. 链式调用创建中间件

```go
// 使用链式调用配置中间件
auth := goauth.New(config).
    WithLogger(&MyLogger{})

r.Use(auth.Authenticate())
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

### 8. 请求体和响应体加解密 🔐

#### 启用加解密中间件

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/difyz9/go-auth"
)

func main() {
    r := gin.Default()
    
    // 创建认证配置
    config := goauth.QuickConfig("my-app", "my-secret")
    
    // 创建加解密配置
    cryptoConfig := goauth.NewCryptoConfig("my-aes-key-32-bytes-long-123456")
    cryptoConfig.EnableDecryption = true  // 启用请求解密
    cryptoConfig.EnableEncryption = true  // 启用响应加密
    cryptoConfig.ForceEncryption = false  // 不强制加密（由客户端控制）
    
    // 创建加解密中间件
    cryptoMiddleware := goauth.NewCryptoMiddleware(cryptoConfig)
    
    // 应用中间件（顺序很重要！）
    r.Use(cryptoMiddleware.DecryptRequest())   // 1. 先解密请求
    r.Use(goauth.New(config).Authenticate())    // 2. 然后认证
    r.Use(cryptoMiddleware.EncryptResponse())   // 3. 最后加密响应
    
    r.POST("/api/users", handleUsers)
    r.Run(":8080")
}
```

#### 客户端发送加密请求

```go
// 准备数据
userData := map[string]interface{}{
    "name":  "张三",
    "email": "zhangsan@example.com",
}

// 序列化
userDataJSON, _ := json.Marshal(userData)

// AES加密
encryptedData, _ := goauth.AESEncrypt(aesKey, userDataJSON)

// 构建加密请求
encryptedRequest := goauth.EncryptedRequest{
    Data:      encryptedData,
    Timestamp: time.Now().Unix(),
}

// 发送请求（添加加密标识头）
headers := map[string]string{
    "X-Encrypted": "1",
}
resp, err := client.Post("/api/users", encryptedRequest, headers)
```

#### 客户端请求加密响应

```go
// 请求加密响应
headers := map[string]string{
    "X-Response-Encrypt": "1",
}

resp, err := client.Get("/api/users/123", headers)

// 解析加密响应
var encryptedResp goauth.EncryptedResponse
json.NewDecoder(resp.Body).Decode(&encryptedResp)

// 解密
decryptedData, err := goauth.AESDecrypt(aesKey, encryptedResp.Data)

// 解析数据
var result map[string]interface{}
json.Unmarshal(decryptedData, &result)
```

#### 完整的加密通信示例

```go
// 同时加密请求和响应
headers := map[string]string{
    "X-Encrypted":        "1", // 请求已加密
    "X-Response-Encrypt": "1", // 要求响应加密
}

// 发送加密请求
resp, err := client.Post("/api/orders", encryptedRequest, headers)

// 服务端会：
// 1. 解密请求体
// 2. 验证签名和权限
// 3. 处理业务逻辑
// 4. 加密响应体
// 5. 返回加密响应
```

#### 加解密配置选项

```go
cryptoConfig := goauth.NewCryptoConfig("your-aes-key")

// 配置选项
cryptoConfig.Enabled = true                    // 是否启用加解密
cryptoConfig.EnableDecryption = true           // 是否启用请求解密
cryptoConfig.EnableEncryption = true           // 是否启用响应加密
cryptoConfig.ForceEncryption = false           // 是否强制加密所有响应
cryptoConfig.SkipPaths = []string{"/health"}   // 跳过加解密的路径
cryptoConfig.EncryptionHeader = "X-Encrypted"  // 加密标识头
cryptoConfig.ResponseEncHeader = "X-Response-Encrypt" // 响应加密请求头
```

#### 检查请求是否被解密

```go
func handleUsers(c *gin.Context) {
    // 检查请求是否经过解密
    wasDecrypted, _ := c.Get("crypto_decrypted")
    originalSize, _ := c.Get("crypto_original_size")
    decryptedSize, _ := c.Get("crypto_decrypted_size")
    
    if wasDecrypted.(bool) {
        fmt.Printf("请求已解密: %d -> %d 字节\n", originalSize, decryptedSize)
    }
    
    // 处理业务逻辑...
}
```

#### AES 加解密函数

```go
// AES 加密
encryptedData, err := goauth.AESEncrypt(key, plaintext)

// AES 解密
decryptedData, err := goauth.AESDecrypt(key, encryptedData)
```

**注意事项**：
- AES 密钥长度应为 16、24 或 32 字节（对应 AES-128、AES-192、AES-256）
- 使用 CBC 模式 + PKCS7 填充
- 密文使用 Base64 编码传输
- 中间件顺序很重要：解密 -> 认证 -> 业务 -> 加密
- 签名是对加密后的数据进行的

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

### POST请求签名优化 🆕

**v0.0.3 新增**：可配置POST请求签名是否包含请求体

```yaml
# 全局配置（推荐：默认不包含请求体）
sign_include_body: false

apps:
  # 普通应用（使用全局配置）
  normal-app:
    app_secret: "secret1"
    
  # 高安全应用（覆盖全局配置）
  secure-app:
    app_secret: "secret2"
    sign_include_body: true  # 包含请求体
```

**客户端配置**：
```go
// 默认方式（不包含请求体，推荐）
client := goauth.NewClient(baseURL, appID, appSecret)

// 高安全方式（包含请求体）
client := goauth.NewClient(baseURL, appID, appSecret,
    goauth.WithSignIncludeBody(true))
```

**详细说明**：参见 [POST_SIGN_QUICK_GUIDE.md](./POST_SIGN_QUICK_GUIDE.md)

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

### Q: POST请求出现401错误怎么办？ 🆕

A: 从 v0.0.3 开始，默认签名不包含请求体。确保：
```yaml
# 服务端配置
sign_include_body: false  # 推荐配置
```
```go
// 客户端配置
client := goauth.NewClient(baseURL, appID, appSecret)
// 默认 SignIncludeBody = false
```

详见：[POST签名优化指南](./POST_SIGN_QUICK_GUIDE.md)

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
5. 启用客户端调试模式：`goauth.WithDebug(true)`
6. 使用 `client.DebugSign(body)` 查看签名过程

## 完整示例

查看 `example.go` 文件获取完整的服务端和客户端示例代码。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
