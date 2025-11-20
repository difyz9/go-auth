package goauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GenerateSign 生成API签名
// 签名算法：HMAC-SHA256
// 签名步骤：
// 1. 将所有参数（除sign外）按参数名ASCII码从小到大排序
// 2. 使用URL键值对的格式（即key1=value1&key2=value2…）拼接成字符串
// 3. 使用应用的AppSecret对上述字符串进行HMAC-SHA256签名
func GenerateSign(params map[string]string, appSecret string) string {
	// 创建参数副本，避免修改原始map
	paramsCopy := make(map[string]string)
	for k, v := range params {
		if k != "sign" && v != "" { // 过滤掉空值和sign参数
			paramsCopy[k] = v
		}
	}

	// 参数排序
	var keys []string
	for k := range paramsCopy {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接字符串 - 不进行URL编码，保持原始值
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, paramsCopy[k]))
	}
	signString := strings.Join(parts, "&")

	// HMAC-SHA256签名
	h := hmac.New(sha256.New, []byte(appSecret))
	h.Write([]byte(signString))
	return hex.EncodeToString(h.Sum(nil)) // 返回小写十六进制
}

// VerifySign 验证API签名
func VerifySign(params map[string]string, appSecret string, receivedSign string) bool {
	expectedSign := GenerateSign(params, appSecret)
	// 支持大小写不敏感的比较
	return strings.EqualFold(expectedSign, receivedSign)
}

// ValidateTimestamp 验证时间戳
func ValidateTimestamp(timestamp string, tolerance int64) bool {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	now := time.Now().Unix()
	diff := now - ts
	if diff < 0 {
		diff = -diff
	}
	
	return diff <= tolerance
}

// GenerateNonce 生成随机字符串
func GenerateNonce(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// SignatureExample 生成API签名示例
func SignatureExample(appID, appSecret string, additionalParams map[string]string) map[string]string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := GenerateNonce(16)

	params := map[string]string{
		"appId":     appID,
		"timestamp": timestamp,
		"nonce":     nonce,
	}

	// 添加额外参数
	for k, v := range additionalParams {
		params[k] = v
	}

	sign := GenerateSign(params, appSecret)
	params["sign"] = sign

	return params
}

// 用于测试的时间函数（可以被mock）
var timeNow = time.Now

// BuildSignParams 构建签名参数（便捷方法）
// includeBody: 是否将请求体包含在签名中（默认false，推荐）
func BuildSignParams(appID string, body interface{}, includeBody ...bool) (map[string]string, error) {
	params := map[string]string{
		"appId":     appID,
		"timestamp": strconv.FormatInt(timeNow().Unix(), 10),
		"nonce":     GenerateNonce(16),
	}
	
	// 确定是否包含请求体
	shouldInclude := false
	if len(includeBody) > 0 {
		shouldInclude = includeBody[0]
	}
	
	// 如果配置启用且有请求体，序列化并加入参数
	if shouldInclude && body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		if len(bodyBytes) > 0 {
			params["requestBody"] = string(bodyBytes)
		}
	}
	
	return params, nil
}

// QuickSign 快速生成签名（便捷方法）
// includeBody: 是否将请求体包含在签名中（默认false，推荐）
func QuickSign(appID, appSecret string, body interface{}, includeBody ...bool) (params map[string]string, sign string, err error) {
	params, err = BuildSignParams(appID, body, includeBody...)
	if err != nil {
		return nil, "", err
	}
	
	sign = GenerateSign(params, appSecret)
	return params, sign, nil
}

// VerifySignWithDebug 验证签名并输出调试信息
func VerifySignWithDebug(params map[string]string, appSecret string, receivedSign string) (bool, string) {
	expectedSign := GenerateSign(params, appSecret)
	
	debugInfo := fmt.Sprintf(
		"预期签名: %s\n收到签名: %s\n签名参数: %v",
		expectedSign,
		receivedSign,
		params,
	)
	
	return strings.EqualFold(expectedSign, receivedSign), debugInfo
}
