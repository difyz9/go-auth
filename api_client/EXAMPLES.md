# 多语言客户端加密使用示例

## Dart 示例

```dart
import 'package:api_client/api_client.dart';

void main() async {
  // ========== 方式1：使用默认配置（不加密） ==========
  print('📝 示例1: 基础API调用（不加密）');
  
  final response1 = await apiClient.get('/api/test');
  print('响应: $response1');
  
  // ========== 方式2：创建启用加密的客户端 ==========
  print('\n📝 示例2: 创建加密客户端');
  
  final encryptedClient = createEncryptedApiClient(
    appId: 'test-app-001',
    appSecret: 'tmcf5m6qcm6k9hrp3sy8rhgafu00ttph',
    aesKey: 'my-super-secret-key-32-bytes!',
    baseURL: 'http://localhost:8089',
  );
  
  // 自动加密请求
  final response2 = await encryptedClient.post('/api/crypto/echo', 
    data: {
      'message': 'Hello, encrypted world!',
      'timestamp': DateTime.now().millisecondsSinceEpoch,
    },
  );
  print('🔐 加密请求响应: $response2');
  
  // ========== 方式3：请求加密响应 ==========
  print('\n📝 示例3: 请求服务器返回加密响应');
  
  final response3 = await encryptedClient.get('/api/sensitive/data',
    requestEncrypted: true,
  );
  print('🔓 解密后的响应: $response3');
  
  // ========== 方式4：单次请求加密 ==========
  print('\n📝 示例4: 单次请求加密（覆盖默认配置）');
  
  final normalClient = ApiClient(ApiConfig(
    appId: 'test-app-001',
    appSecret: 'tmcf5m6qcm6k9hrp3sy8rhgafu00ttph',
    aesKey: 'my-super-secret-key-32-bytes!',
    enableEncryption: false,  // 默认不加密
  ));
  
  // 只加密这一次请求
  final response4 = await normalClient.post('/api/payment',
    data: {
      'amount': 99.99,
      'card': '1234-5678-9012-3456',
    },
    encrypt: true,  // 单次启用加密
  );
  print('💳 支付响应: $response4');
}
```

---

## JavaScript 示例

```javascript
import { apiClient, createEncryptedApiClient } from './apiSign_crypto.js';

async function examples() {
  // ========== 示例1: 基础API调用（不加密） ==========
  console.log('📝 示例1: 基础API调用（不加密）');
  
  const response1 = await apiClient.get('/api/test');
  console.log('响应:', response1);
  
  // ========== 示例2: 创建加密客户端 ==========
  console.log('\n📝 示例2: 创建加密客户端');
  
  const encryptedClient = createEncryptedApiClient({
    appId: 'test-app-001',
    appSecret: 'tmcf5m6qcm6k9hrp3sy8rhgafu00ttph',
    aesKey: 'my-super-secret-key-32-bytes!',
    baseURL: 'http://localhost:8089'
  });
  
  // 自动加密请求
  const response2 = await encryptedClient.post('/api/crypto/echo', {
    message: 'Hello, encrypted world!',
    timestamp: Date.now()
  });
  console.log('🔐 加密请求响应:', response2);
  
  // ========== 示例3: 请求加密响应 ==========
  console.log('\n📝 示例3: 请求服务器返回加密响应');
  
  const response3 = await encryptedClient.get('/api/sensitive/data', null, {
    requestEncrypted: true
  });
  console.log('🔓 解密后的响应:', response3);
  
  // ========== 示例4: 单次请求加密 ==========
  console.log('\n📝 示例4: 单次请求加密');
  
  const normalClient = new ApiClient({
    appId: 'test-app-001',
    appSecret: 'tmcf5m6qcm6k9hrp3sy8rhgafu00ttph',
    aesKey: 'my-super-secret-key-32-bytes!',
    enableEncryption: false  // 默认不加密
  });
  
  // 只加密这一次请求
  const response4 = await normalClient.post('/api/payment', {
    amount: 99.99,
    card: '1234-5678-9012-3456'
  }, {
    encrypt: true  // 单次启用加密
  });
  console.log('💳 支付响应:', response4);
  
  // ========== 示例5: 手动加密/解密 ==========
  console.log('\n📝 示例5: 手动加密/解密');
  
  const plaintext = 'Sensitive data';
  const encrypted = await encryptedClient.aesEncrypt(plaintext);
  console.log('加密:', encrypted);
  
  const decrypted = await encryptedClient.aesDecrypt(encrypted);
  console.log('解密:', decrypted);
}

// 运行示例
examples().catch(console.error);
```

