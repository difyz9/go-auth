# API客户端加解密集成指南

## 概述

本项目已为多语言客户端集成AES-256-CBC加解密功能，与Go后端完全兼容。所有客户端均支持：
- ✅ HMAC-SHA256签名认证
- ✅ AES-256-CBC加密/解密
- ✅ 自动请求/响应加解密
- ✅ 向后兼容（可选启用）

## 支持的语言

### 1. Dart/Flutter 客户端

**文件位置**：`api_client/dart/api_client.dart`

**依赖**：
```yaml
dependencies:
  dio: ^5.0.0
  crypto: ^3.0.3
  shared_preferences: ^2.0.15
  encrypt: ^5.0.1  # 新增：用于AES加密
```

**基础使用**（不加密）：
```dart
// 使用默认配置
final response = await apiClient.get('/api/test');
```

**启用加密**：
```dart
// 方式1：创建带加密的客户端
final encryptedClient = createEncryptedApiClient(
  appId: 'test-app-001',
  appSecret: 'your-secret',
  aesKey: 'your-32-byte-aes-key',
  baseURL: 'http://localhost:8089',
);

// 方式2：使用自定义配置
final client = ApiClient(ApiConfig(
  appId: 'test-app-001',
  appSecret: 'your-secret',
  aesKey: 'your-32-byte-aes-key',
  enableEncryption: true,
));

// 发送加密请求
final response = await client.post('/api/protected', 
  data: {'key': 'value'},
);

// 请求加密响应
final response = await client.get('/api/data',
  requestEncrypted: true,
);

// 单次请求加密（覆盖默认配置）
final response = await client.post('/api/sensitive',
  data: {'sensitive': 'data'},
  encrypt: true,
);
```

**主要特性**：
- 自动加密请求体（当`enableEncryption=true`）
- 自动解密加密响应
- 支持请求级别的加密控制
- 日志输出：🔐（加密）、🔓（解密）、✅（成功）

---

### 2. JavaScript 客户端

**文件位置**：`api_client/javascript/apiSign_crypto.js`

**依赖**：
- 浏览器环境：无需额外依赖（使用Web Crypto API）
- Node.js环境：需要crypto模块

**基础使用**（不加密）：
```javascript
import { apiClient } from './apiSign_crypto.js';

// 使用默认配置
const data = await apiClient.get('/api/test');
```

**启用加密**：
```javascript
import { createEncryptedApiClient } from './apiSign_crypto.js';

// 创建启用加密的客户端
const client = createEncryptedApiClient({
  appId: 'test-app-001',
  appSecret: 'your-secret',
  aesKey: 'your-32-byte-aes-key',
  baseURL: 'http://localhost:8089'
});

// 发送加密请求（自动加密）
const response = await client.post('/api/protected', {
  key: 'value'
});

// 请求加密响应
const response = await client.get('/api/data', null, {
  requestEncrypted: true
});

// 单次请求加密
const response = await client.post('/api/sensitive', 
  { sensitive: 'data' },
  { encrypt: true }
);
```

**手动加密/解密**：
```javascript
// 手动加密
const encrypted = await client.aesEncrypt('plaintext');

// 手动解密
const decrypted = await client.aesDecrypt(encrypted);
```

**特点**：
- 异步加密（Web Crypto API）
- Base64编码
- 随机IV（每次加密）
- 控制台日志：🔐、🔓、✅

---

### 3. TypeScript 客户端

**文件位置**：`api_client/typescript/api-client-crypto.ts`

**依赖**：
```json
{
  "dependencies": {
    "node-fetch": "^3.0.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0"
  }
}
```

**基础使用**（不加密）：
```typescript
import { apiClient, ApiClient } from './api-client-crypto';

// 使用默认配置
const data = await apiClient.get<MyDataType>('/api/test');
```

**启用加密**：
```typescript
import { createEncryptedApiClient, ApiClient } from './api-client-crypto';

// 创建启用加密的客户端
const client = createEncryptedApiClient({
  appId: 'test-app-001',
  appSecret: 'your-secret',
  aesKey: 'your-32-byte-aes-key',
  baseURL: 'http://localhost:8089'
});

// 发送加密请求（类型安全）
interface UserData {
  name: string;
  email: string;
}

const response = await client.post<UserData>('/api/user', {
  name: 'John',
  email: 'john@example.com'
});

// 请求加密响应
const data = await client.get<UserData>('/api/user/1', undefined, {
  requestEncrypted: true
});

// 单次请求加密
const result = await client.post<any>('/api/sensitive', 
  { sensitive: 'data' },
  { encrypt: true }
);
```

**类型定义**：
```typescript
interface ApiConfig {
  appId: string;
  appSecret: string;
  baseURL?: string;
  timeout?: number;
  aesKey?: string;
  enableEncryption?: boolean;
}

interface RequestOptions {
  method?: string;
  data?: any;
  headers?: Record<string, string>;
  encrypt?: boolean;
  requestEncrypted?: boolean;
}
```

**特点**：
- 完整的TypeScript类型支持
- 泛型支持（类型安全）
- Node.js crypto模块
- 同步加密/解密

---

## 加密机制

### AES-256-CBC加密

所有客户端使用相同的加密算法：

1. **密钥处理**：
   - 自动标准化为16/24/32字节
   - 支持任意长度密钥（自动填充或截断）

