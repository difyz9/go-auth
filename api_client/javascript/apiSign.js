/**
 * VTransLink API认证工具类
 * 基于payment_service的认证方式实现请求参数验证
 * 
 * 认证机制：
 * 1. 通过Header传递认证参数：X-App-Id, X-Timestamp, X-Nonce, X-Sign
 * 2. 签名算法：HMAC-SHA256
 * 3. 时间戳有效期：5分钟
 */

// 认证配置
const AUTH_CONFIG = {
  APP_ID: 'test-app-001',
  APP_SECRET: 'tmcf5m6qcm6k9hrp3sy8rhgafu00ttph', // 真实的APP_SECRET
  REQUIRE_SIGN: true
};

// 常量定义
export const API_AUTH_HEADERS = {
  APP_ID: 'X-App-Id',
  TIMESTAMP: 'X-Timestamp',
  NONCE: 'X-Nonce',
  SIGN: 'X-Sign',
};

export const API_AUTH_CONFIG = {
  TIMESTAMP_VALIDITY: 300, // 5分钟
  NONCE_LENGTH: 16,
  SIGN_ALGORITHM: 'HMAC-SHA256',
};

/**
 * API认证工具类
 */
export class ApiAuth {
  constructor(config) {
    this.config = {
      requireSign: true,
      ...config
    };
  }

  /**
   * 生成随机字符串（Nonce）
   * @param {number} length 长度，默认16
   * @returns {string} 随机字符串
   */
  generateNonce(length = 16) {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    
    // 使用crypto.getRandomValues生成更安全的随机数
    if (typeof crypto !== 'undefined' && crypto.getRandomValues) {
      const randomArray = new Uint8Array(length);
      crypto.getRandomValues(randomArray);
      
      for (let i = 0; i < length; i++) {
        result += chars.charAt(randomArray[i] % chars.length);
      }
    } else {
      // 降级方案
      for (let i = 0; i < length; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length));
      }
    }
    
    return result;
  }

  /**
   * 生成时间戳
   * @returns {string} Unix时间戳（秒）
   */
  generateTimestamp() {
    return Math.floor(Date.now() / 1000).toString();
  }

  /**
   * 验证时间戳是否在有效期内
   * @param {string} timestamp 时间戳
   * @returns {boolean} 是否有效
   */
  validateTimestamp(timestamp) {
    const ts = parseInt(timestamp, 10);
    if (isNaN(ts)) {
      return false;
    }

    const now = Math.floor(Date.now() / 1000);
    // 5分钟有效期
    return Math.abs(now - ts) <= API_AUTH_CONFIG.TIMESTAMP_VALIDITY;
  }

  /**
   * 异步生成API签名（使用Web Crypto API）
   * @param {Object} params 参数对象
   * @param {string} appSecret 应用密钥
   * @returns {Promise<string>} 签名字符串
   */
  async generateSignAsync(params, appSecret) {
    const secret = appSecret || this.config.appSecret;
    
    // 创建参数副本，过滤掉空值和sign参数
    const paramsCopy = {};
    for (const [key, value] of Object.entries(params)) {
      if (key !== 'sign' && value !== undefined && value !== null && value !== '') {
        paramsCopy[key] = value;
      }
    }

    // 参数排序
    const sortedKeys = Object.keys(paramsCopy).sort();

    // 拼接字符串
    const parts = sortedKeys.map(key => `${key}=${paramsCopy[key]}`);
    const signString = parts.join('&');

    console.log('Sign string:', signString);

    // 使用Web Crypto API生成HMAC-SHA256签名
    const encoder = new TextEncoder();
    const keyData = encoder.encode(secret);
    const dataData = encoder.encode(signString);

    const cryptoKey = await crypto.subtle.importKey(
      'raw',
      keyData,
      { name: 'HMAC', hash: 'SHA-256' },
      false,
      ['sign']
    );

    const signature = await crypto.subtle.sign('HMAC', cryptoKey, dataData);
    const hashArray = Array.from(new Uint8Array(signature));
    const hexString = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
    
    console.log('Generated signature:', hexString);
    return hexString;
  }

  /**
   * 验证API签名
   * @param {Object} params 参数对象
   * @param {string} receivedSign 接收到的签名
   * @param {string} appSecret 应用密钥
   * @returns {Promise<boolean>} 是否验证通过
   */
  async verifySign(params, receivedSign, appSecret) {
    try {
      const expectedSign = await this.generateSignAsync(params, appSecret);
      // 支持大小写不敏感的比较
      return expectedSign.toLowerCase() === receivedSign.toLowerCase();
    } catch (error) {
      console.error('签名验证失败:', error);
      return false;
    }
  }

  /**
   * 生成认证参数
   * @param {Object} extraParams 额外参数
   * @param {string} requestBody 请求体（可选，用于POST请求）
   * @returns {Promise<Object>} 认证参数对象
   */
  async generateAuthParams(extraParams = {}, requestBody) {
    const timestamp = this.generateTimestamp();
    const nonce = this.generateNonce();
    
    // 基础认证参数
    const authParams = {
      appId: this.config.appId,
      timestamp: timestamp,
      nonce: nonce,
      ...extraParams
    };

    // 如果需要签名验证
    let sign;
    if (this.config.requireSign) {
      // 构建签名参数，包含认证参数
      let paramsForSign = { ...authParams };
      
      // 如果有请求体（JSON格式），需要包含在签名计算中
      // 只有当requestBody存在且不为空时才加入签名
      if (requestBody && typeof requestBody === 'string' && requestBody.trim() !== '') {
        // 后端会将JSON请求体作为 requestBody 参数加入签名
        paramsForSign.requestBody = requestBody;
      }
      
      console.log('Params for sign:', paramsForSign);
      sign = await this.generateSignAsync(paramsForSign);
    }

    return {
      appId: authParams.appId,
      timestamp: authParams.timestamp,
      nonce: authParams.nonce,
      sign
    };
  }

  /**
   * 构建认证请求头
   * @param {Object} extraParams 额外参数
   * @param {string} requestBody 请求体（可选）
   * @returns {Promise<Object>} 请求头对象
   */
  async buildAuthHeaders(extraParams = {}, requestBody) {
    const authParams = await this.generateAuthParams(extraParams, requestBody);

    const headers = {
      [API_AUTH_HEADERS.APP_ID]: authParams.appId,
      [API_AUTH_HEADERS.TIMESTAMP]: authParams.timestamp,
      [API_AUTH_HEADERS.NONCE]: authParams.nonce,
    };

    if (authParams.sign) {
      headers[API_AUTH_HEADERS.SIGN] = authParams.sign;
    }

    return headers;
  }

  /**
   * 生成API签名示例（用于测试和调试）
   * @returns {Promise<Object>} 签名示例对象
   */
  async generateSignExample() {
    const timestamp = this.generateTimestamp();
    const nonce = this.generateNonce();

    const params = {
      appId: this.config.appId,
      timestamp: timestamp,
      nonce: nonce,
    };

    const sign = await this.generateSignAsync(params);

    // 构建签名字符串用于调试
    const sortedKeys = Object.keys(params).sort();
    const parts = sortedKeys.map(key => `${key}=${params[key]}`);
    const signString = parts.join('&');

    return {
      appId: this.config.appId,
      timestamp,
      nonce,
      sign,
      algorithm: API_AUTH_CONFIG.SIGN_ALGORITHM,
      signString
    };
  }

  /**
   * 更新配置
   * @param {Object} config 新的配置
   */
  updateConfig(config) {
    this.config = { ...this.config, ...config };
  }

  /**
   * 获取当前配置（不包含密钥）
   * @returns {Object} 配置信息
   */
  getConfig() {
    const { appSecret, ...configWithoutSecret } = this.config;
    return configWithoutSecret;
  }
}

