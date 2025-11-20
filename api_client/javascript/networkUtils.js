/**
 * 网络请求工具类
 * 统一管理所有API请求，自动添加认证参数
 * 支持GET、POST、PUT、DELETE等HTTP方法
 */

import { generateAuthHeaders } from '../utils/apiSign.js';

// API配置
const API_CONFIG = {
  BASE_URL: 'http://127.0.0.1:8089/api/v1',
  TIMEOUT: 10000,
  RETRY_COUNT: 3,
  RETRY_DELAY: 1000,
};

/**
 * HTTP状态码常量
 */
export const HTTP_STATUS = {
  OK: 200,
  CREATED: 201,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  INTERNAL_SERVER_ERROR: 500,
};

/**
 * 网络错误类型
 */
export class NetworkError extends Error {
  constructor(message, status, data = null) {
    super(message);
    this.name = 'NetworkError';
    this.status = status;
    this.data = data;
  }
}

/**
 * 网络工具类
 */
export class NetworkUtils {
  constructor(config = {}) {
    this.config = {
      ...API_CONFIG,
      ...config
    };
    this.requestInterceptors = [];
    this.responseInterceptors = [];
  }

  /**
   * 添加请求拦截器
   * @param {Function} interceptor 拦截器函数
   */
  addRequestInterceptor(interceptor) {
    this.requestInterceptors.push(interceptor);
  }

  /**
   * 添加响应拦截器
   * @param {Function} interceptor 拦截器函数
   */
  addResponseInterceptor(interceptor) {
    this.responseInterceptors.push(interceptor);
  }

  /**
   * 获取存储的token
   * @returns {string|null} token
   */
  getToken() {
    return localStorage.getItem('vtranslink_access_token');
  }

  /**
   * 设置token
   * @param {string} token 
   */
  setToken(token) {
    if (token) {
      localStorage.setItem('vtranslink_access_token', token);
    } else {
      localStorage.removeItem('vtranslink_access_token');
    }
  }

  /**
   * 清除认证信息
   */
  clearAuth() {
    localStorage.removeItem('vtranslink_access_token');
    localStorage.removeItem('admin_token');
    localStorage.removeItem('admin_info');
    localStorage.removeItem('vtranslink_user_info');
  }

  /**
   * 构建完整的URL
   * @param {string} endpoint API端点
   * @returns {string} 完整URL
   */
  buildUrl(endpoint) {
    // 如果endpoint已经是完整URL，直接返回
    if (endpoint.startsWith('http://') || endpoint.startsWith('https://')) {
      return endpoint;
    }
    
    // 确保endpoint以/开头
    const normalizedEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;
    return `${this.config.BASE_URL}${normalizedEndpoint}`;
  }

  /**
   * 构建请求头
   * @param {Object} customHeaders 自定义请求头
   * @param {Object} requestData 请求数据（用于生成签名）
   * @param {string} method HTTP方法
   * @returns {Promise<Object>} 完整的请求头
   */
  async buildHeaders(customHeaders = {}, requestData = null, method = 'GET') {
    // 基础请求头
    const baseHeaders = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      ...customHeaders
    };

    // 添加Bearer token（如果存在）
    const token = this.getToken();
    if (token) {
      baseHeaders.Authorization = `Bearer ${token}`;
    }

