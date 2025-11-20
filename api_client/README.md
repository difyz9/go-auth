# API Client - 多语言客户端

## 概述

本目录包含与Go-Auth后端配套的多语言客户端实现，支持：
- ✅ HMAC-SHA256签名认证
- ✅ AES-256-CBC加解密（可选）
- ✅ 请求/响应自动加解密
- ✅ 完全兼容Go后端

## 支持的语言

### 1. Dart/Flutter 客户端

**目录**: `dart/`

**文件**:
- `api_client.dart` - 带加密支持的完整客户端

**特性**:
- Dio HTTP客户端
- SharedPreferences token管理
- AES-256-CBC加密/解密
- 拦截器支持
- 日志输出

**依赖**:
```yaml
dependencies:
  dio: ^5.0.0
  crypto: ^3.0.3
  shared_preferences: ^2.0.15
  encrypt: ^5.0.1
```

**快速开始**:
```dart
import 'package:api_client/api_client.dart';

// 创建加密客户端
final client = createEncryptedApiClient(
  appId: 'test-app-001',
  appSecret: 'your-secret',
  aesKey: 'your-32-byte-aes-key',
);

// 发送加密请求
final response = await client.post('/api/test', data: {'key': 'value'});
```

---

### 2. JavaScript 客户端

**目录**: `javascript/`

**文件**:
- `apiSign.js` - 原始客户端（无加密）
- `apiSign_crypto.js` - 带加密支持的客户端 ⭐

**特性**:
- Web Crypto API
- 异步加密/解密
- localStorage token管理
- 浏览器环境
- 控制台日志

**使用**:
```javascript
import { createEncryptedApiClient } from './apiSign_crypto.js';

// 创建加密客户端
const client = createEncryptedApiClient({
  appId: 'test-app-001',
  appSecret: 'your-secret',
  aesKey: 'your-32-byte-aes-key'
});

// 发送加密请求
const response = await client.post('/api/test', { key: 'value' });
```

---

### 3. TypeScript 客户端

**目录**: `typescript/`

**文件**:
- `api-client.ts` - 原始客户端（无加密）
- `api-client-crypto.ts` - 带加密支持的客户端 ⭐

**特性**:
- Node.js crypto模块
- 完整TypeScript类型
- fetch API
- 泛型支持
- 类型安全

**依赖**:
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

**使用**:
```typescript
import { createEncryptedApiClient } from './api-client-crypto';

// 创建加密客户端
const client = createEncryptedApiClient({
  appId: 'test-app-001',
  appSecret: 'your-secret',
  aesKey: 'your-32-byte-aes-key'
});

// 发送加密请求（类型安全）
interface User {
  id: number;
  name: string;
}
const user = await client.post<User>('/api/users', { name: 'John' });
```

---

## 功能对比

| 功能 | Dart | JavaScript | TypeScript |
|------|------|------------|------------|
| HMAC-SHA256签名 | ✅ | ✅ | ✅ |
| AES-256-CBC加密 | ✅ | ✅ | ✅ |
| 自动加密请求 | ✅ | ✅ | ✅ |
| 自动解密响应 | ✅ | ✅ | ✅ |
| Token管理 | ✅ | ✅ | ✅ |
| 类型安全 | ✅ | ❌ | ✅ |
| 拦截器 | ✅ | ❌ | ❌ |
| HTTP客户端 | Dio | fetch | fetch |
| 环境 | Mobile/Desktop | Browser | Node.js |

---

## 配置说明

### 基础配置（所有客户端通用）

```javascript
{
  appId: 'test-app-001',           // 应用ID（必填）
  appSecret: 'your-secret',        // 应用密钥（必填）
  baseURL: 'http://localhost:8089', // API基础URL
  timeout: 10000,                  // 超时时间（毫秒）
  aesKey: 'your-aes-key',          // AES密钥（可选）
  enableEncryption: false          // 是否默认加密
}
```

### 加密配置

**启用全局加密**:
```javascript
// 所有请求默认加密
enableEncryption: true
```

**单次请求加密**:
```javascript
// 只加密特定请求
await client.post('/api/sensitive', data, { encrypt: true });
```

**请求加密响应**:
```javascript
// 请求服务器返回加密响应
await client.get('/api/data', null, { requestEncrypted: true });
```

---

## 文档

- **[CRYPTO_INTEGRATION.md](./CRYPTO_INTEGRATION.md)** - 加密集成完整指南
- **[EXAMPLES.md](./EXAMPLES.md)** - 多语言使用示例
- **[../CRYPTO_GUIDE.md](../CRYPTO_GUIDE.md)** - 加密功能详细说明
- **[../README.md](../README.md)** - Go-Auth主文档

---

## 快速开始

### 1. 选择客户端

根据你的技术栈选择对应的客户端：
- **Flutter/Dart应用** → `dart/api_client.dart`
- **Web前端** → `javascript/apiSign_crypto.js`
- **Node.js后端** → `typescript/api-client-crypto.ts`

### 2. 安装依赖

#### Dart
```yaml
# pubspec.yaml
dependencies:
  dio: ^5.0.0
  crypto: ^3.0.3
  shared_preferences: ^2.0.15
  encrypt: ^5.0.1
```