2. **加密流程**：
   ```
   明文 → UTF-8编码 → AES-CBC加密 → IV+密文 → Base64编码
   ```

3. **解密流程**：
   ```
   Base64解码 → 分离IV和密文 → AES-CBC解密 → UTF-8解码 → 明文
   ```

4. **IV生成**：
   - 每次加密使用随机IV（16字节）
   - IV附加在密文前面

### 请求/响应格式

**加密请求**：
```json
{
  "data": "base64-encoded-encrypted-data",
  "timestamp": 1234567890
}
```

**加密响应**：
```json
{
  "encrypted": true,
  "data": "base64-encoded-encrypted-data",
  "timestamp": 1234567890
}
```

**解密后的响应**：
```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

---

## 配置选项

### 全局配置

所有客户端支持以下配置：

| 选项 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `appId` | string | ✅ | 应用ID |
| `appSecret` | string | ✅ | 应用密钥（用于签名） |
| `baseURL` | string | ❌ | API基础URL |
| `timeout` | number | ❌ | 请求超时（毫秒） |
| `aesKey` | string | ❌ | AES密钥（用于加密） |
| `enableEncryption` | boolean | ❌ | 是否默认加密 |

### 请求级别选项

| 选项 | 类型 | 说明 |
|------|------|------|
| `encrypt` | boolean | 是否加密此请求体 |
| `requestEncrypted` | boolean | 是否请求加密响应 |

---

## 最佳实践

### 1. 密钥管理

```dart
// Dart: 从环境变量或配置文件读取
final aesKey = Platform.environment['AES_KEY'] ?? '';

// JavaScript/TypeScript: 使用环境变量
const aesKey = process.env.AES_KEY || '';
```

### 2. 错误处理

```dart
// Dart
try {
  final response = await client.post('/api/test', data: data);
} catch (e) {
  print('请求失败: $e');
}
```

```javascript
// JavaScript/TypeScript
try {
  const response = await client.post('/api/test', data);
} catch (error) {
  console.error('请求失败:', error);
}
```

### 3. 选择性加密

```dart
// 只加密敏感接口
final client = ApiClient(ApiConfig(
  appId: 'test-app-001',
  appSecret: 'your-secret',
  aesKey: 'your-aes-key',
  enableEncryption: false,  // 默认不加密
));

// 敏感操作时手动启用加密
await client.post('/api/payment', 
  data: paymentData,
  encrypt: true,  // 单次启用加密
);
```

### 4. 调试模式

所有客户端在加密/解密时会输出日志：
- 🔐 请求已加密
- 🔓 收到加密响应，正在解密...
- ✅ 响应解密成功
- ❌ 加密/解密失败

---

## 与Go后端对接

### 后端配置

```go
// 启用加密中间件
cryptoCfg := &goauth.CryptoConfig{
    Enabled: true,
    AESKey:  "your-32-byte-aes-key",  // 与客户端相同
}

middleware := goauth.New(
    goauth.WithConfig(cfg),
    goauth.WithCrypto(cryptoCfg),
)

r.Use(middleware.AuthMiddleware())
```

### 请求头

客户端自动添加以下请求头：
- `X-Encrypted: 1` - 表示请求已加密
- `X-Response-Encrypt: 1` - 请求返回加密响应

---

## 迁移指南

### 从未加密客户端迁移

1. **添加依赖**（Dart需要）：
   ```yaml
   encrypt: ^5.0.1
   ```

2. **更新配置**：
   ```dart
   // 旧的配置
   final client = ApiClient(ApiConfig(
     appId: 'test-app-001',
     appSecret: 'your-secret',
   ));

   // 新的配置（向后兼容）
   final client = ApiClient(ApiConfig(
     appId: 'test-app-001',
     appSecret: 'your-secret',
     aesKey: 'your-aes-key',  // 新增
     enableEncryption: false,  // 默认不加密，保持兼容
   ));
   ```

3. **逐步启用加密**：
   - 先在测试环境测试
   - 对敏感接口启用加密
   - 逐步扩展到所有接口

---

## 性能考虑

### 加密性能

- **Go后端**：~35-40μs/次
- **Dart客户端**：~1-2ms/次
- **JavaScript**：~2-3ms/次（异步）
- **TypeScript**：~1-2ms/次

### 优化建议

1. **选择性加密**：只加密敏感数据
2. **缓存密钥**：密钥对象可以复用
3. **批量操作**：减少加密次数

---

## 常见问题

### 1. 加密失败
- 检查AES密钥是否正确
- 确保密钥长度合适（推荐32字节）
- 查看控制台错误日志

### 2. 解密失败
- 确保前后端使用相同的密钥
- 检查Base64编码是否正确
- 验证IV是否正确分离

### 3. 签名验证失败
- 签名和加密是独立的
- 确保`appSecret`正确
- 检查时间戳是否在有效期内

---

## 示例代码

完整示例请参考：
- Dart: `examples/crypto_example.go`
- JavaScript: `api_client/javascript/apiSign_crypto.js`
- TypeScript: `api_client/typescript/api-client-crypto.ts`

---

## 技术支持

遇到问题请查看：
- `CRYPTO_GUIDE.md` - 加密功能详细指南
- `README.md` - 项目总体文档
- `QUICK_REFERENCE.md` - 快速参考

---

**更新日期**: 2024-01-20
**版本**: 1.0.0
