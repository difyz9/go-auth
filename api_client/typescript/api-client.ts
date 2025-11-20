/**
 * 统一网络请求工具类
 * 简化版本，包含：
 * 1. 基础HTTP请求功能
 * 2. 自动API签名认证 (X-App-Id, X-Sign等)
 * 3. 自动JWT token认证 (Authorization: Bearer)
 * 4. 自动错误处理
 */

import crypto from 'crypto';

// 请求配置接口
interface RequestConfig {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  headers?: Record<string, string>;
  params?: Record<string, unknown>;
  timeout?: number;
}

// API认证配置
interface ApiConfig {
  appId: string;
  appSecret: string;
  baseURL?: string;
  timeout?: number;
}

/**
 * 统一网络请求客户端
 */
export class ApiClient {
  private config: ApiConfig;
  private baseURL: string;

  constructor(config: ApiConfig) {
    this.config = config;
    
    // 环境相关的baseURL设置
    if (process.env.NODE_ENV === 'production' || process.env.VERCEL === '1') {
      this.baseURL = config.baseURL || '';  // 生产环境使用相对路径
    } else {
      this.baseURL = config.baseURL || process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8089';
    }
  }

  /**
   * 生成API签名
   */
  private generateSignature(params: Record<string, unknown>, body?: unknown): {
    timestamp: string;
    nonce: string;
    sign: string;
  } {
    const timestamp = Math.floor(Date.now() / 1000).toString();
    const nonce = Math.random().toString(36).substring(2, 15);
    
    // 构建签名参数（包含基础认证参数）
    const signParams: Record<string, string> = {
      appId: this.config.appId,
      timestamp: timestamp,
      nonce: nonce,
    };
    
    // 添加查询参数
    if (params && Object.keys(params).length > 0) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          signParams[key] = String(value);
        }
      });
    }
    
    // 添加请求体参数（对于POST请求）
    if (body && typeof body === 'string') {
      signParams['requestBody'] = body;
    }
    
    // 按key排序并构建签名字符串
    const sortedKeys = Object.keys(signParams).sort();
    const signString = sortedKeys
      .map(key => `${key}=${signParams[key]}`)
      .join('&');
    
    console.log('🔐 签名字符串:', signString.replace(this.config.appSecret, '***SECRET***'));
    
    // 使用HMAC-SHA256生成签名
    const sign = crypto.createHmac('sha256', this.config.appSecret)
      .update(signString)
      .digest('hex');
    
    return { timestamp, nonce, sign };
  }

  /**
   * 获取JWT token
   */
  private getJwtToken(): string | null {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('auth_token');
    }
    return null;
  }

  /**
   * 检查是否为公开端点（不需要JWT token）
   */
  private isPublicEndpoint(url: string): boolean {
    return url.includes('/auth/login') ||
           url.includes('/auth/register') ||
           url.includes('/auth/reset-password') ||
           url.includes('/auth/forgot-password');
  }

  /**
   * 构建请求头
   */
  private buildHeaders(url: string, config: RequestConfig, body?: unknown): Record<string, string> {
    const params = config.params || {};
    const { timestamp, nonce, sign } = this.generateSignature(params, body);
    
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'X-App-Id': this.config.appId,
      'X-Timestamp': timestamp,
      'X-Nonce': nonce,
      'X-Sign': sign,
      ...config.headers,
    };

    // 为需要JWT认证的端点添加Authorization头
    if (!this.isPublicEndpoint(url)) {
      const token = this.getJwtToken();
      if (token) {
        headers.Authorization = `Bearer ${token}`;
      }
    }

    return headers;
  }

  /**
   * 构建完整URL
   */
  private buildUrl(url: string, params?: Record<string, unknown>): string {
    const fullUrl = url.startsWith('http') ? url : `${this.baseURL}${url}`;
    
    if (params && Object.keys(params).length > 0) {
      const searchParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          searchParams.append(key, String(value));
        }
      });
      return `${fullUrl}?${searchParams.toString()}`;
    }
    
    return fullUrl;
  }

  /**
   * 发送请求
   */
  private async request<T = unknown>(
    url: string, 
    config: RequestConfig = {}, 
    data?: unknown
  ): Promise<T> {
    const method = config.method || 'GET';
    const timeout = config.timeout || this.config.timeout || 10000;
    
    // 准备请求体
    const body = data ? JSON.stringify(data) : undefined;
    
    // 构建请求头
    const headers = this.buildHeaders(url, config, body);
    
    // 构建URL（GET请求需要添加查询参数）
    const requestUrl = method === 'GET' ? this.buildUrl(url, config.params) : this.buildUrl(url);
    
    console.log(`🌐 ${method} ${requestUrl}`);
    console.log('📤 Headers:', Object.keys(headers));
    
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), timeout);

      const response = await fetch(requestUrl, {
        method,
        headers,
        body: method !== 'GET' ? body : undefined,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      const responseText = await response.text();
      let responseData: unknown;

      try {
        responseData = JSON.parse(responseText);
      } catch {
        responseData = responseText;
      }

      if (!response.ok) {
        const error = new Error((responseData as { message?: string })?.message || `HTTP ${response.status}`);
        (error as unknown as { status: number; data: unknown }).status = response.status;
        (error as unknown as { status: number; data: unknown }).data = responseData;
        throw error;
      }

      console.log('✅ Request successful');
      return (responseData as { data?: T }).data || (responseData as T);
      
    } catch (error: unknown) {
      console.error('❌ Request failed:', (error as Error).message);
      
      // 处理认证错误
      if ((error as { status?: number }).status === 401 && typeof window !== 'undefined') {
        localStorage.removeItem('auth_token');
        localStorage.removeItem('auth_user');
        // 如果是认证错误且不是登录页面，可以跳转到登录页
        if (!window.location.pathname.includes('/auth/login')) {
          window.location.href = '/auth/login';
        }
      }
      
      throw error;
    }
  }

  /**
   * GET 请求
   */
  async get<T = unknown>(url: string, params?: Record<string, unknown>): Promise<T> {
    return this.request<T>(url, { method: 'GET', params });
  }

  /**
   * POST 请求
   */
  async post<T = unknown>(url: string, data?: unknown): Promise<T> {
    return this.request<T>(url, { method: 'POST' }, data);
  }

  /**
   * PUT 请求
   */
  async put<T = unknown>(url: string, data?: unknown): Promise<T> {
    return this.request<T>(url, { method: 'PUT' }, data);
  }

  /**
   * DELETE 请求
   */
  async delete<T = unknown>(url: string): Promise<T> {
    return this.request<T>(url, { method: 'DELETE' });
  }

  /**
   * PATCH 请求
   */
  async patch<T = unknown>(url: string, data?: unknown): Promise<T> {
    return this.request<T>(url, { method: 'PATCH' }, data);
  }
}

/**
 * 默认API客户端配置
 */
const DEFAULT_CONFIG: ApiConfig = {
  appId: 'test-app-001',
  appSecret: 'tmcf5m6qcm6k9hrp3sy8rhgafu00ttph',
};

/**
 * 默认API客户端实例
 */
export const apiClient = new ApiClient(DEFAULT_CONFIG);

/**
 * 创建自定义API客户端
 */
export function createApiClient(config: ApiConfig): ApiClient {
  return new ApiClient(config);
}

// 默认导出 - 使用更明确的名称
export default apiClient;