#### JavaScript/TypeScript
```bash
# JavaScript (浏览器) - 无需额外依赖

# TypeScript (Node.js)
npm install node-fetch
npm install -D @types/node
```

### 3. 创建客户端

```dart
// Dart
final client = createEncryptedApiClient(
  appId: 'test-app-001',
  appSecret: 'your-secret',
  aesKey: 'your-aes-key',
);
```

```javascript
// JavaScript
const client = createEncryptedApiClient({
  appId: 'test-app-001',
  appSecret: 'your-secret',
  aesKey: 'your-aes-key'
});
```

```typescript
// TypeScript
const client = createEncryptedApiClient({
  appId: 'test-app-001',
  appSecret: 'your-secret',
  aesKey: 'your-aes-key'
});
```

### 4. 发送请求

```dart
// Dart
final data = await client.post('/api/test', data: {'key': 'value'});
```

```javascript
// JavaScript/TypeScript
const data = await client.post('/api/test', { key: 'value' });
```

---

## 与Go后端集成

### 后端配置

```go
// config.yaml
apps:
  - app_id: test-app-001
    app_secret: your-secret
    enabled: true

crypto:
  enabled: true
  aes_key: your-32-byte-aes-key
```

```go
// main.go
cryptoCfg := &goauth.CryptoConfig{
    Enabled: true,
    AESKey:  "your-32-byte-aes-key",
}

middleware := goauth.New(
    goauth.MustLoadYAML("config.yaml"),
    goauth.WithCrypto(cryptoCfg),
)

r.Use(middleware.AuthMiddleware())
```

### 请求流程

```
客户端                           服务器
  |                                |
  | -- 1. 加密请求体 -->           |
  | -- 2. HMAC签名 -->             |
  | -- 3. 发送请求 -->             |
  |                                | -- 4. 验证签名
  |                                | -- 5. 解密请求
  |                                | -- 6. 处理业务
  |                                | -- 7. 加密响应
  | <-- 8. 返回响应 --             |
  | -- 9. 解密响应                 |
  | -- 10. 返回数据                |
```

---

## 安全建议

### 1. 密钥管理
- ✅ 使用环境变量存储密钥
- ✅ 不要在代码中硬编码密钥
- ✅ 定期轮换密钥
- ❌ 不要提交密钥到版本控制

### 2. 传输安全
- ✅ 使用HTTPS
- ✅ 验证服务器证书
- ✅ 启用证书固定（可选）

### 3. 加密策略
- ✅ 敏感数据必须加密
- ✅ 选择性加密以提高性能
- ✅ 使用足够长的密钥（32字节）

---

## 性能优化

### 1. 选择性加密
```javascript
// 只加密敏感操作
const client = new ApiClient({
  enableEncryption: false  // 默认不加密
});

// 敏感操作时启用
await client.post('/api/payment', data, { encrypt: true });
```

### 2. 缓存策略
```javascript
// 缓存非敏感数据
const cache = new Map();
const cacheKey = '/api/config';

if (cache.has(cacheKey)) {
  return cache.get(cacheKey);
}

const data = await client.get(cacheKey);
cache.set(cacheKey, data);
```

### 3. 批量操作
```javascript
// 批量请求减少加密次数
const requests = users.map(user => 
  client.post('/api/users', user, { encrypt: true })
);
const results = await Promise.all(requests);
```

---

## 故障排查

### 加密失败

**症状**: ❌ AES加密失败
**原因**: 
- AES密钥为空
- 密钥格式错误
- 加密库未正确安装

**解决**:
```dart
// Dart: 检查依赖
flutter pub get

// 检查密钥
print('AES Key: ${client.aesKey}');
```

### 解密失败

**症状**: ❌ AES解密失败
**原因**:
- 前后端密钥不一致
- Base64编码错误
- IV分离错误

**解决**:
```javascript
// 验证密钥一致性
console.log('Client Key:', client.aesKey);
console.log('Server Key:', serverKey);

// 测试加解密
const test = await client.aesEncrypt('test');
const result = await client.aesDecrypt(test);
console.log('Test:', result === 'test' ? '✅' : '❌');
```

### 签名验证失败

**症状**: 401 Unauthorized
**原因**:
- appSecret不正确
- 时间戳过期
- 签名算法不匹配

**解决**:
```javascript
// 检查配置
console.log('App ID:', client.config.appId);
console.log('App Secret:', client.config.appSecret);

// 检查时间同步
const serverTime = await client.get('/api/time');
const localTime = Date.now();
console.log('时间差:', Math.abs(serverTime - localTime), 'ms');
```

---

## 版本历史

- **v1.0.0** (2024-01-20)
  - ✅ 初始版本
  - ✅ Dart客户端加密支持
  - ✅ JavaScript客户端加密支持
  - ✅ TypeScript客户端加密支持
  - ✅ 完整文档和示例

---

## 贡献

欢迎提交问题和改进建议！

---

## 许可证

MIT License

---

**更新日期**: 2024-01-20
**维护者**: Go-Auth Team
