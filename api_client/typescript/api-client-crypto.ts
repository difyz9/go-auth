/**
 * API Client with Crypto Support (TypeScript)
 * 支持AES加解密的API客户端
 * 
 * 特性：
 * 1. HMAC-SHA256签名认证
 * 2. AES-256-CBC加密（可选）
 * 3. 请求/响应自动加解密
 * 4. 与Go后端完全兼容
 * 5. 完整的TypeScript类型支持
 */

import crypto from 'crypto';

/**
 * API配置接口
 */
export interface ApiConfig {
  appId: string;
  appSecret: string;
  baseURL?: string;
  timeout?: number;
  aesKey?: string; // 可选：AES密钥（用于加解密）
  enableEncryption?: boolean; // 是否启用加密
}

/**
 * 请求选项接口
 */
export interface RequestOptions {
  method?: string;
  data?: any;
  headers?: Record<string, string>;
  encrypt?: boolean; // 是否加密请求
  requestEncrypted?: boolean; // 是否请求加密响应
}

/**
 * 加密请求结构
 */
export interface EncryptedRequest {
  data: string;
  timestamp: number;
}

/**
 * 加密响应结构
 */
export interface EncryptedResponse {
  encrypted: boolean;
  data: string;
  timestamp: number;
}

/**
 * API响应接口
 */
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data?: T;
}

/**
 * 默认配置
 */
const defaultConfig: ApiConfig = {
  appId: 'test-app-001',
  appSecret: 'tmcf5m6qcm6k9hrp3sy8rhgafu00ttph',
  baseURL: 'http://localhost:8089',
  timeout: 10000,
  aesKey: undefined,
  enableEncryption: false
};

/**
 * API客户端类
 */
export class ApiClient {
  private config: ApiConfig;
  private baseURL: string;
  private token: string = '';

  constructor(config: Partial<ApiConfig> = {}) {
    this.config = { ...defaultConfig, ...config };
    this.baseURL = this.config.baseURL || 'http://localhost:8089';
    this.token = this.loadToken();
  }

  /**
   * 检查是否启用加密
   */
  private get encryptionEnabled(): boolean {
    return this.config.enableEncryption || false;
  }

  /**
   * 获取AES密钥
   */
  private get aesKey(): string {
    return this.config.aesKey || '';
  }

  /**
   * 加载Token
   */
  private loadToken(): string {
    // Node.js环境，可以从环境变量或配置文件加载
    return process.env.API_TOKEN || '';
  }

  /**
   * 保存Token
   */
  public setToken(token: string): void {
    this.token = token;
    // 在实际应用中，可以保存到环境变量或配置文件
  }

  /**
   * 清除Token
   */
  public clearToken(): void {
    this.token = '';
  }

  /**
   * 标准化AES密钥长度（16/24/32字节）
   */
  private normalizeKey(key: string): Buffer {
    const keyBytes = Buffer.from(key, 'utf8');
    
    if (keyBytes.length >= 32) {
      return keyBytes.slice(0, 32);
    } else if (keyBytes.length >= 24) {
      return keyBytes.slice(0, 24);
    } else if (keyBytes.length >= 16) {
      return keyBytes.slice(0, 16);
    } else {
      const padded = Buffer.alloc(16);
      keyBytes.copy(padded);
      return padded;
    }
  }

  /**
   * AES加密（AES-256-CBC）
   * @param plaintext - 明文
   * @returns Base64编码的密文
   */
  private aesEncrypt(plaintext: string): string | null {
    if (!this.aesKey) return null;

    try {
      const key = this.normalizeKey(this.aesKey);
      const iv = crypto.randomBytes(16);
      
      const cipher = crypto.createCipheriv('aes-256-cbc', key, iv);
      let encrypted = cipher.update(plaintext, 'utf8');
      encrypted = Buffer.concat([encrypted, cipher.final()]);
      
      // 合并IV和密文，然后Base64编码
      const combined = Buffer.concat([iv, encrypted]);
      return combined.toString('base64');
    } catch (error) {
      console.error('❌ AES加密失败:', error);
      return null;
    }
  }

  /**
   * AES解密（AES-256-CBC）
   * @param ciphertext - Base64编码的密文
   * @returns 明文
   */
  private aesDecrypt(ciphertext: string): string | null {
    if (!this.aesKey) return null;

    try {
      const combined = Buffer.from(ciphertext, 'base64');
      
      if (combined.length < 16) {
        throw new Error('密文长度太短');
      }
      
      const iv = combined.slice(0, 16);
      const encrypted = combined.slice(16);
      
      const key = this.normalizeKey(this.aesKey);
      const decipher = crypto.createDecipheriv('aes-256-cbc', key, iv);
      
      let decrypted = decipher.update(encrypted);
      decrypted = Buffer.concat([decrypted, decipher.final()]);
      
      return decrypted.toString('utf8');
    } catch (error) {
      console.error('❌ AES解密失败:', error);
      return null;
    }
  }

  /**
   * 生成随机字符串
   */
  private generateNonce(length: number = 16): string {
    return crypto.randomBytes(length).toString('hex').slice(0, length);
  }

