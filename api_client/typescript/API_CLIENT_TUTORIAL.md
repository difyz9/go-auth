# API Client 使用教程

## 概述

`api-client.ts` 是一个统一的网络请求工具类，它整合了：
- **API签名认证**：自动生成 X-App-Id, X-Timestamp, X-Nonce, X-Sign 等认证头
- **JWT token认证**：自动为受保护端点添加 Authorization Bearer token
- **智能端点识别**：自动判断公开端点和受保护端点
- **错误处理**：统一的错误处理和自动token清理
- **环境适配**：开发环境和生产环境的自动适配

## 快速开始

### 1. 基本导入

```typescript
import apiClient from '@/lib/api-client';
```

### 2. 最简单的使用

```typescript
// GET 请求
const userInfo = await apiClient.get('/api/v1/auth/verify');

// POST 请求
const loginResult = await apiClient.post('/api/v1/auth/login', {
  email: 'admin@126.com',
  password: 'Ab123456'
});
```

## 详细功能说明

### 1. 支持的HTTP方法

```typescript
// GET 请求
const data = await apiClient.get<ResponseType>('/api/endpoint');

// POST 请求
const result = await apiClient.post<ResponseType>('/api/endpoint', requestData);

// PUT 请求
const updated = await apiClient.put<ResponseType>('/api/endpoint', updateData);

// DELETE 请求
const deleted = await apiClient.delete<ResponseType>('/api/endpoint');

// PATCH 请求
const patched = await apiClient.patch<ResponseType>('/api/endpoint', patchData);
```

### 2. 带参数的GET请求

```typescript
// 查询参数会自动构建到URL中
const orders = await apiClient.get('/api/v1/orders', {
  page: 1,
  pageSize: 10,
  status: 'completed'
});
// 实际请求: GET /api/v1/orders?page=1&pageSize=10&status=completed
```

### 3. TypeScript 类型支持

```typescript
interface User {
  id: number;
  username: string;
  email: string;
}

interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
}

// 使用泛型指定返回类型
const loginResult = await apiClient.post<LoginResponse>('/api/v1/auth/login', {
  email: 'user@example.com',
  password: 'password'
});

// TypeScript 会自动推断 loginResult 的类型
console.log(loginResult.accessToken); // ✅ 类型安全
```

## 认证机制详解

### 1. API签名认证（所有请求自动包含）

API客户端会为每个请求自动生成以下认证头：

```
X-App-Id: test-app-001
X-Timestamp: 1758021611
X-Nonce: abc123def456
X-Sign: d4f053be19a8c40a6a47...
```

**签名算法**：
1. 收集所有参数（appId, timestamp, nonce, 请求参数, 请求体）
2. 按参数名排序
3. 构建签名字符串：`key1=value1&key2=value2`
4. 使用 HMAC-SHA256 + AppSecret 生成签名

### 2. JWT Token认证（受保护端点自动添加）

```typescript
// 登录后，token会自动保存到 localStorage
const loginResult = await apiClient.post('/api/v1/auth/login', loginData);
localStorage.setItem('auth_token', loginResult.accessToken);

// 后续的受保护端点请求会自动添加 Authorization 头
const userInfo = await apiClient.get('/api/v1/auth/verify');
// 自动添加: Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### 3. 公开端点 vs 受保护端点

**公开端点**（只需要API签名，不需要JWT token）：
- `/auth/login`
- `/auth/register`
- `/auth/reset-password`
- `/auth/forgot-password`

**受保护端点**（需要API签名 + JWT token）：
- 其他所有端点

```typescript
// 公开端点 - 只有API签名
await apiClient.post('/api/v1/auth/login', credentials);

// 受保护端点 - API签名 + JWT token
await apiClient.get('/api/v1/auth/verify');
await apiClient.get('/api/v1/orders');
await apiClient.post('/api/v1/payment/create', paymentData);
```

## 错误处理

### 1. 自动错误处理

```typescript
try {
  const result = await apiClient.post('/api/v1/some-endpoint', data);
  console.log('成功:', result);
} catch (error) {
  console.error('请求失败:', error.message);
  
  // 错误对象包含额外信息
  if (error.status === 401) {
    console.log('认证失败，请重新登录');
  }
  
  if (error.data) {
    console.log('服务器返回的错误数据:', error.data);
  }
}
```

### 2. 自动token清理和跳转

当遇到401错误时，API客户端会自动：
1. 清除过期的token
2. 清除用户信息
3. 跳转到登录页面（如果当前不在登录页面）

```typescript
// 401错误自动处理
try {
  await apiClient.get('/api/v1/protected-endpoint');
} catch (error) {
  // API客户端已经自动清理了token并跳转
  // 你只需要处理UI反馈
}
```

## 环境配置

### 1. 开发环境

```typescript
// 开发环境使用完整的后端地址
// baseURL: http://localhost:8089
```

### 2. 生产环境

```typescript
// 生产环境使用相对路径，配合 vercel.json 代理
// baseURL: '' (空字符串，使用相对路径)
```

### 3. 自定义配置

```typescript
import { createApiClient } from '@/lib/api-client';

const customApi = createApiClient({
  appId: 'my-custom-app',
  appSecret: 'my-custom-secret',
  baseURL: 'https://apiClient.myservice.com',
  timeout: 30000
});

