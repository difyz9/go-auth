package goauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
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
// 使用 crypto/rand 生成密码学安全的随机数
func GenerateNonce(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	
	// 使用 crypto/rand 生成随机字节
	if _, err := rand.Read(b); err != nil {
		// 如果随机数生成失败，使用时间戳作为后备方案
		// 这种情况极少发生，但为了保证系统可用性，提供降级方案
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	
	// 将随机字节映射到字符集
	for i := 0; i < length; i++ {
		b[i] = charset[b[i]%byte(len(charset))]
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
func BuildSignParams(appID string, body interface{}) (map[string]string, error) {
	params := map[string]string{
		"appId":     appID,
		"timestamp": strconv.FormatInt(timeNow().Unix(), 10),
		"nonce":     GenerateNonce(16),
	}
	
	return params, nil
}

// QuickSign 快速生成签名（便捷方法）
func QuickSign(appID, appSecret string, body interface{}) (params map[string]string, sign string, err error) {
	params, err = BuildSignParams(appID, body)
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
