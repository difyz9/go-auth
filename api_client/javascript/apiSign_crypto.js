/**
 * API Client with Crypto Support
 * 支持AES加解密的API客户端
 * 
 * 特性：
 * 1. HMAC-SHA256签名认证
 * 2. AES-256-CBC加密（可选）
 * 3. 请求/响应自动加解密
 * 4. 与Go后端完全兼容
 */

// API配置
const ApiConfig = {
  appId: 'test-app-001',
  appSecret: 'tmcf5m6qcm6k9hrp3sy8rhgafu00ttph',
  baseURL: 'http://localhost:8089',
  timeout: 10000,
  aesKey: null, // 可选：AES密钥（用于加解密）
  enableEncryption: false // 是否启用加密
};

/**
 * API客户端类
 */
class ApiClient {
  constructor(config = ApiConfig) {
    this.config = { ...ApiConfig, ...config };
    this.baseURL = this.config.baseURL;
    this.token = localStorage.getItem('token') || '';
  }

  // 检查是否启用加密
  get encryptionEnabled() {
    return this.config.enableEncryption || false;
  }

  // 获取AES密钥
  get aesKey() {
    return this.config.aesKey || '';
  }

  /**
   * 标准化AES密钥长度（16/24/32字节）
   */
  normalizeKey(key) {
    const encoder = new TextEncoder();
    const keyBytes = encoder.encode(key);
    
    if (keyBytes.length >= 32) {
      return key.substring(0, 32);
    } else if (keyBytes.length >= 24) {
      return key.substring(0, 24);
    } else if (keyBytes.length >= 16) {
      return key.substring(0, 16);
    } else {
      return key.padEnd(16, '0');
    }
  }

  /**
   * AES加密（AES-256-CBC）
   * @param {string} plaintext - 明文
   * @returns {Promise<string|null>} Base64编码的密文
   */
  async aesEncrypt(plaintext) {
    if (!this.aesKey) return null;

    try {
      const encoder = new TextEncoder();
      const data = encoder.encode(plaintext);
      const keyStr = this.normalizeKey(this.aesKey);
      const keyData = encoder.encode(keyStr);
      
      // 生成随机IV
      const iv = crypto.getRandomValues(new Uint8Array(16));
      
      // 导入密钥
      const cryptoKey = await crypto.subtle.importKey(
        'raw',
        keyData,
        { name: 'AES-CBC' },
        false,
        ['encrypt']
      );
      
      // 加密
      const encrypted = await crypto.subtle.encrypt(
        { name: 'AES-CBC', iv: iv },
        cryptoKey,
        data
      );
      
      // 合并IV和密文
      const combined = new Uint8Array(iv.length + encrypted.byteLength);
      combined.set(iv, 0);
      combined.set(new Uint8Array(encrypted), iv.length);
      
      // Base64编码
      return btoa(String.fromCharCode(...combined));
    } catch (error) {
      console.error('❌ AES加密失败:', error);
      return null;
    }
  }

  /**
   * AES解密（AES-256-CBC）
   * @param {string} ciphertext - Base64编码的密文
   * @returns {Promise<string|null>} 明文
   */
  async aesDecrypt(ciphertext) {
    if (!this.aesKey) return null;

    try {
      // Base64解码
      const combined = Uint8Array.from(atob(ciphertext), c => c.charCodeAt(0));
      
      if (combined.length < 16) {
        throw new Error('密文长度太短');
      }
      
      // 分离IV和密文
      const iv = combined.slice(0, 16);
      const encrypted = combined.slice(16);
      
      const encoder = new TextEncoder();
      const keyStr = this.normalizeKey(this.aesKey);
      const keyData = encoder.encode(keyStr);
      
      // 导入密钥
      const cryptoKey = await crypto.subtle.importKey(
        'raw',
        keyData,
        { name: 'AES-CBC' },
        false,
        ['decrypt']
      );
      
      // 解密
      const decrypted = await crypto.subtle.decrypt(
        { name: 'AES-CBC', iv: iv },
        cryptoKey,
        encrypted
      );
      
      // 转换为字符串
      const decoder = new TextDecoder();
      return decoder.decode(decrypted);
    } catch (error) {
      console.error('❌ AES解密失败:', error);
      return null;
    }
  }

