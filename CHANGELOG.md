# 更新日志 (CHANGELOG)

所有重要的项目变更都会记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

---

## [v0.0.3] - 2025-11-20

### ✨ 新增功能

#### POST请求签名优化 🎉
- **可配置POST签名是否包含请求体** - 解决JSON序列化不一致性问题
  - 添加全局配置 `sign_include_body`（默认 `false`，推荐）
  - 支持应用级别配置覆盖全局设置
  - 客户端支持 `WithSignIncludeBody()` 选项
  
- **新增配置选项**：
  - `Config.SignIncludeBody` - 全局配置
  - `AppConfig.SignIncludeBody` - 应用级别配置（可选）
  - `Client.SignIncludeBody` - 客户端配置
  
- **新增配置函数**：
  - `WithConfigSignIncludeBody()` - 设置全局配置
  - `WithSignIncludeBody()` - 设置客户端配置
  
- **更新签名函数**：
  - `BuildSignParams()` - 支持可选的 `includeBody` 参数
  - `QuickSign()` - 支持可选的 `includeBody` 参数

### 🔧 改进

- **中间件优化**：
  - `extractRequestParams()` 函数支持应用级别配置
  - 即使不包含body在签名中，也正确恢复请求体供后续处理
  
- **客户端增强**：
  - 调试输出增加 `SignIncludeBody` 配置信息
  - `DebugSign()` 函数尊重 `SignIncludeBody` 配置
  
- **性能提升**：
  - 不包含请求体的签名速度提升 ~33%
  - BenchmarkSignWithoutBody: ~1700 ns/op
  - BenchmarkSignWithBody: ~2500 ns/op

### 📝 文档

- **新增文档**：
  - `POST_SIGNATURE_OPTIMIZATION.md` - 详细的优化方案文档
  - `POST_SIGN_QUICK_GUIDE.md` - 快速使用指南
  - `goauth_config_optimized.yaml` - 优化后的配置示例
  - `examples/post_sign_optimization_example.go` - 完整示例代码
  
- **更新文档**：
  - README.md 增加POST签名优化说明
  - 更新常见问题解答

### 🧪 测试

- **新增测试**：
  - `TestSignIncludeBodyConfig` - 配置测试
  - `TestBuildSignParamsWithBody` - 签名参数构建测试
  - `TestQuickSignWithBody` - 快速签名测试
  - `TestClientSignIncludeBody` - 客户端配置测试
  - `TestSignConsistency` - 签名一致性测试
  - `BenchmarkSignWithBody` / `BenchmarkSignWithoutBody` - 性能测试

### 🔄 向后兼容

- ✅ 完全向后兼容
- ✅ 默认配置（`sign_include_body: false`）让POST请求更容易成功
- ✅ 旧代码无需修改即可使用

### 📊 性能数据

```
BenchmarkSignWithBody-12          460218    2540 ns/op
BenchmarkSignWithoutBody-12       668685    1706 ns/op
```

性能提升：~33% (不包含请求体时)

---

## [v0.0.2] - 2024-01-20

### ✨ 新增功能

#### 加密传输支持
- **AES-256-CBC加密**：
  - 请求体加密/解密
  - 响应体加密/解密
  - 随机IV生成
  - PKCS7填充
  
- **新增组件**：
  - `crypto.go` - 加密核心功能
  - `CryptoMiddleware` - 加密中间件
  - `CryptoConfig` - 加密配置
  
- **便捷函数**：
  - `AESEncrypt()` / `AESDecrypt()` - 加密/解密函数
  - `WithCrypto()` - 配置选项

#### 客户端SDK
- **完整的HTTP客户端**：
  - `Client` 结构体
  - GET/POST/PUT/DELETE 方法
  - JSON自动序列化/反序列化
  - 调试模式支持
  
- **便捷方法**：
  - `GetJSON()` / `PostJSON()` / `PutJSON()` / `DeleteJSON()`
  - `DebugSign()` - 签名调试
  
