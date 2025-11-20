# GoAuth 快速开始指南

## 5分钟快速集成

### 步骤1：安装依赖

```bash
# 进入你的项目目录
cd your-project

# 如果是同一个仓库，直接引用
# 如果是独立项目，需要先发布包或使用 replace 指令

# 添加到 go.mod
go mod edit -require github.com/gin-gonic/gin@latest
go mod edit -require gopkg.in/yaml.v3@latest
```

### 步骤2：复制 goauth 包

```bash
# 如果在同一项目中
cp -r pkg/goauth your-project/pkg/

# 或者直接引用（推荐）
# import "your-module/pkg/goauth"
```

### 步骤3：创建配置文件

在项目根目录创建 `goauth_config.yaml`:

```yaml
timestamp_tolerance: 300
default_rate_limit: 1000
enable_ip_check: true

apps:
  my-app:
    app_id: my-app-001
    app_secret: my-secret-key-123456
    app_name: 我的应用
    require_sign: true
    enabled: true
    rate_limit: 100
    ip_whitelist:
      - "*"
```

### 步骤4：集成到代码

创建 `main.go`:

```go
package main

import (
    "github.com/gin-gonic/gin"
    "your-module/pkg/goauth"
)

func main() {
    // 1. 加载配置
    config := goauth.NewConfig()
    config.LoadFromYAML("goauth_config.yaml")
    
    // 2. 创建中间件
    auth := goauth.NewAuthMiddleware(goauth.Options{
        Config: config,
    })
    
    // 3. 创建路由
    r := gin.Default()
    
    // 4. 应用认证中间件
    api := r.Group("/api")
    api.Use(auth.Authenticate())
    {
        api.GET("/hello", func(c *gin.Context) {
            c.JSON(200, gin.H{"message": "Hello, authenticated!"})
        })
    }
    
    r.Run(":8080")
}
```

### 步骤5：测试

启动服务器：
```bash
go run main.go
```

测试请求：
```bash
# 使用提供的测试脚本
curl -X GET "http://localhost:8080/api/hello" \
  -H "X-App-Id: my-app-001" \
  -H "X-Timestamp: $(date +%s)" \
  -H "X-Nonce: $(openssl rand -hex 8)" \
  -H "X-Sign: $(./generate_sign.sh my-app-001 my-secret-key-123456)"
```

## 常见使用场景

### 场景1：不需要签名验证（快速测试）

```yaml
apps:
  test-app:
    app_id: test-app
    app_secret: test-secret
    require_sign: false  # 关闭签名验证
    enabled: true
```

### 场景2：内网应用（无IP限制）

```yaml
apps:
  internal-app:
    app_id: internal-app
    app_secret: internal-secret
    require_sign: true
    enabled: true
    ip_whitelist:
      - "*"  # 允许所有IP
```

### 场景3：生产环境（严格控制）

```yaml
apps:
  prod-app:
    app_id: prod-app
    app_secret: strong-secret-key-here
    require_sign: true
    enabled: true
    rate_limit: 1000
    ip_whitelist:
      - "192.168.1.100"
      - "192.168.1.101"
```

### 场景4：动态配置（代码管理）

```go
func main() {
    config := goauth.NewConfig()
    
    // 从数据库或其他来源加载应用配置
    apps := loadAppsFromDatabase()
    for _, app := range apps {
        config.AddApp(&goauth.AppConfig{
            AppID:       app.ID,
            AppSecret:   app.Secret,
            RequireSign: app.RequireSign,
            Enabled:     app.Enabled,
            RateLimit:   app.RateLimit,
            IPWhitelist: app.IPWhitelist,
        })
    }
    
    auth := goauth.NewAuthMiddleware(goauth.Options{
        Config: config,
    })
    
    // ... 其他代码
}
```

## 客户端调用示例

### Go 客户端

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "strconv"
    "time"
    "your-module/pkg/goauth"
)

