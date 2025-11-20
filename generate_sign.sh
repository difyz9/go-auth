#!/bin/bash

# GoAuth 签名生成工具
# 用法: ./generate_sign.sh <appId> <appSecret> [timestamp] [nonce] [additional_params...]

if [ $# -lt 2 ]; then
    echo "用法: $0 <appId> <appSecret> [timestamp] [nonce] [additional_params...]"
    echo ""
    echo "示例:"
    echo "  $0 test-app-001 test-secret"
    echo "  $0 test-app-001 test-secret 1700000000 abc123"
    echo "  $0 test-app-001 test-secret 1700000000 abc123 user_id=123 amount=100"
    exit 1
fi

APP_ID="$1"
APP_SECRET="$2"
TIMESTAMP="${3:-$(date +%s)}"
NONCE="${4:-$(openssl rand -hex 8)}"

# 收集参数
declare -A params
params["appId"]="$APP_ID"
params["timestamp"]="$TIMESTAMP"
params["nonce"]="$NONCE"

# 添加额外参数
shift 4 2>/dev/null || shift $#
for param in "$@"; do
    key=$(echo "$param" | cut -d= -f1)
    value=$(echo "$param" | cut -d= -f2-)
    params["$key"]="$value"
done

# 排序并拼接参数
sign_string=""
for key in $(echo "${!params[@]}" | tr ' ' '\n' | sort); do
    if [ -n "$sign_string" ]; then
        sign_string="${sign_string}&"
    fi
    sign_string="${sign_string}${key}=${params[$key]}"
done

# 生成签名
sign=$(echo -n "$sign_string" | openssl dgst -sha256 -hmac "$APP_SECRET" | cut -d' ' -f2)

# 输出结果
echo "=== GoAuth 签名信息 ==="
echo "App ID:     $APP_ID"
echo "Timestamp:  $TIMESTAMP"
echo "Nonce:      $NONCE"
echo "Sign String: $sign_string"
echo "Signature:  $sign"
echo ""
echo "=== cURL 命令示例 ==="
echo "curl -X GET 'http://localhost:8080/api/hello' \\"
echo "  -H 'X-App-Id: $APP_ID' \\"
echo "  -H 'X-Timestamp: $TIMESTAMP' \\"
echo "  -H 'X-Nonce: $NONCE' \\"
echo "  -H 'X-Sign: $sign'"
