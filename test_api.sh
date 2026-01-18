#!/bin/bash

# GoAuth API 测试脚本
# 用法: ./test_api.sh [base_url] [app_id] [app_secret]

BASE_URL="${1:-http://localhost:8080}"
APP_ID="${2:-test-app-001}"
APP_SECRET="${3:-test-secret-key-12345678}"

echo "========================================"
echo "GoAuth API 测试"
echo "========================================"
echo "服务地址: $BASE_URL"
echo "应用ID:   $APP_ID"
echo ""

# 生成认证参数
TIMESTAMP=$(date +%s)
NONCE=$(openssl rand -hex 8)

# 生成签名的函数
generate_sign() {
    local params="$1"
    local sign_string=$(echo "$params" | tr '&' '\n' | sort | tr '\n' '&' | sed 's/&$//')
    echo -n "$sign_string" | openssl dgst -sha256 -hmac "$APP_SECRET" | cut -d' ' -f2
}

echo "========================================"
echo "测试 1: GET 请求（无请求体）"
echo "========================================"

# 构建签名参数
params="appId=${APP_ID}&timestamp=${TIMESTAMP}&nonce=${NONCE}"
sign=$(generate_sign "$params")

echo "请求参数:"
echo "  App ID:    $APP_ID"
echo "  Timestamp: $TIMESTAMP"
echo "  Nonce:     $NONCE"
echo "  Sign:      $sign"
echo ""

echo "执行请求..."
response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  -X GET "${BASE_URL}/api/users" \
  -H "X-App-Id: ${APP_ID}" \
  -H "X-Timestamp: ${TIMESTAMP}" \
  -H "X-Nonce: ${NONCE}" \
  -H "X-Sign: ${sign}")

http_code=$(echo "$response" | grep "HTTP_STATUS" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_STATUS/d')

echo "响应状态: $http_code"
echo "响应内容:"
echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"
echo ""

if [ "$http_code" == "200" ]; then
    echo "✅ 测试通过"
else
    echo "❌ 测试失败"
fi

echo ""
echo "========================================"
echo "测试 2: POST 请求（不包含请求体在签名中）"
echo "========================================"

# 重新生成参数（避免 nonce 重复）
TIMESTAMP=$(date +%s)
NONCE=$(openssl rand -hex 8)

# 请求体
REQUEST_BODY='{"user_id":123,"amount":99.99}'

# 构建签名参数（不包含请求体）
params="appId=${APP_ID}&nonce=${NONCE}&timestamp=${TIMESTAMP}"
sign=$(generate_sign "$params")

echo "请求参数:"
echo "  App ID:    $APP_ID"
echo "  Timestamp: $TIMESTAMP"
echo "  Nonce:     $NONCE"
echo "  Body:      $REQUEST_BODY"
echo "  Sign:      $sign"
echo ""

echo "执行请求..."
response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  -X POST "${BASE_URL}/api/orders" \
  -H "Content-Type: application/json" \
  -H "X-App-Id: ${APP_ID}" \
  -H "X-Timestamp: ${TIMESTAMP}" \
  -H "X-Nonce: ${NONCE}" \
  -H "X-Sign: ${sign}" \
  -d "$REQUEST_BODY")

http_code=$(echo "$response" | grep "HTTP_STATUS" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_STATUS/d')

echo "响应状态: $http_code"
echo "响应内容:"
echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"
echo ""

if [ "$http_code" == "200" ]; then
    echo "✅ 测试通过"
else
    echo "❌ 测试失败"
fi

echo ""
echo "========================================"
echo "测试 3: 错误的签名"
echo "========================================"

TIMESTAMP=$(date +%s)
NONCE=$(openssl rand -hex 8)
WRONG_SIGN="wrong-signature-here"

echo "请求参数:"
echo "  App ID:    $APP_ID"
echo "  Timestamp: $TIMESTAMP"
echo "  Nonce:     $NONCE"
echo "  Sign:      $WRONG_SIGN (错误的签名)"
echo ""

echo "执行请求..."
response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  -X GET "${BASE_URL}/api/users" \
  -H "X-App-Id: ${APP_ID}" \
  -H "X-Timestamp: ${TIMESTAMP}" \
  -H "X-Nonce: ${NONCE}" \
  -H "X-Sign: ${WRONG_SIGN}")

http_code=$(echo "$response" | grep "HTTP_STATUS" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_STATUS/d')

echo "响应状态: $http_code"
echo "响应内容:"
echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"
echo ""

if [ "$http_code" == "401" ]; then
    echo "✅ 测试通过（正确拒绝了错误的签名）"
else
    echo "❌ 测试失败（应该返回401）"
fi

echo ""
echo "========================================"
echo "测试 4: 缺少参数"
echo "========================================"

echo "请求参数:"
echo "  App ID:    $APP_ID"
echo "  (缺少其他参数)"
echo ""

echo "执行请求..."
response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  -X GET "${BASE_URL}/api/users" \
  -H "X-App-Id: ${APP_ID}")

http_code=$(echo "$response" | grep "HTTP_STATUS" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_STATUS/d')

echo "响应状态: $http_code"
echo "响应内容:"
echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"
echo ""

if [ "$http_code" == "400" ]; then
    echo "✅ 测试通过（正确拒绝了不完整的参数）"
else
    echo "❌ 测试失败（应该返回400）"
fi

echo ""
echo "========================================"
echo "测试完成"
echo "========================================"
