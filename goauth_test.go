package goauth

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGenerateSign(t *testing.T) {
	params := map[string]string{
		"appId":     "test-app",
		"timestamp": "1700000000",
		"nonce":     "abc123",
	}
	appSecret := "test-secret"

	sign := GenerateSign(params, appSecret)
	
	if sign == "" {
		t.Error("签名不能为空")
	}
	
	// 验证签名长度（SHA256 的十六进制表示为64位）
	if len(sign) != 64 {
		t.Errorf("签名长度错误，期望64，实际%d", len(sign))
	}
}

func TestVerifySign(t *testing.T) {
	params := map[string]string{
		"appId":     "test-app",
		"timestamp": "1700000000",
		"nonce":     "abc123",
	}
	appSecret := "test-secret"

	sign := GenerateSign(params, appSecret)
	
	// 测试正确的签名
	if !VerifySign(params, appSecret, sign) {
		t.Error("签名验证失败")
	}
	
	// 测试错误的签名
	if VerifySign(params, appSecret, "wrong-sign") {
		t.Error("错误的签名应该验证失败")
	}
	
	// 测试大小写不敏感
	upperSign := strings.ToUpper(sign)
	if !VerifySign(params, appSecret, upperSign) {
		t.Error("签名应该大小写不敏感")
	}
}

func TestValidateTimestamp(t *testing.T) {
	tolerance := int64(300)
	
	// 测试当前时间
	now := time.Now().Unix()
	if !ValidateTimestamp(strconv.FormatInt(now, 10), tolerance) {
		t.Error("当前时间戳应该有效")
	}
	
	// 测试过期时间
	past := now - 400
	if ValidateTimestamp(strconv.FormatInt(past, 10), tolerance) {
		t.Error("过期的时间戳应该无效")
	}
	
	// 测试未来时间
	future := now + 400
	if ValidateTimestamp(strconv.FormatInt(future, 10), tolerance) {
		t.Error("未来的时间戳应该无效")
	}
	
	// 测试边界值
	boundary := now - 299
	if !ValidateTimestamp(strconv.FormatInt(boundary, 10), tolerance) {
		t.Error("边界值时间戳应该有效")
	}
	
	// 测试无效格式
	if ValidateTimestamp("invalid", tolerance) {
		t.Error("无效格式的时间戳应该无效")
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce1 := GenerateNonce(16)
	nonce2 := GenerateNonce(16)
	
	if len(nonce1) != 16 {
		t.Errorf("Nonce长度错误，期望16，实际%d", len(nonce1))
	}
	
	if nonce1 == nonce2 {
		t.Error("生成的Nonce不应该相同")
	}
}

func TestConfig(t *testing.T) {
	config := NewConfig()
	
	// 测试添加应用
	app := &AppConfig{
		AppID:     "test-app",
		AppSecret: "test-secret",
		Enabled:   true,
	}
	config.AddApp(app)
	
	// 测试获取应用
	retrieved, exists := config.GetApp("test-app")
	if !exists {
		t.Error("应该能获取到添加的应用")
	}
	if retrieved.AppID != "test-app" {
		t.Error("获取的应用ID不正确")
	}
	
	// 测试删除应用
	config.RemoveApp("test-app")
	_, exists = config.GetApp("test-app")
	if exists {
		t.Error("删除后不应该还能获取到应用")
	}
}

func TestConfigLoadSave(t *testing.T) {
	config := NewConfig()
	config.AddApp(&AppConfig{
		AppID:       "test-app",
		AppSecret:   "test-secret",
		AppName:     "测试应用",
		RequireSign: true,
		Enabled:     true,
		RateLimit:   100,
		IPWhitelist: []string{"*"},
	})
	
	// 测试保存到JSON
	err := config.SaveToJSON("/tmp/goauth_test.json")
	if err != nil {
		t.Errorf("保存JSON失败: %v", err)
	}
	
	// 测试从JSON加载
	newConfig := NewConfig()
	err = newConfig.LoadFromJSON("/tmp/goauth_test.json")
	if err != nil {
		t.Errorf("加载JSON失败: %v", err)
	}
	
	app, exists := newConfig.GetApp("test-app")
	if !exists {
		t.Error("从文件加载后应该能获取到应用")
	}
	if app.AppSecret != "test-secret" {
		t.Error("加载的应用信息不正确")
	}
}

func BenchmarkGenerateSign(b *testing.B) {
	params := map[string]string{
		"appId":     "test-app",
		"timestamp": "1700000000",
		"nonce":     "abc123",
	}
	appSecret := "test-secret"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateSign(params, appSecret)
	}
}

func BenchmarkVerifySign(b *testing.B) {
	params := map[string]string{
		"appId":     "test-app",
		"timestamp": "1700000000",
		"nonce":     "abc123",
	}
	appSecret := "test-secret"
	sign := GenerateSign(params, appSecret)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifySign(params, appSecret, sign)
	}
}
