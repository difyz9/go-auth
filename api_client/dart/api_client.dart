import 'dart:convert';
import 'dart:math';
import 'dart:typed_data';
import 'package:dio/dio.dart';
import 'package:crypto/crypto.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:encrypt/encrypt.dart' as encrypt;

/// API客户端配置
class ApiConfig {
  final String appId;
  final String appSecret;
  final String? baseURL;
  final int? timeout;
  final String? aesKey;  // 可选的AES密钥（用于加解密）
  final bool? enableEncryption;  // 是否启用加解密

  const ApiConfig({
    required this.appId,
    required this.appSecret,
    this.baseURL,
    this.timeout,
    this.aesKey,
    this.enableEncryption,
  });
}

/// 请求配置
class RequestConfig {
  final Map<String, dynamic>? params;
  final Map<String, String>? headers;
  final int? timeout;

  const RequestConfig({
    this.params,
    this.headers,
    this.timeout,
  });
}

/// API响应类型
class ApiResponse<T> {
  final int code;
  final String message;
  final T? data;

  const ApiResponse({
    required this.code,
    required this.message,
    this.data,
  });

  factory ApiResponse.fromJson(Map<String, dynamic> json, T? data) {
    return ApiResponse<T>(
      code: json['code'] ?? 0,
      message: json['message'] ?? '',
      data: data,
    );
  }
}

/// 加密请求结构
class EncryptedRequest {
  final String data;
  final int timestamp;

  const EncryptedRequest({
    required this.data,
    required this.timestamp,
  });

  Map<String, dynamic> toJson() {
    return {
      'data': data,
      'timestamp': timestamp,
    };
  }
}

/// 加密响应结构
class EncryptedResponse {
  final bool encrypted;
  final String data;
  final int timestamp;

  const EncryptedResponse({
    required this.encrypted,
    required this.data,
    required this.timestamp,
  });

  factory EncryptedResponse.fromJson(Map<String, dynamic> json) {
    return EncryptedResponse(
      encrypted: json['encrypted'] ?? false,
      data: json['data'] ?? '',
      timestamp: json['timestamp'] ?? 0,
    );
  }
}

/// 统一网络请求客户端
/// 
/// 功能特性：
/// 1. 自动API签名认证 (X-App-Id, X-Sign等)
/// 2. 自动JWT token认证 (Authorization: Bearer)
/// 3. 智能端点识别（公开/受保护端点）
/// 4. 统一错误处理和自动token清理
/// 5. 环境适配和请求日志
class ApiClient {
  final ApiConfig _config;
  final Dio _dio;
  final String _baseURL;

  // SharedPreferences 实例
  SharedPreferences? _prefs;

  // 公开端点列表（不需要JWT token）
  static const List<String> _publicEndpoints = [
    '/auth/login',
    '/auth/register',
    '/auth/reset-password',
    '/auth/forgot-password',
  ];

