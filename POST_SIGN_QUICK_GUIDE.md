# POST请求签名优化 - 快速指南

## 问题背景

**旧版本问题**：
- POST请求签名包含完整请求体（`requestBody`参数）
- JSON序列化不确定性导致签名可能失败
- 大型请求体导致签名字符串过长
- 实现复杂，调试困难

**新版本优化**：
- ✅ 默认不包含请求体（更灵活、简单）
- ✅ 可配置是否包含请求体
- ✅ 支持全局和应用级别配置
- ✅ 向后兼容

---

## 快速使用

### 1. 推荐方式（默认，不包含请求体）

#### 服务端配置

```yaml
# goauth_config.yaml
sign_include_body: false  # 推荐：默认false

apps:
  test-app-001:
    app_secret: "your-secret"
    enabled: true
    require_sign: true
    # 使用全局配置（不包含请求体）
```

#### 客户端代码

```go
// 创建客户端（默认不包含请求体）
client := goauth.NewClient(
    "http://localhost:8089",
    "test-app-001",
    "your-secret",
)

// POST请求 - 签名只包含 appId, timestamp, nonce
resp, err := client.Post("/api/test", map[string]interface{}{
    "text":  "Hello",
    "voice": "zh-CN",
})
```

### 2. 高安全方式（包含请求体）

#### 服务端配置

```yaml
sign_include_body: false  # 全局默认false

apps:
  # 高安全应用（覆盖全局配置）
  secure-app:
    app_secret: "secure-secret"
    enabled: true
    require_sign: true
    sign_include_body: true  # 覆盖：启用请求体签名
```

#### 客户端代码

```go
// 创建高安全客户端（包含请求体）
client := goauth.NewClient(
    "http://localhost:8089",
    "secure-app",
    "secure-secret",
    goauth.WithSignIncludeBody(true),  // 启用请求体签名
)

// POST请求 - 签名包含完整请求体
resp, err := client.Post("/api/payment", map[string]interface{}{
    "amount": 1000.00,
})
```

---

## 配置说明

### 全局配置

```go
config := goauth.NewConfig(
    goauth.WithSignIncludeBody(false),  // 全局设置
)
```

| 配置值 | 说明 | 推荐场景 |
|--------|------|----------|
| `false` | 不包含请求体（默认） | 大多数场景 |
| `true` | 包含请求体 | 高安全需求 |

### 应用级别配置

```yaml
apps:
  app1:
    # 不设置 sign_include_body = 使用全局配置
    
  app2:
    sign_include_body: true  # 覆盖全局配置
```

```go
includeBody := true
config.AddApp(&goauth.AppConfig{
    AppID:           "app2",
    SignIncludeBody: &includeBody,  // 覆盖全局配置
})
```

---

## 签名对比

### 不包含请求体（推荐）

**签名参数**：
```
appId=test-app-001&nonce=abc123&timestamp=1732089600
```

**优点**：
- ✅ 简单、清晰
- ✅ 不受JSON序列化影响
- ✅ 签名字符串短
- ✅ 易于调试

**安全性**：
- HMAC-SHA256签名（防篡改认证参数）
- 时间戳验证（防重放）
- Nonce随机值（增强安全性）
- HTTPS传输（防中间人）

### 包含请求体

**签名参数**：
```
appId=test-app-001&nonce=abc123&requestBody={"text":"Hello"}&timestamp=1732089600
```

**优点**：
- ✅ 验证请求体完整性
- ✅ 更高安全性

**缺点**：
- ⚠️ 需要JSON序列化一致性
- ⚠️ 签名字符串可能很长
- ⚠️ 实现复杂度高

---

## 迁移指南

### 从旧版本迁移

#### 1. 更新配置文件

```yaml
# 添加此行（使用默认推荐配置）
sign_include_body: false
```

#### 2. 更新客户端代码

**旧代码**（自动包含请求体）：
```go
client := goauth.NewClient(baseURL, appID, appSecret)
// POST请求会自动包含请求体在签名中
```

**新代码**（明确配置）：
```go
// 推荐方式（不包含请求体）
client := goauth.NewClient(
    baseURL, appID, appSecret,
    goauth.WithSignIncludeBody(false),  // 明确设置
)

// 或者高安全方式（包含请求体）
client := goauth.NewClient(
    baseURL, appID, appSecret,
    goauth.WithSignIncludeBody(true),
)
```

#### 3. 测试验证

```go
// 启用调试模式
client := goauth.NewClient(
    baseURL, appID, appSecret,
    goauth.WithSignIncludeBody(false),
    goauth.WithDebug(true),  // 查看签名过程
)

// 查看签名参数
client.DebugSign(requestBody)
```

---

## 调试技巧

### 1. 启用调试模式