---

## TypeScript 示例

```typescript
import { apiClient, createEncryptedApiClient, ApiClient } from './api-client-crypto';

// 定义数据类型
interface TestData {
  message: string;
  timestamp: number;
}

interface PaymentData {
  amount: number;
  card: string;
}

interface UserData {
  id: number;
  name: string;
  email: string;
}

async function examples() {
  // ========== 示例1: 基础API调用（不加密） ==========
  console.log('📝 示例1: 基础API调用（不加密）');
  
  const response1 = await apiClient.get<TestData>('/api/test');
  console.log('响应:', response1);
  
  // ========== 示例2: 创建加密客户端（类型安全） ==========
  console.log('\n📝 示例2: 创建加密客户端');
  
  const encryptedClient = createEncryptedApiClient({
    appId: 'test-app-001',
    appSecret: 'tmcf5m6qcm6k9hrp3sy8rhgafu00ttph',
    aesKey: 'my-super-secret-key-32-bytes!',
    baseURL: 'http://localhost:8089'
  });
  
  // 自动加密请求（类型安全）
  const response2 = await encryptedClient.post<TestData>('/api/crypto/echo', {
    message: 'Hello, encrypted world!',
    timestamp: Date.now()
  });
  console.log('🔐 加密请求响应:', response2);
  
  // ========== 示例3: 请求加密响应 ==========
  console.log('\n📝 示例3: 请求服务器返回加密响应');
  
  const response3 = await encryptedClient.get<UserData>('/api/sensitive/data', undefined, {
    requestEncrypted: true
  });
  console.log('🔓 解密后的响应:', response3);
  
  // ========== 示例4: 单次请求加密 ==========
  console.log('\n📝 示例4: 单次请求加密');
  
  const normalClient = new ApiClient({
    appId: 'test-app-001',
    appSecret: 'tmcf5m6qcm6k9hrp3sy8rhgafu00ttph',
    aesKey: 'my-super-secret-key-32-bytes!',
    enableEncryption: false  // 默认不加密
  });
  
  // 只加密这一次请求
  const response4 = await normalClient.post<PaymentData>('/api/payment', {
    amount: 99.99,
    card: '1234-5678-9012-3456'
  }, {
    encrypt: true  // 单次启用加密
  });
  console.log('💳 支付响应:', response4);
  
  // ========== 示例5: 完整的CRUD操作 ==========
  console.log('\n📝 示例5: 完整的CRUD操作');
  
  // 创建用户（加密）
  const newUser = await encryptedClient.post<UserData>('/api/users', {
    name: 'John Doe',
    email: 'john@example.com'
  });
  console.log('✅ 用户已创建:', newUser);
  
  // 获取用户（请求加密响应）
  const user = await encryptedClient.get<UserData>(`/api/users/${newUser.id}`, undefined, {
    requestEncrypted: true
  });
  console.log('👤 用户信息:', user);
  
  // 更新用户（加密）
  const updated = await encryptedClient.put<UserData>(`/api/users/${newUser.id}`, {
    name: 'John Smith',
    email: 'john.smith@example.com'
  });
  console.log('🔄 用户已更新:', updated);
  
  // 删除用户
  await encryptedClient.delete(`/api/users/${newUser.id}`);
  console.log('🗑️ 用户已删除');
}

// 运行示例
examples().catch(console.error);
```