  ApiClient(this._config) : 
    _baseURL = _config.baseURL ?? 'http://localhost:8089',
    _dio = Dio() {
    
    // 配置Dio
    _dio.options.baseUrl = _baseURL;
    _dio.options.connectTimeout = Duration(milliseconds: _config.timeout ?? 10000);
    _dio.options.receiveTimeout = Duration(milliseconds: _config.timeout ?? 10000);
    _dio.options.sendTimeout = Duration(milliseconds: _config.timeout ?? 10000);

    // 添加请求拦截器
    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: _onRequest,
      onResponse: _onResponse,
      onError: _onError,
    ));

    // 初始化SharedPreferences
    _initPreferences();
  }

  /// 检查是否启用加密
  bool get _encryptionEnabled => _config.enableEncryption ?? false;

  /// 获取AES密钥
  String get _aesKey => _config.aesKey ?? '';

  /// AES加密
  String? _aesEncrypt(String plaintext) {
    if (_aesKey.isEmpty) return null;
    
    try {
      final key = encrypt.Key.fromUtf8(_normalizeKey(_aesKey));
      final iv = encrypt.IV.fromSecureRandom(16);
      final encrypter = encrypt.Encrypter(encrypt.AES(key, mode: encrypt.AESMode.cbc));
      
      final encrypted = encrypter.encrypt(plaintext, iv: iv);
      
      // 将IV和密文合并
      final combined = Uint8List.fromList([...iv.bytes, ...encrypted.bytes]);
      return base64.encode(combined);
    } catch (e) {
      print('❌ AES加密失败: $e');
      return null;
    }
  }

  /// AES解密
  String? _aesDecrypt(String ciphertext) {
    if (_aesKey.isEmpty) return null;
    
    try {
      final combined = base64.decode(ciphertext);
      if (combined.length < 16) {
        throw Exception('密文长度太短');
      }
      
      final iv = encrypt.IV(Uint8List.fromList(combined.sublist(0, 16)));
      final encrypted = encrypt.Encrypted(Uint8List.fromList(combined.sublist(16)));
      
      final key = encrypt.Key.fromUtf8(_normalizeKey(_aesKey));
      final encrypter = encrypt.Encrypter(encrypt.AES(key, mode: encrypt.AESMode.cbc));
      
      return encrypter.decrypt(encrypted, iv: iv);
    } catch (e) {
      print('❌ AES解密失败: $e');
      return null;
    }
  }

  /// 标准化AES密钥长度（16/24/32字节）
  String _normalizeKey(String key) {
    final bytes = utf8.encode(key);
    if (bytes.length >= 32) {
      return key.substring(0, 32);
    } else if (bytes.length >= 24) {
      return key.substring(0, 24);
    } else if (bytes.length >= 16) {
      return key.substring(0, 16);
    } else {
      return key.padRight(16, '0');
    }
  }

  /// 初始化SharedPreferences
  Future<void> _initPreferences() async {
    _prefs = await SharedPreferences.getInstance();
  }

  /// 生成API签名
  Map<String, String> _generateSignature(Map<String, dynamic> params, String? body) {
    final timestamp = (DateTime.now().millisecondsSinceEpoch / 1000).floor().toString();
    final nonce = _generateNonce();

    // 构建签名参数（包含基础认证参数）
    final signParams = <String, String>{
      'appId': _config.appId,
      'timestamp': timestamp,
      'nonce': nonce,
    };

    // 添加查询参数
    if (params.isNotEmpty) {
      params.forEach((key, value) {
        if (value != null) {
          signParams[key] = value.toString();
        }
      });
    }

    // 添加请求体参数（对于POST请求）
    if (body != null && body.isNotEmpty) {
      signParams['requestBody'] = body;
    }

    // 按key排序并构建签名字符串
    final sortedKeys = signParams.keys.toList()..sort();
    final signString = sortedKeys
        .map((key) => '$key=${signParams[key]}')
        .join('&');

    print('🔐 签名字符串: ${signString.replaceAll(_config.appSecret, '***SECRET***')}');

    // 使用HMAC-SHA256生成签名
    final key = utf8.encode(_config.appSecret);
    final bytes = utf8.encode(signString);
    final hmacSha256 = Hmac(sha256, key);
    final digest = hmacSha256.convert(bytes);
    final sign = digest.toString();

    return {
      'timestamp': timestamp,
      'nonce': nonce,
      'sign': sign,
    };
  }

  /// 生成随机字符串
  String _generateNonce() {
    const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
    final random = Random();
    return String.fromCharCodes(Iterable.generate(
      12, (_) => chars.codeUnitAt(random.nextInt(chars.length))
    ));
  }

  /// 获取JWT token
  Future<String?> _getJwtToken() async {
    if (_prefs == null) {
      await _initPreferences();
    }
    return _prefs?.getString('auth_token');
  }

  /// 保存JWT token
  Future<void> _saveJwtToken(String token) async {
    if (_prefs == null) {
      await _initPreferences();
    }
    await _prefs?.setString('auth_token', token);
    print('📝 Token已保存到SharedPreferences');
  }

  /// 清除JWT token
  Future<void> _clearJwtToken() async {
    if (_prefs == null) {
      await _initPreferences();
    }
    await _prefs?.remove('auth_token');
    await _prefs?.remove('auth_user');
    print('🗑️ Token已从SharedPreferences清除');
  }

  /// 检查是否为公开端点（不需要JWT token）
  bool _isPublicEndpoint(String url) {
    return _publicEndpoints.any((endpoint) => url.contains(endpoint));
  }

  /// 构建请求头
  Future<Map<String, String>> _buildHeaders(
    String url,
    RequestConfig? config,
    String? body,
  ) async {
    final params = config?.params ?? <String, dynamic>{};
    final signatureData = _generateSignature(params, body);

    final headers = <String, String>{
      'Content-Type': 'application/json',
      'X-App-Id': _config.appId,
      'X-Timestamp': signatureData['timestamp']!,
      'X-Nonce': signatureData['nonce']!,
      'X-Sign': signatureData['sign']!,
    };

    // 添加自定义头部
    if (config?.headers != null) {
      headers.addAll(config!.headers!);
    }

    // 为受保护端点添加Authorization头
    if (!_isPublicEndpoint(url)) {
      final token = await _getJwtToken();
      if (token != null && token.isNotEmpty) {
        headers['Authorization'] = 'Bearer $token';
      }
    }

    return headers;
  }

  /// 请求拦截器 - 在发送请求前
  Future<void> _onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    print('🌐 ${options.method} ${options.uri}');
    print('📤 Headers: ${options.headers.keys.toList()}');
    
    handler.next(options);
  }

  /// 响应拦截器 - 在收到响应后
  void _onResponse(
    Response response,
    ResponseInterceptorHandler handler,
  ) {
    print('📥 Response Status: ${response.statusCode}');
    print('✅ Request successful');
    
    handler.next(response);
  }

  /// 错误拦截器 - 在发生错误时
  Future<void> _onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    print('❌ Request failed: ${err.message}');
    
    // 处理认证错误
    if (err.response?.statusCode == 401) {
      print('🔒 认证失败，清除token');
      await _clearJwtToken();
      
      // 这里可以触发跳转到登录页面的回调
      // 例如：_onAuthError?.call();
    }
    
    handler.next(err);
  }

  /// 发送请求的核心方法
  Future<T> _request<T>(
    String url,
    String method, {
    dynamic requestData,
    RequestConfig? config,
    T Function(dynamic)? fromJson,
    bool? encrypt,  // 是否加密请求
    bool? requestEncrypted,  // 是否请求加密响应
  }) async {
    try {
      // 准备请求体
      String? body;
      dynamic finalRequestData = requestData;
      
      if (requestData != null && method != 'GET') {
        // 检查是否需要加密请求
        final shouldEncrypt = encrypt ?? _encryptionEnabled;
        
        if (shouldEncrypt && _aesKey.isNotEmpty) {
          final plaintext = jsonEncode(requestData);
          final encrypted = _aesEncrypt(plaintext);
          
          if (encrypted != null) {
            final encryptedRequest = EncryptedRequest(
              data: encrypted,
              timestamp: DateTime.now().millisecondsSinceEpoch ~/ 1000,
            );
            finalRequestData = encryptedRequest.toJson();
            print('🔐 请求已加密');
          }
        }
        
        body = jsonEncode(finalRequestData);
      }

      // 构建请求头
      final headers = await _buildHeaders(url, config, body);
      
      // 添加加密相关的请求头
      if (encrypt == true || _encryptionEnabled) {
        headers['X-Encrypted'] = '1';
      }
      if (requestEncrypted == true) {
        headers['X-Response-Encrypt'] = '1';
      }

      // 构建请求选项
      final options = Options(
        method: method,
        headers: headers,
        responseType: ResponseType.json,
      );

      // 发送请求
      final response = await _dio.request<Map<String, dynamic>>(
        url,
        data: method != 'GET' ? body : null,
        queryParameters: method == 'GET' ? config?.params : null,
        options: options,
      );

      // 解析响应
      final responseData = response.data;
      if (responseData == null) {
        throw DioException(
          requestOptions: response.requestOptions,
          message: '响应数据为空',
        );
      }

      // 检查是否是加密响应
      if (responseData['encrypted'] == true && responseData['data'] is String) {
        print('🔓 收到加密响应，正在解密...');
        final encryptedResp = EncryptedResponse.fromJson(responseData);
        final decrypted = _aesDecrypt(encryptedResp.data);
        
        if (decrypted == null) {
          throw DioException(
            requestOptions: response.requestOptions,
            message: '响应解密失败',
          );
        }
        
        // 解析解密后的JSON
        final decryptedData = jsonDecode(decrypted);
        print('✅ 响应解密成功');
        
        // 如果解密后的数据也包含标准响应格式，继续处理
        if (decryptedData is Map && decryptedData.containsKey('code')) {
          final code = decryptedData['code'] as int? ?? 0;
          final message = decryptedData['message'] as String? ?? '';
          final data = decryptedData['data'];
          
          if (code != 200) {
            throw DioException(
              requestOptions: response.requestOptions,
              message: message.isNotEmpty ? message : 'Error code: $code',
              response: response,
            );
          }
          
          if (fromJson != null && data != null) {
            return fromJson(data);
          } else if (data != null) {
            return data as T;
          }
        }
        
        // 直接返回解密后的数据
        return decryptedData as T;
      }

      // 处理未加密的响应
      final code = responseData['code'] as int? ?? 0;
      final message = responseData['message'] as String? ?? '';
      final data = responseData['data'];

      if (code != 200) {
        throw DioException(
          requestOptions: response.requestOptions,
          message: message.isNotEmpty ? message : 'HTTP ${response.statusCode}',
          response: response,
        );
      }

      // 返回数据
      if (fromJson != null && data != null) {
        return fromJson(data);
      } else if (data != null) {
        return data as T;
      } else {
        return responseData as T;
      }

    } on DioException catch (e) {
      print('❌ 请求异常: ${e.message}');
      
      // 重新抛出异常，保持错误信息
      rethrow;
    } catch (e) {
      print('❌ 未知错误: $e');
      throw DioException(
        requestOptions: RequestOptions(path: url),
        message: e.toString(),
      );
    }
  }

  /// GET 请求
  Future<T> get<T>(
    String url, {
    Map<String, dynamic>? params,
    T Function(dynamic)? fromJson,
    bool? requestEncrypted,  // 是否请求加密响应
  }) {
    return _request<T>(
      url,
      'GET',
      config: RequestConfig(params: params),
      fromJson: fromJson,
      requestEncrypted: requestEncrypted,
    );
  }

  /// POST 请求
  Future<T> post<T>(
    String url, {
    dynamic data,
    T Function(dynamic)? fromJson,
    bool? encrypt,  // 是否加密请求
    bool? requestEncrypted,  // 是否请求加密响应
  }) {
    return _request<T>(
      url,
      'POST',
      requestData: data,
      fromJson: fromJson,
      encrypt: encrypt,
      requestEncrypted: requestEncrypted,
    );
  }

  /// PUT 请求
  Future<T> put<T>(
    String url, {
    dynamic data,
    T Function(dynamic)? fromJson,
  }) {
    return _request<T>(
      url,
      'PUT',
      requestData: data,
      fromJson: fromJson,
    );
  }

  /// DELETE 请求
  Future<T> delete<T>(
    String url, {
    T Function(dynamic)? fromJson,
  }) {
    return _request<T>(
      url,
      'DELETE',
      fromJson: fromJson,
    );
  }

  /// PATCH 请求
  Future<T> patch<T>(
    String url, {
    dynamic data,
    T Function(dynamic)? fromJson,
  }) {
    return _request<T>(
      url,
      'PATCH',
      requestData: data,
      fromJson: fromJson,
    );
  }

  /// 手动保存token（用于登录后）
  Future<void> saveToken(String token) async {
    await _saveJwtToken(token);
  }

  /// 手动清除token（用于登出）
  Future<void> clearToken() async {
    await _clearJwtToken();
  }

  /// 获取当前token
  Future<String?> getToken() async {
    return await _getJwtToken();
  }

  /// 检查是否已登录
  Future<bool> isLoggedIn() async {
    final token = await _getJwtToken();
    return token != null && token.isNotEmpty;
  }

  /// 释放资源
  void dispose() {
    _dio.close();
  }
}

/// 默认API客户端配置
const defaultConfig = ApiConfig(
  appId: 'test-app-001',
  appSecret: 'tmcf5m6qcm6k9hrp3sy8rhgafu00ttph',
  aesKey: null,  // 可选：添加AES密钥以启用加解密
  enableEncryption: false,  // 默认不启用加密
);

/// 默认API客户端实例
final apiClient = ApiClient(defaultConfig);

/// 创建自定义API客户端
ApiClient createApiClient(ApiConfig config) {
  return ApiClient(config);
}

/// 创建启用加密的API客户端
ApiClient createEncryptedApiClient({
  required String appId,
  required String appSecret,
  required String aesKey,
  String? baseURL,
}) {
  return ApiClient(ApiConfig(
    appId: appId,
    appSecret: appSecret,
    aesKey: aesKey,
    baseURL: baseURL,
    enableEncryption: true,
  ));
}