```go
client := goauth.NewClient(
    baseURL, appID, appSecret,
    goauth.WithSignIncludeBody(false),
    goauth.WithDebug(true),  // 启用调试
)
```

**输出示例**：
```
[GoAuth Client] Request: POST /api/test
[GoAuth Client] AppID: test-app-001
[GoAuth Client] Timestamp: 1732089600
[GoAuth Client] Nonce: abc123xyz456
[GoAuth Client] SignIncludeBody: false  // 关键：是否包含请求体
[GoAuth Client] Sign: 1a2b3c4d5e6f...
[GoAuth Client] Body: {"text":"Hello"}
```

### 2. 手动调试签名

```go
client.DebugSign(map[string]interface{}{
    "text": "Hello",
})
```

**输出**：
```
=== 签名调试信息 ===
AppID:     test-app-001
AppSecret: tmcf5m6qcm6k9hrp3sy8rhgafu00ttph
Timestamp: 1732089600
Nonce:     abc123xyz456

签名参数:
  appId: test-app-001
  timestamp: 1732089600
  nonce: abc123xyz456
  // 注意：没有 requestBody 参数

生成的签名: 1a2b3c4d5e6f...
==================
```

### 3. 对比签名

```go
// 测试不包含请求体
params1, sign1, _ := goauth.QuickSign(appID, appSecret, body, false)
fmt.Printf("不包含请求体 - 签名: %s\n", sign1)

// 测试包含请求体
params2, sign2, _ := goauth.QuickSign(appID, appSecret, body, true)
fmt.Printf("包含请求体 - 签名: %s\n", sign2)
```

---

## 最佳实践

### 1. 默认推荐配置

```yaml
# 全局配置
sign_include_body: false  # 推荐

apps:
  # 大多数应用使用默认配置
  normal-app:
    app_secret: "secret1"
    require_sign: true
```

**原因**：
- 简单、灵活
- 避免JSON序列化问题
- 足够的安全性（HMAC + 时间戳 + Nonce）

### 2. 针对敏感接口启用

```yaml
apps:
  # 支付相关应用启用请求体签名
  payment-app:
    app_secret: "secret2"
    require_sign: true
    sign_include_body: true  # 高安全
```

### 3. 混合使用

```go
// 普通客户端
normalClient := goauth.NewClient(
    baseURL, appID1, secret1,
    goauth.WithSignIncludeBody(false),
)

// 支付客户端
paymentClient := goauth.NewClient(
    baseURL, appID2, secret2,
    goauth.WithSignIncludeBody(true),
)

// 根据场景选择客户端
normalClient.Post("/api/data", data)      // 普通请求
paymentClient.Post("/api/payment", payment)  // 支付请求
```

---

## 常见问题

### Q1: 需要修改现有代码吗？

**A**: 不需要，向后兼容。

- 旧代码可以继续使用
- 默认配置（`sign_include_body: false`）让POST请求更容易成功
- 可以逐步迁移

### Q2: 不包含请求体安全吗？

**A**: 对于大多数场景足够安全。

已有的安全机制：
- ✅ HMAC-SHA256签名（防篡改）
- ✅ 时间戳验证（防重放）
- ✅ Nonce随机值
- ✅ HTTPS传输

建议：
- 普通接口：使用 `sign_include_body: false`
- 敏感接口（支付、重要操作）：使用 `sign_include_body: true`

### Q3: 如何选择配置？

| 场景 | 推荐配置 | 说明 |
|------|---------|------|
| 普通数据接口 | `false` | 简单、灵活 |
| 查询接口 | `false` | 无需验证请求体 |
| 支付接口 | `true` | 高安全需求 |
| 敏感数据修改 | `true` | 验证请求完整性 |
| 大型请求体 | `false` | 避免签名字符串过长 |
| 多语言客户端 | `false` | 避免JSON序列化问题 |

### Q4: 为什么POST请求401？

**可能原因**：
1. 服务端配置 `sign_include_body: true`，客户端配置 `false`
2. 反之亦然

**解决方法**：
- 确保服务端和客户端配置一致
- 启用调试模式查看签名参数
- 使用 `DebugSign()` 对比签名

---

## 总结

✅ **推荐配置**：
```yaml
sign_include_body: false  # 默认false
```

✅ **推荐客户端**：
```go
client := goauth.NewClient(baseURL, appID, appSecret)
// 默认 SignIncludeBody = false
```

✅ **优势**：
- 简单易用
- 避免JSON序列化问题
- 更好的兼容性
- 足够的安全性

🔒 **高安全场景**：
- 针对敏感接口启用 `sign_include_body: true`
- 应用级别配置覆盖
- 灵活混合使用

---

**更新日期**: 2025-11-20  
**版本**: v0.0.3
