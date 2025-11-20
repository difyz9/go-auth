# POST请求签名优化方案

## 问题描述

当前POST请求签名机制将完整的请求体（`requestBody`）包含在签名参数中，这可能导致：

1. **JSON序列化不确定性**：不同的JSON序列化库可能产生不同的字符串（字段顺序、空格等）
2. **签名复杂度高**：大型请求体导致签名字符串过长
3. **兼容性问题**：某些客户端可能难以实现完全一致的请求体签名

## 优化方案

### 方案1：提供配置选项（推荐）⭐

**优点**：
- 向后兼容
- 灵活配置
- 适应不同安全需求

**实现**：添加配置项控制是否在签名中包含请求体

```go
type Config struct {
    // ... 现有配置 ...
    
    // SignIncludeBody 签名是否包含请求体
    // true: 签名包含完整请求体（更安全，但要求JSON序列化一致）
    // false: 签名只包含基础参数（更灵活，推荐）
    SignIncludeBody bool `yaml:"sign_include_body" json:"sign_include_body"`
}
```

### 方案2：使用请求体摘要

**优点**：
- 验证请求体完整性
- 不依赖JSON序列化顺序
- 签名字符串长度固定

**实现**：使用SHA256摘要代替完整请求体

```go
// 不包含完整body，使用摘要
if len(bodyBytes) > 0 {
    hash := sha256.Sum256(bodyBytes)
    params["bodyHash"] = hex.EncodeToString(hash[:])
}
```

### 方案3：分离签名和请求体验证

**优点**：
- 签名只验证认证参数
- 另外验证请求体完整性（如需要）
- 职责清晰

**实现**：
```go
// 签名只包含基础参数
params := map[string]string{
    "appId":     appID,
    "timestamp": timestamp,
    "nonce":     nonce,
    // 不包含 requestBody
}

// 可选：通过单独的Header传递请求体摘要
if needBodyValidation {
    bodyHash := sha256.Sum256(bodyBytes)
    headers["X-Body-Hash"] = hex.EncodeToString(bodyHash[:])
}
```

## 推荐实现

### 1. 添加配置选项（默认不包含请求体）

```yaml
# config.yaml
sign_include_body: false  # 默认false，更灵活

apps:
  - app_id: test-app-001
    app_secret: your-secret
    enabled: true
    require_sign: true
    sign_include_body: false  # 应用级别配置
```

### 2. 更新代码实现

#### Config结构体
```go
type Config struct {
    Apps               []AppConfig `yaml:"apps" json:"apps"`
    TimestampTolerance int64       `yaml:"timestamp_tolerance" json:"timestamp_tolerance"`
    EnableIPCheck      bool        `yaml:"enable_ip_check" json:"enable_ip_check"`
    SignIncludeBody    bool        `yaml:"sign_include_body" json:"sign_include_body"` // 新增
}

type AppConfig struct {
    AppID           string   `yaml:"app_id" json:"app_id"`
    AppSecret       string   `yaml:"app_secret" json:"app_secret"`
    Enabled         bool     `yaml:"enabled" json:"enabled"`
    RequireSign     bool     `yaml:"require_sign" json:"require_sign"`
    SignIncludeBody *bool    `yaml:"sign_include_body,omitempty" json:"sign_include_body,omitempty"` // 新增，可选覆盖全局配置
    IPWhitelist     []string `yaml:"ip_whitelist" json:"ip_whitelist"`
    RateLimit       int      `yaml:"rate_limit" json:"rate_limit"`
}
```

#### 中间件逻辑
```go
func (m *AuthMiddleware) extractRequestParams(c *gin.Context, appID, timestamp, nonce string, app *AppConfig) (map[string]string, error) {
    params := map[string]string{
        "appId":     appID,
        "timestamp": timestamp,
        "nonce":     nonce,
    }

    // 确定是否包含请求体
    includeBody := m.config.SignIncludeBody
    if app.SignIncludeBody != nil {
        includeBody = *app.SignIncludeBody // 应用级别配置覆盖全局配置
    }

    // 获取请求体参数（如果配置启用）
    if includeBody && (c.Request.Method == "POST" || c.Request.Method == "PUT") {
        contentType := c.GetHeader("Content-Type")
        if strings.Contains(contentType, "application/json") {
            bodyBytes, err := c.GetRawData()
            if err != nil {
                return nil, err
            }

            if len(bodyBytes) > 0 {
                params["requestBody"] = string(bodyBytes)
                c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
            }
        }
    }

    // 获取查询参数
    for k, v := range c.Request.URL.Query() {
        if len(v) > 0 && k != "sign" {
            params[k] = v[0]
        }
    }

    return params, nil
}
```

