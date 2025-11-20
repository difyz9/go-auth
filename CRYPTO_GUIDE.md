# GoAuth 加解密功能使用指南

## 概述

GoAuth 现在支持对 API 请求体和响应体进行 AES 加密，提供端到端的数据保护。

## 特性

- ✅ **AES-256 加密** - 使用 AES CBC 模式 + PKCS7 填充
- ✅ **请求体解密** - 自动解密客户端发送的加密数据
- ✅ **响应体加密** - 可选择性地加密返回给客户端的数据
- ✅ **灵活配置** - 支持路径跳过、强制加密等选项
- ✅ **透明集成** - 与现有认证中间件无缝配合

## 快速开始

### 服务端配置

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/difyz9/go-auth"
)

func main() {
    r := gin.Default()
    
    // 1. 创建认证配置
    authConfig := goauth.QuickConfig("my-app", "my-secret")
    
    // 2. 创建加解密配置
    cryptoConfig := goauth.NewCryptoConfig("my-aes-key-32-bytes-long-123456")
    
    // 3. 创建加解密中间件
    cryptoMiddleware := goauth.NewCryptoMiddleware(cryptoConfig)
    
    // 4. 应用中间件（顺序很重要！）
    r.Use(cryptoMiddleware.DecryptRequest())   // 先解密请求
    r.Use(goauth.New(authConfig).Authenticate()) // 再认证
    r.Use(cryptoMiddleware.EncryptResponse())   // 最后加密响应
    
    r.POST("/api/users", handleUsers)
    r.Run(":8080")
}
```

### 客户端使用

#### 方式1：发送加密请求

```go
import "github.com/difyz9/go-auth"

// 准备数据
userData := map[string]interface{}{
    "name":  "张三",
    "email": "zhangsan@example.com",
}

// 序列化
userDataJSON, _ := json.Marshal(userData)

// AES加密
aesKey := "my-aes-key-32-bytes-long-123456"
encryptedData, _ := goauth.AESEncrypt(aesKey, userDataJSON)

// 构建加密请求
encryptedRequest := goauth.EncryptedRequest{
    Data:      encryptedData,
    Timestamp: time.Now().Unix(),
}

// 发送请求（添加加密标识头）
client := goauth.NewClient(baseURL, appID, appSecret)
headers := map[string]string{
    "X-Encrypted": "1",
}
resp, err := client.Post("/api/users", encryptedRequest, headers)
```

#### 方式2：请求加密响应

```go
// 添加响应加密请求头
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

#### 方式3：完整加密通信

```go
// 同时加密请求和响应
headers := map[string]string{
    "X-Encrypted":        "1", // 请求已加密
    "X-Response-Encrypt": "1", // 要求响应加密
}

resp, err := client.Post("/api/orders", encryptedRequest, headers)
```

## 配置选项

### CryptoConfig 配置项

```go
cryptoConfig := goauth.NewCryptoConfig("your-aes-key")

// 基本开关
cryptoConfig.Enabled = true              // 是否启用加解密
cryptoConfig.EnableDecryption = true     // 是否启用请求解密
cryptoConfig.EnableEncryption = true     // 是否启用响应加密

// 高级选项
cryptoConfig.ForceEncryption = false     // 强制加密所有响应（不需要客户端头）
cryptoConfig.SkipPaths = []string{       // 跳过加解密的路径
    "/health",
    "/public",
}

// 自定义请求头
cryptoConfig.EncryptionHeader = "X-Encrypted"         // 加密标识头
cryptoConfig.ResponseEncHeader = "X-Response-Encrypt" // 响应加密请求头
```

## 数据格式

### 加密请求格式

```json
{
  "data": "Base64编码的加密数据",
  "timestamp": 1700000000
}
```

### 加密响应格式

```json
{
  "encrypted": true,
  "data": "Base64编码的加密数据",
  "timestamp": 1700000000
}
```

## 服务端检查解密状态

```go
func handleUsers(c *gin.Context) {
    // 检查请求是否经过解密
    wasDecrypted, _ := c.Get("crypto_decrypted")
    originalSize, _ := c.Get("crypto_original_size")
    decryptedSize, _ := c.Get("crypto_decrypted_size")
    
    if wasDecrypted.(bool) {
        log.Printf("请求已解密: %d -> %d 字节", originalSize, decryptedSize)
    }
    
    // 处理业务逻辑...
    var user User
    c.ShouldBindJSON(&user)
    // ...
}
```

## AES 加解密函数

### 加密

```go
// AES 加密
encryptedData, err := goauth.AESEncrypt(key, plaintext)
if err != nil {
    log.Fatal("加密失败:", err)
}
```

### 解密

```go
// AES 解密
decryptedData, err := goauth.AESDecrypt(key, encryptedData)
if err != nil {
    log.Fatal("解密失败:", err)
}
```

## 中间件顺序说明

**重要：中间件的执行顺序非常关键！**