---

## 配置对比

### 普通客户端（无加密）

```dart
// Dart
final client = ApiClient(ApiConfig(
  appId: 'test-app-001',
  appSecret: 'your-secret',
));
```

```javascript
// JavaScript
const client = new ApiClient({
  appId: 'test-app-001',
  appSecret: 'your-secret'
});
```

```typescript
// TypeScript
const client = new ApiClient({
  appId: 'test-app-001',
  appSecret: 'your-secret'
});
```

### 加密客户端

```dart
// Dart
final client = createEncryptedApiClient(
  appId: 'test-app-001',
  appSecret: 'your-secret',
  aesKey: 'your-32-byte-aes-key',
);
```

```javascript
// JavaScript
const client = createEncryptedApiClient({
  appId: 'test-app-001',
  appSecret: 'your-secret',
  aesKey: 'your-32-byte-aes-key'
});
```

```typescript
// TypeScript
const client = createEncryptedApiClient({
  appId: 'test-app-001',
  appSecret: 'your-secret',
  aesKey: 'your-32-byte-aes-key'
});
```

---

## 环境变量配置

### Dart

```dart
// 从环境变量读取
import 'dart:io';

final aesKey = Platform.environment['AES_KEY'] ?? '';
final client = createEncryptedApiClient(
  appId: Platform.environment['APP_ID']!,
  appSecret: Platform.environment['APP_SECRET']!,
  aesKey: aesKey,
);
```

### JavaScript/TypeScript

```javascript
// 从环境变量读取（Node.js）
const client = createEncryptedApiClient({
  appId: process.env.APP_ID,
  appSecret: process.env.APP_SECRET,
  aesKey: process.env.AES_KEY
});
```

### .env 文件示例

```env
APP_ID=test-app-001
APP_SECRET=tmcf5m6qcm6k9hrp3sy8rhgafu00ttph
AES_KEY=my-super-secret-key-32-bytes!
API_BASE_URL=http://localhost:8089
```

---

## 调试技巧

### 1. 查看加密日志

所有客户端都会输出加密/解密日志：

```
🔐 请求已加密
🔓 收到加密响应，正在解密...
✅ 响应解密成功
```

### 2. 测试加密是否工作

```dart
// Dart
final encrypted = client._aesEncrypt('test');
print('加密结果: $encrypted');

final decrypted = client._aesDecrypt(encrypted);
print('解密结果: $decrypted');
```

```javascript
// JavaScript
const encrypted = await client.aesEncrypt('test');
console.log('加密结果:', encrypted);

const decrypted = await client.aesDecrypt(encrypted);
console.log('解密结果:', decrypted);
```

### 3. 验证与后端兼容性

```dart
// 发送测试请求
final response = await encryptedClient.post('/api/crypto/test', 
  data: {'test': 'data'},
);

// 检查响应是否正确解密
assert(response != null);
```

---

## 性能测试

### JavaScript性能测试

```javascript
// 测试加密性能
const iterations = 1000;
const start = Date.now();

for (let i = 0; i < iterations; i++) {
  await client.aesEncrypt('test data');
}

const elapsed = Date.now() - start;
console.log(`${iterations}次加密耗时: ${elapsed}ms`);
console.log(`平均每次: ${elapsed/iterations}ms`);
```

---

## 常见错误处理

```typescript
// TypeScript错误处理示例
try {
  const response = await encryptedClient.post('/api/test', data);
  console.log('✅ 成功:', response);
} catch (error) {
  if (error.message.includes('解密失败')) {
    console.error('❌ 密钥可能不正确');
  } else if (error.message.includes('HTTP')) {
    console.error('❌ 网络错误:', error.message);
  } else {
    console.error('❌ 未知错误:', error);
  }
}
```

---

**提示**: 更多详细信息请参考 `CRYPTO_INTEGRATION.md`