  /**
   * 生成随机字符串
   */
  generateNonce(length = 16) {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    for (let i = 0; i < length; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
  }

  /**
   * 生成签名
   */
  async generateSign(url, body, timestamp, nonce) {
    const appId = this.config.appId;
    const appSecret = this.config.appSecret;
    
    // 构建签名参数
    const params = {
      app_id: appId,
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
    const encoder = new TextEncoder();
    const keyData = encoder.encode(appSecret);
    const msgData = encoder.encode(signStr);
    
    const cryptoKey = await crypto.subtle.importKey(
      'raw',
      keyData,
      { name: 'HMAC', hash: 'SHA-256' },
      false,
      ['sign']
    );
    
    const signature = await crypto.subtle.sign('HMAC', cryptoKey, msgData);
    
    // 转换为十六进制
    return Array.from(new Uint8Array(signature))
      .map(b => b.toString(16).padStart(2, '0'))
      .join('');
  }

  /**
   * 构建请求头
   */
  async buildHeaders(url, body = null, customHeaders = {}) {
    const timestamp = Math.floor(Date.now() / 1000);
    const nonce = this.generateNonce();
    const sign = await this.generateSign(url, body, timestamp, nonce);
    
    const headers = {
      'X-App-Id': this.config.appId,
      'X-Timestamp': timestamp.toString(),
      'X-Nonce': nonce,
      'X-Sign': sign,
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
  async request(url, options = {}) {
    try {
      const method = options.method || 'GET';
      let data = options.data || null;
      const headers = options.headers || {};
      const encrypt = options.encrypt !== undefined ? options.encrypt : this.encryptionEnabled;
      const requestEncrypted = options.requestEncrypted || false;

      // 构建完整URL
      const fullURL = url.startsWith('http') ? url : `${this.baseURL}${url}`;

      // 准备请求体
      let body = null;
      if (data && method !== 'GET') {
        // 检查是否需要加密请求
        if (encrypt && this.aesKey) {
          const plaintext = JSON.stringify(data);
          const encrypted = await this.aesEncrypt(plaintext);
          
          if (encrypted) {
            data = {
              data: encrypted,
              timestamp: Math.floor(Date.now() / 1000)
            };
            console.log('🔐 请求已加密');
          }
        }
        
        body = JSON.stringify(data);
        headers['Content-Type'] = 'application/json';
      }

      // 添加加密相关的请求头
      if (encrypt) {
        headers['X-Encrypted'] = '1';
      }
      if (requestEncrypted) {
        headers['X-Response-Encrypt'] = '1';
      }

      // 添加认证头
      const authHeaders = await this.buildHeaders(url, body, headers);
      Object.assign(headers, authHeaders);

      // 发送请求
      const response = await fetch(fullURL, {
        method,
        headers,
        body,
        timeout: this.config.timeout
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      // 解析响应
      const responseData = await response.json();

      // 检查是否是加密响应
      if (responseData.encrypted === true && typeof responseData.data === 'string') {
        console.log('🔓 收到加密响应，正在解密...');
        const decrypted = await this.aesDecrypt(responseData.data);
        
        if (!decrypted) {
          throw new Error('响应解密失败');
        }
        
        // 解析解密后的JSON
        const decryptedData = JSON.parse(decrypted);
        console.log('✅ 响应解密成功');
        
        // 如果解密后的数据也包含标准响应格式，继续处理
        if (decryptedData.code !== undefined) {
          const code = decryptedData.code || 0;
          const message = decryptedData.message || '';
          const resData = decryptedData.data;
          
          if (code !== 200) {
            throw new Error(message || `Error code: ${code}`);
          }
          
          return resData !== undefined ? resData : decryptedData;
        }
        
        // 直接返回解密后的数据
        return decryptedData;
      }

      // 处理未加密的响应
      const code = responseData.code || 0;
      const message = responseData.message || '';
      const resData = responseData.data;

      if (code !== 200) {
        throw new Error(message || `HTTP ${response.status}`);
      }

      return resData !== undefined ? resData : responseData;
    } catch (error) {
      console.error('❌ 请求失败:', error.message);
      throw error;
    }
  }

  /**
   * GET 请求
   */
  async get(url, params = null, options = {}) {
    const queryString = params ? '?' + new URLSearchParams(params).toString() : '';
    return this.request(url + queryString, { method: 'GET', ...options });
  }

  /**
   * POST 请求
   */
  async post(url, data = null, options = {}) {
    return this.request(url, { method: 'POST', data, ...options });
  }

  /**
   * PUT 请求
   */
  async put(url, data = null, options = {}) {
    return this.request(url, { method: 'PUT', data, ...options });
  }

  /**
   * DELETE 请求
   */
  async delete(url, options = {}) {
    return this.request(url, { method: 'DELETE', ...options });
  }

  /**
   * 设置Token
   */
  setToken(token) {
    this.token = token;
    localStorage.setItem('token', token);
  }

  /**
   * 清除Token
   */
  clearToken() {
    this.token = '';
    localStorage.removeItem('token');
  }
}

// 导出默认实例
export const apiClient = new ApiClient();

/**
 * 创建启用加密的API客户端
 */
export function createEncryptedApiClient({ appId, appSecret, aesKey, baseURL }) {
  return new ApiClient({
    appId,
    appSecret,
    aesKey,
    baseURL,
    enableEncryption: true
  });
}

// 导出类供自定义使用
export default ApiClient;