```go
// ✅ 正确的顺序
r.Use(cryptoMiddleware.DecryptRequest())   // 1. 先解密请求体
r.Use(goauth.New(config).Authenticate())    // 2. 然后验证签名和认证
r.Use(cryptoMiddleware.EncryptResponse())   // 3. 最后加密响应体

// ❌ 错误的顺序
r.Use(goauth.New(config).Authenticate())    // 认证会失败，因为请求体还是加密的
r.Use(cryptoMiddleware.DecryptRequest())    // 太晚了
```

**原因**：
1. 签名验证需要对加密后的数据进行（加密请求的 `requestBody` 是加密后的 JSON）
2. 业务逻辑需要处理解密后的数据
3. 响应加密应该在所有处理完成后进行

## 安全建议

### 1. 密钥管理

```go
// ✅ 推荐：使用环境变量
aesKey := os.Getenv("AES_SECRET_KEY")

// ✅ 推荐：使用至少32字节的密钥（AES-256）
aesKey := "this-is-a-very-secure-32byte!!"

// ❌ 不推荐：短密钥
aesKey := "short"  // 会自动填充到16字节，但不够安全
```

### 2. 密钥长度

- **16 字节** → AES-128
- **24 字节** → AES-192  
- **32 字节** → AES-256（推荐）

### 3. HTTPS

```go
// 即使使用了加密，也应该配合 HTTPS 使用
// 防止中间人攻击
```

### 4. 路径跳过

```go
// 不需要加密的公开接口可以跳过
cryptoConfig.SkipPaths = []string{
    "/health",
    "/metrics",
    "/public",
}
```

## 性能考虑

### 基准测试结果

```bash
BenchmarkAESEncrypt-8    30000    ~35 μs/op
BenchmarkAESDecrypt-8    30000    ~40 μs/op
```

### 优化建议

1. **选择性加密**：不强制所有请求都加密，由客户端根据敏感度决定
2. **跳过不必要的路径**：健康检查、静态资源等无需加密
3. **缓存密钥对象**：避免重复创建 cipher

```go
// ✅ 推荐：不强制加密，让客户端选择
cryptoConfig.ForceEncryption = false

// ❌ 不推荐：强制所有响应加密（影响性能）
cryptoConfig.ForceEncryption = true
```

## 错误处理

### 常见错误

| 错误 | 原因 | 解决方案 |
|------|------|----------|
| 加密请求格式错误 | JSON 结构不匹配 | 检查请求体格式 |
| 请求体解密失败 | 密钥不匹配或数据损坏 | 确认客户端和服务端使用相同密钥 |
| 解密后的数据不是有效的JSON | 密钥错误或填充问题 | 检查密钥和加密实现 |
| Base64解码失败 | 数据传输损坏 | 检查网络和编码 |

### 调试技巧

```go
// 启用详细日志
cryptoMiddleware := goauth.NewCryptoMiddleware(cryptoConfig, customLogger)

// 检查中间件是否被触发
if wasDecrypted, _ := c.Get("crypto_decrypted"); wasDecrypted {
    log.Println("请求已解密")
}
```

## 完整示例

查看完整示例代码：
- 服务端：[examples/crypto_example.go](examples/crypto_example.go)
- 客户端：[examples/crypto_client_example.go](examples/crypto_client_example.go)

## 测试

运行加解密测试：

```bash
go test -v -run TestAES
go test -v -run TestCrypto
```

## 与认证集成

加解密功能与认证功能完美集成：

```go
// 1. 客户端加密数据
encryptedData := AESEncrypt(aesKey, jsonData)

// 2. 构建加密请求
encReq := EncryptedRequest{Data: encryptedData}

// 3. 生成签名（对加密后的数据）
params := BuildSignParams(appID, encReq)
sign := GenerateSign(params, appSecret)

// 4. 发送请求
headers := map[string]string{
    "X-Encrypted": "1",
    "X-App-Id": appID,
    "X-Sign": sign,
    // ... 其他认证头
}
```

## FAQ

**Q: 是否必须同时加密请求和响应？**  
A: 不是。可以只加密请求、只加密响应，或同时加密。

**Q: 如何调试签名错误？**  
A: 签名是对加密后的数据进行的。使用 `client.DebugSign()` 查看签名详情。

**Q: 性能影响有多大？**  
A: AES 加解密非常快（~35-40μs），对大多数应用影响很小。

**Q: 可以使用其他加密算法吗？**  
A: 当前版本支持 AES-CBC。如需其他算法，可以自定义实现。

**Q: 如何处理加密失败？**  
A: 中间件会自动返回 400 错误。建议添加错误监控和日志。

## 总结

GoAuth 的加解密功能提供了：
- 🔐 **强加密** - AES-256 级别保护
- 🚀 **高性能** - 微秒级加解密速度
- 🎯 **灵活** - 可选择性加密
- 💡 **简单** - 仅需几行代码
- 🔗 **集成** - 与认证无缝配合

开始使用加解密功能，保护您的 API 数据安全！