func main() {
    appID := "my-app-001"
    appSecret := "my-secret-key-123456"
    
    // 准备参数
    timestamp := strconv.FormatInt(time.Now().Unix(), 10)
    nonce := goauth.GenerateNonce(16)
    
    // 生成签名
    params := map[string]string{
        "appId":     appID,
        "timestamp": timestamp,
        "nonce":     nonce,
    }
    sign := goauth.GenerateSign(params, appSecret)
    
    // 发送请求
    req, _ := http.NewRequest("GET", "http://localhost:8080/api/hello", nil)
    req.Header.Set("X-App-Id", appID)
    req.Header.Set("X-Timestamp", timestamp)
    req.Header.Set("X-Nonce", nonce)
    req.Header.Set("X-Sign", sign)
    
    resp, _ := http.DefaultClient.Do(req)
    defer resp.Body.Close()
    
    // 处理响应...
}
```

### JavaScript 客户端

```javascript
const crypto = require('crypto');

function generateSign(params, appSecret) {
    // 排序参数
    const sortedKeys = Object.keys(params).sort();
    
    // 拼接字符串
    const signString = sortedKeys
        .map(key => `${key}=${params[key]}`)
        .join('&');
    
    // HMAC-SHA256
    const hmac = crypto.createHmac('sha256', appSecret);
    hmac.update(signString);
    return hmac.digest('hex');
}

// 使用示例
const appId = 'my-app-001';
const appSecret = 'my-secret-key-123456';
const timestamp = Math.floor(Date.now() / 1000).toString();
const nonce = Math.random().toString(36).substring(7);

const params = {
    appId: appId,
    timestamp: timestamp,
    nonce: nonce
};

const sign = generateSign(params, appSecret);

// 发送请求
fetch('http://localhost:8080/api/hello', {
    headers: {
        'X-App-Id': appId,
        'X-Timestamp': timestamp,
        'X-Nonce': nonce,
        'X-Sign': sign
    }
});
```

### Python 客户端

```python
import hmac
import hashlib
import time
import random
import string
import requests

def generate_sign(params, app_secret):
    # 排序参数
    sorted_params = sorted(params.items())
    
    # 拼接字符串
    sign_string = '&'.join([f"{k}={v}" for k, v in sorted_params])
    
    # HMAC-SHA256
    sign = hmac.new(
        app_secret.encode('utf-8'),
        sign_string.encode('utf-8'),
        hashlib.sha256
    ).hexdigest()
    
    return sign

# 使用示例
app_id = 'my-app-001'
app_secret = 'my-secret-key-123456'
timestamp = str(int(time.time()))
nonce = ''.join(random.choices(string.ascii_letters + string.digits, k=16))

params = {
    'appId': app_id,
    'timestamp': timestamp,
    'nonce': nonce
}

sign = generate_sign(params, app_secret)

# 发送请求
response = requests.get(
    'http://localhost:8080/api/hello',
    headers={
        'X-App-Id': app_id,
        'X-Timestamp': timestamp,
        'X-Nonce': nonce,
        'X-Sign': sign
    }
)
```

## 故障排查

### 问题1：签名验证失败

**检查清单：**
- [ ] appId 和 appSecret 是否正确
- [ ] 时间戳是否在有效期内
- [ ] 参数是否完整（appId, timestamp, nonce）
- [ ] 参数排序是否正确
- [ ] 拼接字符串格式是否为 `key1=value1&key2=value2`
- [ ] 是否使用了 HMAC-SHA256 算法
- [ ] 签名是否为小写十六进制

**调试方法：**
```go
// 服务端打印
params := map[string]string{
    "appId": "my-app",
    "timestamp": "1700000000",
    "nonce": "abc123",
}
expected := goauth.GenerateSign(params, "my-secret")
fmt.Println("Expected sign:", expected)
```

### 问题2：时间戳无效

**原因：**
- 客户端和服务端时间不同步
- 请求发送太慢

**解决方法：**
- 同步服务器时间（使用 NTP）
- 增加时间容差：`timestamp_tolerance: 600`（10分钟）

### 问题3：IP 被拒绝

**检查：**
- 确认客户端真实 IP
- 检查 IP 白名单配置
- 如果在代理后面，确认获取到正确的 IP

### 问题4：速率限制

**解决：**
- 增加应用的 `rate_limit` 值
- 减少请求频率
- 使用多个应用分散请求

## 下一步

- 查看 [README.md](README.md) 了解完整功能
- 查看 [example.go](example.go) 了解更多示例
- 根据需求自定义配置文件
- 集成到生产环境时记得更改密钥

## 支持

如有问题，请查看项目文档或提交 Issue。
