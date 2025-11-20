/// API客户端测试配置
/// 基于真实的JavaScript测试配置
class TestConfig {
  // 后端服务配置
  static const String baseUrl = 'http://localhost:8089';
  static const String healthEndpoint = '/health';
  
  // API认证配置 (来自test-real-api-client.js)
  static const String appId = 'test-app-001';
  static const String appSecret = 'tmcf5m6qcm6k9hrp3sy8rhgafu00ttph';
  
  // 测试用户信息 (来自test-real-api-client.js)
  static const String testEmail = 'admin@126.com';
  static const String testPassword = 'Ab123456';
  
  // API端点配置
  static const String loginEndpoint = '/api/v1/auth/login';
  static const String verifyEndpoint = '/api/v1/auth/verify';
  static const String userProfileEndpoint = '/api/v1/user/profile';
  static const String paymentOrdersEndpoint = '/api/v1/payment/orders';
  static const String appConfigEndpoint = '/api/v1/app/config';
  
  // 超时配置
  static const int timeoutSeconds = 10;
  static const int healthCheckTimeoutSeconds = 5;
  
  // 性能测试配置
  static const int performanceTestCount = 10;
  
  // 预期响应字段
  static const List<String> expectedLoginFields = [
    'accessToken',
    'user',
  ];
  
  static const List<String> expectedUserFields = [
    'id',
    'username',
    'email',
    'isVip',
  ];
}

/// 测试辅助工具类
class TestHelper {
  /// 验证响应数据结构
  static bool validateLoginResponse(Map<String, dynamic> data) {
    for (final field in TestConfig.expectedLoginFields) {
      if (!data.containsKey(field)) {
        print('⚠️ 登录响应缺少字段: $field');
        return false;
      }
    }
    return true;
  }
  
  /// 验证用户信息结构
  static bool validateUserInfo(Map<String, dynamic> user) {
    for (final field in TestConfig.expectedUserFields) {
      if (!user.containsKey(field)) {
        print('⚠️ 用户信息缺少字段: $field');
        return false;
      }
    }
    return true;
  }
  
  /// 打印测试结果
  static void printTestResult(String testName, bool success, [String? message]) {
    final status = success ? '✅' : '❌';
    print('$status $testName: ${success ? "通过" : "失败"}');
    if (message != null) {
      print('   $message');
    }
  }
  
  /// 格式化Token显示
  static String formatToken(String token) {
    if (token.length <= 20) return token;
    return '${token.substring(0, 20)}...';
  }
}
