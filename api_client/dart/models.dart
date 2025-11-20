import 'package:json_annotation/json_annotation.dart';

part 'models.g.dart';

/// 用户登录请求
@JsonSerializable()
class LoginRequest {
  final String email;
  final String password;

  const LoginRequest({
    required this.email,
    required this.password,
  });

  factory LoginRequest.fromJson(Map<String, dynamic> json) =>
      _$LoginRequestFromJson(json);

  Map<String, dynamic> toJson() => _$LoginRequestToJson(this);
}

/// 用户登录响应
@JsonSerializable()
class LoginResponse {
  final String accessToken;
  final String? refreshToken;
  final UserInfo user;

  const LoginResponse({
    required this.accessToken,
    this.refreshToken,
    required this.user,
  });

  factory LoginResponse.fromJson(Map<String, dynamic> json) =>
      _$LoginResponseFromJson(json);

  Map<String, dynamic> toJson() => _$LoginResponseToJson(this);
}

/// 用户信息
@JsonSerializable()
class UserInfo {
  final String id;
  final String username;
  final String email;
  final String? nickname;
  final String? avatar;
  final bool isVip;
  final bool isActive;
  @JsonKey(name: 'created_at')
  final String createdAt;
  @JsonKey(name: 'updated_at')
  final String updatedAt;

  const UserInfo({
    required this.id,
    required this.username,
    required this.email,
    this.nickname,
    this.avatar,
    required this.isVip,
    required this.isActive,
    required this.createdAt,
    required this.updatedAt,
  });

  factory UserInfo.fromJson(Map<String, dynamic> json) =>
      _$UserInfoFromJson(json);

  Map<String, dynamic> toJson() => _$UserInfoToJson(this);
}

/// 支付订单
@JsonSerializable()
class PaymentOrder {
  final int id;
  final String appId;
  final String userId;
  final String orderNo;
  final String? tradeNo;
  final String subject;
  final double amount;
  final int status;
  final String payWay;
  final String orderType;
  final String? extra;
  final int? payTime;
  @JsonKey(name: 'created_at')
  final String createdAt;
  @JsonKey(name: 'updated_at')
  final String updatedAt;

  const PaymentOrder({
    required this.id,
    required this.appId,
    required this.userId,
    required this.orderNo,
    this.tradeNo,
    required this.subject,
    required this.amount,
    required this.status,
    required this.payWay,
    required this.orderType,
    this.extra,
    this.payTime,
    required this.createdAt,
    required this.updatedAt,
  });

  factory PaymentOrder.fromJson(Map<String, dynamic> json) =>
      _$PaymentOrderFromJson(json);

  Map<String, dynamic> toJson() => _$PaymentOrderToJson(this);
}

/// 创建支付订单请求
@JsonSerializable()
class CreatePaymentRequest {
  final double amount;
  final String subject;
  final String orderNo;
  final String payWay;
  final String? notifyUrl;
  final String? returnUrl;
  final String? extra;

  const CreatePaymentRequest({
    required this.amount,
    required this.subject,
    required this.orderNo,
    required this.payWay,
    this.notifyUrl,
    this.returnUrl,
    this.extra,
  });

  factory CreatePaymentRequest.fromJson(Map<String, dynamic> json) =>
      _$CreatePaymentRequestFromJson(json);

  Map<String, dynamic> toJson() => _$CreatePaymentRequestToJson(this);
}

/// 支付响应
@JsonSerializable()
class PaymentResponse {
  final String orderId;
  final String? payUrl;
  final String? qrCode;
  final String? formData;

  const PaymentResponse({
    required this.orderId,
    this.payUrl,
    this.qrCode,
    this.formData,
  });

  factory PaymentResponse.fromJson(Map<String, dynamic> json) =>
      _$PaymentResponseFromJson(json);

  Map<String, dynamic> toJson() => _$PaymentResponseToJson(this);
}

/// 分页响应
@JsonSerializable(genericArgumentFactories: true)
class PaginatedResponse<T> {
  final List<T>? list;
  final int total;
  final int page;
  final int pageSize;

  const PaginatedResponse({
    this.list,
    required this.total,
    required this.page,
    required this.pageSize,
  });

  factory PaginatedResponse.fromJson(
    Map<String, dynamic> json,
    T Function(Object? json) fromJsonT,
  ) =>
      _$PaginatedResponseFromJson(json, fromJsonT);

  Map<String, dynamic> toJson(Object Function(T value) toJsonT) =>
      _$PaginatedResponseToJson(this, toJsonT);
}

/// API应用信息
@JsonSerializable()
class ApiApp {
  final int id;
  final String appId;
  final String? appSecret;
  final String appName;
  final int status;
  final String description;
  final String callbackUrl;
  final String ipWhitelist;
  final int rateLimit;
  final bool? requireSign;
  @JsonKey(name: 'created_at')
  final String createdAt;
  @JsonKey(name: 'updated_at')
  final String updatedAt;

  const ApiApp({
    required this.id,
    required this.appId,
    this.appSecret,
    required this.appName,
    required this.status,
    required this.description,
    required this.callbackUrl,
    required this.ipWhitelist,
    required this.rateLimit,
    this.requireSign,
    required this.createdAt,
    required this.updatedAt,
  });

  factory ApiApp.fromJson(Map<String, dynamic> json) =>
      _$ApiAppFromJson(json);

  Map<String, dynamic> toJson() => _$ApiAppToJson(this);
}