// 创建默认实例
const defaultApiAuth = new ApiAuth({
  appId: AUTH_CONFIG.APP_ID,
  appSecret: AUTH_CONFIG.APP_SECRET,
  requireSign: AUTH_CONFIG.REQUIRE_SIGN
});

/**
 * 简化的API认证头生成函数（向后兼容）
 * @param {Object} requestData 请求数据（可选，用于POST请求）
 * @returns {Promise<Object>} 认证头对象
 */
export async function generateAuthHeaders(requestData = undefined) {
  try {
    // 只有当requestData不是undefined且不是null时，才转换为JSON字符串
    const requestBody = (requestData !== undefined && requestData !== null) 
      ? JSON.stringify(requestData) 
      : undefined;
    
    console.log('generateAuthHeaders - requestData:', requestData);
    console.log('generateAuthHeaders - requestBody:', requestBody);
    
    return await defaultApiAuth.buildAuthHeaders({}, requestBody);
  } catch (error) {
    console.error('生成认证头失败:', error);
    throw error;
  }
}

/**
 * 生成时间戳
 * @returns {string} Unix时间戳（秒）
 */
export function generateTimestamp() {
  return defaultApiAuth.generateTimestamp();
}

/**
 * 生成随机字符串
 * @param {number} length 长度
 * @returns {string} 随机字符串
 */
export function generateNonce(length = 16) {
  return defaultApiAuth.generateNonce(length);
}

/**
 * 验证时间戳是否有效（5分钟有效期）
 * @param {string} timestamp 时间戳
 * @returns {boolean} 是否有效
 */
export function validateTimestamp(timestamp) {
  return defaultApiAuth.validateTimestamp(timestamp);
}

/**
 * 生成API签名示例
 * @returns {Promise<Object>} 包含所有认证参数的对象
 */
export async function generateApiExample() {
  return await defaultApiAuth.generateSignExample();
}

// 导出默认实例和类
export { defaultApiAuth };
export default ApiAuth;