#### 客户端逻辑
```go
type Client struct {
    BaseURL         string
    AppID           string
    AppSecret       string
    HTTPClient      *http.Client
    Debug           bool
    SignIncludeBody bool // 新增配置项
}

func NewClient(baseURL, appID, appSecret string, opts ...ClientOption) *Client {
    client := &Client{
        BaseURL:         baseURL,
        AppID:           appID,
        AppSecret:       appSecret,
        HTTPClient:      &http.Client{Timeout: 30 * time.Second},
        Debug:           false,
        SignIncludeBody: false, // 默认不包含请求体
    }
    
    for _, opt := range opts {
        opt(client)
    }
    
    return client
}

// WithSignIncludeBody 配置是否在签名中包含请求体
func WithSignIncludeBody(include bool) ClientOption {
    return func(c *Client) {
        c.SignIncludeBody = include
    }
}

func (c *Client) Request(method, path string, body interface{}, headers ...map[string]string) (*http.Response, error) {
    var bodyBytes []byte
    var err error
    
    if body != nil {
        bodyBytes, err = json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("序列化请求体失败: %w", err)
        }
    }
    
    // 生成认证参数
    timestamp := strconv.FormatInt(time.Now().Unix(), 10)
    nonce := GenerateNonce(16)
    
    // 构建签名参数
    params := map[string]string{
        "appId":     c.AppID,
        "timestamp": timestamp,
        "nonce":     nonce,
    }
    
    // 根据配置决定是否包含请求体
    if c.SignIncludeBody && len(bodyBytes) > 0 {
        params["requestBody"] = string(bodyBytes)
    }
    
    // 生成签名
    sign := GenerateSign(params, c.AppSecret)
    
    // ... 其余逻辑保持不变 ...
}
```

## 使用示例

### 服务端配置

```yaml
# 全局默认不包含请求体
sign_include_body: false

apps:
  # 普通应用（不包含请求体）
  - app_id: test-app-001
    app_secret: secret1
    enabled: true
    require_sign: true
    # 不设置 sign_include_body，使用全局配置

  # 高安全应用（包含请求体）
  - app_id: secure-app-002
    app_secret: secret2
    enabled: true
    require_sign: true
    sign_include_body: true  # 覆盖全局配置
```

### 客户端使用

```go
// 默认方式（不包含请求体，推荐）
client := goauth.NewClient(
    "http://localhost:8089",
    "test-app-001",
    "secret1",
)

// POST请求正常工作
resp, err := client.Post("/api/v1/tts", map[string]interface{}{
    "text": "Hello",
    "voice": "zh-CN",
})

// 高安全模式（包含请求体）
secureClient := goauth.NewClient(
    "http://localhost:8089",
    "secure-app-002",
    "secret2",
    goauth.WithSignIncludeBody(true), // 明确启用
)
```

## 迁移指南

### 对现有代码的影响

**如果保持默认配置（`sign_include_body: false`）**：
- ✅ POST请求更容易调试
- ✅ 不受JSON序列化影响
- ✅ 兼容更多客户端
- ⚠️ 不验证请求体完整性

**如果启用请求体签名（`sign_include_body: true`）**：
- ✅ 验证请求体完整性
- ✅ 更高安全性
- ⚠️ 需要确保JSON序列化一致性

### 平滑迁移步骤

1. **更新服务端配置**
```yaml
sign_include_body: false  # 全局设为false
```

2. **更新客户端**
```go
// 默认不包含请求体
client := goauth.NewClient(baseURL, appID, appSecret)
```

3. **针对敏感接口启用**
```go
// 为特定应用启用请求体签名
secureClient := goauth.NewClient(
    baseURL, 
    appID, 
    appSecret,
    goauth.WithSignIncludeBody(true),
)
```

## 最佳实践

### 1. 默认推荐配置

```yaml
# 大多数场景推荐配置
sign_include_body: false
```

**原因**：
- 更简单、更灵活
- 避免JSON序列化问题
- 时间戳和nonce已提供足够的防重放保护

### 2. 何时启用请求体签名

启用 `sign_include_body: true` 的场景：
- 金融交易接口
- 支付相关接口
- 敏感数据修改接口
- 需要严格验证请求体完整性的场景

### 3. 性能考虑

| 配置 | 性能 | 安全性 | 复杂度 |
|------|------|--------|--------|
| `sign_include_body: false` | 高 | 中 | 低 |
| `sign_include_body: true` | 中 | 高 | 中 |

### 4. 调试建议

```go
// 启用调试模式查看签名参数
client := goauth.NewClient(
    baseURL, 
    appID, 
    appSecret,
    goauth.WithDebug(true),
)

// 手动调试签名
client.DebugSign(requestBody)
```

## 安全性分析

### 不包含请求体的安全性

**已有的安全机制**：
1. ✅ HMAC-SHA256签名（防篡改appId、timestamp、nonce）
2. ✅ 时间戳验证（防重放攻击）
3. ✅ Nonce随机字符串（增加签名随机性）
4. ✅ HTTPS传输（防中间人攻击）

**潜在风险**：
- ⚠️ 请求体可能在传输中被篡改（但HTTPS已提供保护）
- ⚠️ 无法验证请求体完整性

**缓解措施**：
1. 强制使用HTTPS
2. 对敏感接口启用请求体签名
3. 实现业务层数据验证

### 包含请求体的安全性

**额外保护**：
- ✅ 验证请求体完整性
- ✅ 防止请求体篡改

**代价**：
- ⚠️ 实现复杂度增加
- ⚠️ JSON序列化必须一致
- ⚠️ 调试难度增加

## 总结

### 推荐方案

✅ **采用方案1**：添加`sign_include_body`配置选项

**默认配置**：
```yaml
sign_include_body: false  # 默认不包含请求体
```

**理由**：
1. 向后兼容
2. 更灵活、更简单
3. 适用于大多数场景
4. 可针对特定应用启用请求体签名

### 实施优先级

1. **P0 - 立即实施**
   - 添加`sign_include_body`配置项
   - 默认设为`false`
   - 更新文档

2. **P1 - 短期优化**
   - 添加应用级别配置覆盖
   - 更新客户端SDK
   - 添加调试工具

3. **P2 - 长期优化**
   - 考虑请求体摘要方案
   - 性能监控和优化
   - 更多安全选项

---

**更新日期**: 2025-11-20
**版本**: 1.0.0