    try {
      // 为POST/PUT请求生成API认证头
      const shouldIncludeRequestBody = ['POST', 'PUT', 'PATCH'].includes(method.toUpperCase());
      const authHeaders = await generateAuthHeaders(shouldIncludeRequestBody ? requestData : undefined);
      
      // 合并所有请求头
      const finalHeaders = {
        ...baseHeaders,
        ...authHeaders
      };

      console.log(`[NetworkUtils] ${method} 请求头:`, finalHeaders);
      return finalHeaders;
    } catch (error) {
      console.error('[NetworkUtils] 生成认证头失败:', error);
      // 如果认证头生成失败，返回基础头（保持向后兼容）
      return baseHeaders;
    }
  }

  /**
   * 处理响应
   * @param {Response} response fetch响应对象
   * @returns {Promise<Object>} 处理后的响应数据
   */
  async handleResponse(response) {
    const contentType = response.headers.get('content-type');
    let data;

    // 根据内容类型解析响应
    if (contentType && contentType.includes('application/json')) {
      data = await response.json();
    } else {
      data = await response.text();
    }

    // 执行响应拦截器
    for (const interceptor of this.responseInterceptors) {
      data = await interceptor(data, response);
    }

    // 检查响应状态
    if (!response.ok) {
      const error = new NetworkError(
        data.message || `HTTP ${response.status}: ${response.statusText}`,
        response.status,
        data
      );
      
      // 处理401未授权错误
      if (response.status === HTTP_STATUS.UNAUTHORIZED) {
        console.warn('[NetworkUtils] 检测到401错误，清除认证信息');
        this.clearAuth();
        // 可以在这里触发重新登录逻辑
        window.dispatchEvent(new CustomEvent('auth:unauthorized'));
      }
      
      throw error;
    }

    return data;
  }

  /**
   * 延时函数
   * @param {number} ms 延时毫秒数
   * @returns {Promise}
   */
  delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * 执行HTTP请求（带重试机制）
   * @param {string} url 请求URL
   * @param {Object} options fetch选项
   * @param {number} retryCount 重试次数
   * @returns {Promise<Object>} 响应数据
   */
  async executeRequest(url, options, retryCount = this.config.RETRY_COUNT) {
    try {
      console.log(`[NetworkUtils] 发送请求: ${options.method || 'GET'} ${url}`);
      
      // 设置超时
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.config.TIMEOUT);
      
      const response = await fetch(url, {
        ...options,
        signal: controller.signal
      });
      
      clearTimeout(timeoutId);
      return await this.handleResponse(response);
      
    } catch (error) {
      console.error(`[NetworkUtils] 请求失败:`, error);
      
      // 如果是网络错误且还有重试次数
      if (retryCount > 0 && (error.name === 'TypeError' || error.name === 'AbortError')) {
        console.log(`[NetworkUtils] 重试请求 (剩余${retryCount}次)`);
        await this.delay(this.config.RETRY_DELAY);
        return this.executeRequest(url, options, retryCount - 1);
      }
      
      throw error;
    }
  }

  /**
   * 通用请求方法
   * @param {string} method HTTP方法
   * @param {string} endpoint API端点
   * @param {Object} data 请求数据
   * @param {Object} options 额外选项
   * @returns {Promise<Object>} 响应数据
   */
  async request(method, endpoint, data = null, options = {}) {
    const url = this.buildUrl(endpoint);
    const headers = await this.buildHeaders(options.headers, data, method);
    
    // 执行请求拦截器
    let requestConfig = {
      method: method.toUpperCase(),
      headers,
      ...options
    };

    // 添加请求体（仅对需要body的请求方法）
    if (data && ['POST', 'PUT', 'PATCH'].includes(method.toUpperCase())) {
      requestConfig.body = JSON.stringify(data);
    }

    // 执行请求拦截器
    for (const interceptor of this.requestInterceptors) {
      requestConfig = await interceptor(requestConfig);
    }

    return this.executeRequest(url, requestConfig);
  }

  /**
   * GET请求
   * @param {string} endpoint API端点
   * @param {Object} params 查询参数
   * @param {Object} options 额外选项
   * @returns {Promise<Object>} 响应数据
   */
  async get(endpoint, params = {}, options = {}) {
    // 构建查询字符串
    const queryString = new URLSearchParams(params).toString();
    const finalEndpoint = queryString ? `${endpoint}?${queryString}` : endpoint;
    
    return this.request('GET', finalEndpoint, null, options);
  }

  /**
   * POST请求
   * @param {string} endpoint API端点
   * @param {Object} data 请求数据
   * @param {Object} options 额外选项
   * @returns {Promise<Object>} 响应数据
   */
  async post(endpoint, data = {}, options = {}) {
    return this.request('POST', endpoint, data, options);
  }

  /**
   * PUT请求
   * @param {string} endpoint API端点
   * @param {Object} data 请求数据
   * @param {Object} options 额外选项
   * @returns {Promise<Object>} 响应数据
   */
  async put(endpoint, data = {}, options = {}) {
    return this.request('PUT', endpoint, data, options);
  }

  /**
   * DELETE请求
   * @param {string} endpoint API端点
   * @param {Object} options 额外选项
   * @returns {Promise<Object>} 响应数据
   */
  async delete(endpoint, options = {}) {
    return this.request('DELETE', endpoint, null, options);
  }

  /**
   * PATCH请求
   * @param {string} endpoint API端点
   * @param {Object} data 请求数据
   * @param {Object} options 额外选项
   * @returns {Promise<Object>} 响应数据
   */
  async patch(endpoint, data = {}, options = {}) {
    return this.request('PATCH', endpoint, data, options);
  }

  /**
   * 上传文件
   * @param {string} endpoint API端点
   * @param {FormData} formData 文件数据
   * @param {Object} options 额外选项
   * @returns {Promise<Object>} 响应数据
   */
  async upload(endpoint, formData, options = {}) {
    const url = this.buildUrl(endpoint);
    
    // 对于文件上传，不设置Content-Type，让浏览器自动设置
    const headers = await this.buildHeaders(
      { ...options.headers }, 
      null, // 文件上传不需要JSON序列化
      'POST'
    );
    
    // 删除Content-Type，让浏览器自动设置（包含boundary）
    delete headers['Content-Type'];

    const requestConfig = {
      method: 'POST',
      headers,
      body: formData,
      ...options
    };

    return this.executeRequest(url, requestConfig);
  }

  /**
   * 下载文件
   * @param {string} endpoint API端点
   * @param {string} filename 文件名
   * @param {Object} params 查询参数
   * @returns {Promise<void>}
   */
  async download(endpoint, filename, params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const finalEndpoint = queryString ? `${endpoint}?${queryString}` : endpoint;
    const url = this.buildUrl(finalEndpoint);
    
    const headers = await this.buildHeaders({}, null, 'GET');
    
    const response = await fetch(url, {
      method: 'GET',
      headers
    });

    if (!response.ok) {
      throw new NetworkError(`下载失败: ${response.statusText}`, response.status);
    }

    const blob = await response.blob();
    const downloadUrl = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = downloadUrl;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(downloadUrl);
  }
}

// 创建默认实例
const networkUtils = new NetworkUtils();

// 添加默认的响应拦截器
networkUtils.addResponseInterceptor(async (data, response) => {
  // 统一处理响应格式
  if (typeof data === 'object' && data !== null) {
    // 记录API响应日志
    console.log(`[NetworkUtils] API响应:`, {
      url: response.url,
      status: response.status,
      data: data
    });
  }
  return data;
});

// 监听授权失效事件
window.addEventListener('auth:unauthorized', () => {
  console.warn('[NetworkUtils] 用户授权失效，跳转到登录页面');
  // 可以在这里添加跳转到登录页面的逻辑
  if (window.location.pathname !== '/admin/login' && window.location.pathname !== '/login') {
    window.location.href = '/admin/login';
  }
});

export default networkUtils;