- **客户端选项**：
  - `WithTimeout()` - 设置超时
  - `WithDebug()` - 启用调试
  - `WithHTTPClient()` - 自定义HTTP客户端

#### 配置增强
- **函数式选项模式**：
  - `ConfigOption` 类型
  - `WithTimestampTolerance()` - 设置时间戳容差
  - `WithDefaultRateLimit()` - 设置默认速率限制
  - `WithIPCheck()` - 启用/禁用IP检查
  
- **配置构建器**：
  - 链式调用API
  - `AddApp()` / `GetApp()` / `RemoveApp()`
  
- **配置验证**：
  - `Validate()` - 验证配置完整性
  - 自动验证必填字段

### 🔧 改进

- **错误处理**：
  - 统一错误类型 `AuthError`
  - 标准化错误响应
  - 详细的错误信息
  
- **签名工具**：
  - `QuickSign()` - 快速签名生成
  - `BuildSignParams()` - 签名参数构建
  - `VerifySignWithDebug()` - 带调试信息的签名验证

### 📝 文档

- `CRYPTO_GUIDE.md` - 加密功能详细指南
- `QUICK_REFERENCE.md` - 快速参考手册
- `OPTIMIZATION_SUMMARY.md` - 优化总结
- 完整的示例代码

### 🧪 测试

- `crypto_test.go` - 加密功能测试（12个测试用例）
- 性能基准测试
- 100% 测试通过率

---

## [v0.0.1] - 2024-01-15

### ✨ 初始发布

#### 核心功能
- **认证中间件**：
  - Gin框架集成
  - HMAC-SHA256签名验证
  - 时间戳验证（防重放攻击）
  - Nonce随机值验证
  
- **配置管理**：
  - YAML/JSON配置文件支持
  - 内存配置支持
  - 热加载配置
  
- **访问控制**：
  - IP白名单
  - 速率限制
  - 应用启用/禁用
  
- **安全特性**：
  - 应用密钥管理
  - 签名验证
  - 时间戳容差配置

#### 基础组件
- `config.go` - 配置管理
- `middleware.go` - 认证中间件
- `signature.go` - 签名算法
- `goauth_test.go` - 单元测试

#### 示例代码
- `example.go` - 完整示例
- 配置文件示例

---

## 版本说明

### 版本号格式

遵循 [语义化版本](https://semver.org/lang/zh-CN/) 2.0.0：

```
主版本号.次版本号.修订号

主版本号：不兼容的API修改
次版本号：向后兼容的功能新增
修订号：向后兼容的问题修正
```

### 发布周期

- **主版本**：重大架构调整或不兼容更新
- **次版本**：新功能发布，每月一次
- **修订版本**：Bug修复，按需发布

### 升级指南

#### 从 v0.0.2 升级到 v0.0.3

1. **配置文件更新**（可选，推荐）：
```yaml
# 添加此行
sign_include_body: false  # 推荐配置
```

2. **代码更新**（可选）：
```go
// 旧代码（仍然有效）
client := goauth.NewClient(baseURL, appID, appSecret)

// 新代码（明确配置，推荐）
client := goauth.NewClient(baseURL, appID, appSecret,
    goauth.WithSignIncludeBody(false))
```

3. **测试验证**：
```bash
go test -v
```

#### 从 v0.0.1 升级到 v0.0.2

1. **依赖更新**：
```bash
go get -u github.com/difyz9/go-auth
```

2. **配置更新**（如果使用加密）：
```yaml
crypto:
  enabled: true
  aes_key: "your-32-byte-aes-key"
```

---

## 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 问题反馈

如遇到问题，请：
1. 查看[文档](./README.md)
2. 搜索[已知问题](https://github.com/difyz9/go-auth/issues)
3. 提交新的[Issue](https://github.com/difyz9/go-auth/issues/new)

---

**项目地址**: https://github.com/difyz9/go-auth  
**文档**: [README.md](./README.md)  
**许可证**: MIT