  /**
   * 生成签名
   */
  private generateSign(url: string, body: string | null, timestamp: number, nonce: string): string {
    const params: Record<string, string> = {
      app_id: this.config.appId,
      timestamp: timestamp.toString(),
      nonce: nonce,
      url: url
    };
    
    if (body) {
      params.body = body;
    }
    
    // 按字母顺序排序
    const sortedKeys = Object.keys(params).sort();
    const signStr = sortedKeys.map(key => `${key}=${params[key]}`).join('&');
    
    // HMAC-SHA256签名
    const hmac = crypto.createHmac('sha256', this.config.appSecret);
    hmac.update(signStr);
    return hmac.digest('hex');
  }

  /**
   * 构建请求头
   */
  private buildHeaders(url: string, body: string | null, customHeaders: Record<string, string> = {}): Record<string, string> {
    const timestamp = Math.floor(Date.now() / 1000);
    const nonce = this.generateNonce();
    const sign = this.generateSign(url, body, timestamp, nonce);
    
    const headers: Record<string, string> = {
      'X-App-Id': this.config.appId,
      'X-Timestamp': timestamp.toString(),
      'X-Nonce': nonce,
      'X-Sign': sign,
      'Content-Type': 'application/json',
      ...customHeaders
    };
    
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }
    
    return headers;
  }

  /**
   * 发送请求的核心方法
   */
  public async request<T = any>(url: string, options: RequestOptions = {}): Promise<T> {
    try {
      const method = options.method || 'GET';
      let data = options.data || null;
      const headers = options.headers || {};
      const encrypt = options.encrypt !== undefined ? options.encrypt : this.encryptionEnabled;
      const requestEncrypted = options.requestEncrypted || false;

      // 构建完整URL
      const fullURL = url.startsWith('http') ? url : `${this.baseURL}${url}`;

      // 准备请求体
      let body: string | null = null;
      if (data && method !== 'GET') {
        // 检查是否需要加密请求
        if (encrypt && this.aesKey) {
          const plaintext = JSON.stringify(data);
          const encrypted = this.aesEncrypt(plaintext);
          
          if (encrypted) {
            data = {
              data: encrypted,
              timestamp: Math.floor(Date.now() / 1000)
            } as EncryptedRequest;
            console.log('🔐 请求已加密');
          }
        }
        
        body = JSON.stringify(data);
      }

      // 添加加密相关的请求头
      if (encrypt) {
        headers['X-Encrypted'] = '1';
      }
      if (requestEncrypted) {
        headers['X-Response-Encrypt'] = '1';
      }

      // 添加认证头
      const authHeaders = this.buildHeaders(url, body, headers);
      Object.assign(headers, authHeaders);

      // 发送请求（使用fetch或其他HTTP客户端）
      const response = await fetch(fullURL, {
        method,
        headers,
        body: body || undefined,
        signal: AbortSignal.timeout(this.config.timeout || 10000)
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      // 解析响应
      const responseData = await response.json() as any;

      // 检查是否是加密响应
      if (responseData.encrypted === true && typeof responseData.data === 'string') {
        console.log('🔓 收到加密响应，正在解密...');
        const decrypted = this.aesDecrypt(responseData.data);
        
        if (!decrypted) {
          throw new Error('响应解密失败');
        }
        
        // 解析解密后的JSON
        const decryptedData = JSON.parse(decrypted);
        console.log('✅ 响应解密成功');
        
        // 如果解密后的数据也包含标准响应格式，继续处理
        if (decryptedData.code !== undefined) {
          const apiResp = decryptedData as ApiResponse<T>;
          
          if (apiResp.code !== 200) {
            throw new Error(apiResp.message || `Error code: ${apiResp.code}`);
          }
          
          return apiResp.data !== undefined ? apiResp.data : decryptedData;
        }
        
        // 直接返回解密后的数据
        return decryptedData;
      }

      // 处理未加密的响应
      const apiResp = responseData as ApiResponse<T>;
      
      if (apiResp.code !== 200) {
        throw new Error(apiResp.message || `HTTP ${response.status}`);
      }

      return apiResp.data !== undefined ? apiResp.data : responseData;
    } catch (error) {
      console.error('❌ 请求失败:', error);
      throw error;
    }
  }

  /**
   * GET 请求
   */
  public async get<T = any>(url: string, params?: Record<string, any>, options: Omit<RequestOptions, 'method' | 'data'> = {}): Promise<T> {
    const queryString = params ? '?' + new URLSearchParams(params).toString() : '';
    return this.request<T>(url + queryString, { method: 'GET', ...options });
  }

  /**
   * POST 请求
   */
  public async post<T = any>(url: string, data?: any, options: Omit<RequestOptions, 'method' | 'data'> = {}): Promise<T> {
    return this.request<T>(url, { method: 'POST', data, ...options });
  }

  /**
   * PUT 请求
   */
  public async put<T = any>(url: string, data?: any, options: Omit<RequestOptions, 'method' | 'data'> = {}): Promise<T> {
    return this.request<T>(url, { method: 'PUT', data, ...options });
  }

  /**
   * DELETE 请求
   */
  public async delete<T = any>(url: string, options: Omit<RequestOptions, 'method' | 'data'> = {}): Promise<T> {
    return this.request<T>(url, { method: 'DELETE', ...options });
  }
}

/**
 * 创建启用加密的API客户端
 */
export function createEncryptedApiClient(config: {
  appId: string;
  appSecret: string;
  aesKey: string;
  baseURL?: string;
}): ApiClient {
  return new ApiClient({
    ...config,
    enableEncryption: true
  });
}

/**
 * 默认API客户端实例
 */
export const apiClient = new ApiClient();

export default ApiClient;