// 使用自定义客户端
const result = await customApi.get('/api/v1/data');
```

## 实际使用示例

### 1. 用户认证流程

```typescript
// 1. 用户登录
try {
  const loginResult = await apiClient.post('/api/v1/auth/login', {
    email: 'user@example.com',
    password: 'password123'
  });
  
  // 2. 保存token
  localStorage.setItem('auth_token', loginResult.accessToken);
  localStorage.setItem('auth_user', JSON.stringify(loginResult.user));
  
  // 3. 获取用户信息（自动使用token）
  const userInfo = await apiClient.get('/api/v1/auth/verify');
  console.log('用户信息:', userInfo);
  
} catch (error) {
  console.error('登录失败:', error.message);
}
```

### 2. 数据获取

```typescript
// 获取订单列表（带分页和筛选）
const getOrders = async (filters = {}) => {
  try {
    const orders = await apiClient.get('/api/v1/orders', {
      page: 1,
      pageSize: 20,
      ...filters
    });
    
    return orders;
  } catch (error) {
    console.error('获取订单失败:', error.message);
    throw error;
  }
};

// 使用
const orders = await getOrders({
  status: 'completed',
  payWay: 'alipay'
});
```

### 3. 数据提交

```typescript
// 创建支付订单
const createPayment = async (paymentData) => {
  try {
    const result = await apiClient.post('/api/v1/payment/create', {
      amount: paymentData.amount,
      subject: paymentData.subject,
      payWay: paymentData.payWay,
      orderNo: paymentData.orderNo,
      // ... 其他参数
    });
    
    return result;
  } catch (error) {
    console.error('创建支付订单失败:', error.message);
    throw error;
  }
};
```

### 4. 文件上传（如果需要）

```typescript
// 注意：当前版本的API客户端主要处理JSON数据
// 如果需要上传文件，可能需要特殊处理
const uploadFile = async (file) => {
  const formData = new FormData();
  formData.append('file', file);
  
  // 对于文件上传，可能需要不同的处理方式
  // 这取决于后端的具体实现
};
```

## 调试和日志

API客户端包含内置的调试日志：

```typescript
// 发送请求时会自动输出
// 🌐 POST http://localhost:8089/api/v1/auth/login
// 📤 Headers: ['Content-Type', 'X-App-Id', 'X-Timestamp', ...]
// 🔐 签名字符串: appId=test-app-001&nonce=...&requestBody=...
// ✅ Request successful
```

## 最佳实践

### 1. 统一错误处理

```typescript
// 创建一个通用的API调用包装器
const apiCall = async (apiFunction, errorMessage = '操作失败') => {
  try {
    return await apiFunction();
  } catch (error) {
    console.error(errorMessage, error.message);
    
    // 统一的错误处理逻辑
    if (error.status === 401) {
      // 已经自动跳转到登录页面
      return;
    }
    
    // 显示用户友好的错误信息
    toast.error(error.message || errorMessage);
    throw error;
  }
};

// 使用
const orders = await apiCall(
  () => apiClient.get('/api/v1/orders'),
  '获取订单列表失败'
);
```

### 2. TypeScript 接口定义

```typescript
// 定义API响应接口
interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

interface Order {
  id: number;
  orderNo: string;
  amount: number;
  status: string;
  createdAt: string;
}

// 使用接口
const orders = await apiClient.get<ApiResponse<Order[]>>('/api/v1/orders');
```

### 3. 请求拦截和响应处理

```typescript
// 如果需要全局的请求拦截，可以扩展API客户端
class ExtendedApiClient extends ApiClient {
  async request(url, config, data) {
    // 请求前处理
    console.log('发送请求:', url);
    
    try {
      const result = await super.request(url, config, data);
      
      // 响应后处理
      console.log('请求成功:', result);
      return result;
    } catch (error) {
      // 错误处理
      console.error('请求失败:', error);
      throw error;
    }
  }
}
```

## 故障排除

### 1. 签名验证失败

```
错误: 签名验证失败
```

**可能原因**：
- AppSecret 配置错误
- 时间戳超出有效期（5分钟）
- 请求体格式不正确

**解决方案**：
- 检查 `DEFAULT_CONFIG` 中的 `appSecret`
- 确保系统时间正确
- 检查请求数据格式

### 2. 认证令牌错误

```
错误: 未提供认证令牌
```

**可能原因**：
- Token 已过期
- Token 未正确保存
- 端点被错误识别为受保护端点

**解决方案**：
- 重新登录获取新token
- 检查 localStorage 中的 `auth_token`
- 确认端点是否应该是公开端点

### 3. 网络连接问题

```
错误: 连接被拒绝
```

**可能原因**：
- 后端服务未启动
- 网络问题
- URL配置错误

**解决方案**：
- 确认后端服务运行在 `http://localhost:8089`
- 检查网络连接
- 验证环境变量 `NEXT_PUBLIC_API_BASE_URL`

## 总结

`api-client.ts` 提供了一个简洁、功能完整的网络请求解决方案：

✅ **简单易用**：统一的 `apiClient.get()`, `apiClient.post()` 接口
✅ **自动认证**：无需手动处理签名和token
✅ **类型安全**：完整的 TypeScript 支持
✅ **错误处理**：统一的错误处理和自动恢复
✅ **环境适配**：开发和生产环境的自动切换

这个工具类完全替代了之前复杂的多文件认证系统，提供了更好的开发体验和维护性。
