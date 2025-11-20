package goauth

import (
	"testing"
)

// TestSignIncludeBodyConfig 测试签名包含请求体的配置
func TestSignIncludeBodyConfig(t *testing.T) {
	// 测试默认配置
	config := NewConfig()
	if config.SignIncludeBody != false {
		t.Errorf("默认配置应该是 false, 实际是 %v", config.SignIncludeBody)
	}
	
	// 测试使用配置选项
	config2 := NewConfig(WithConfigSignIncludeBody(true))
	if config2.SignIncludeBody != true {
		t.Errorf("配置选项应该生效, 期望 true, 实际是 %v", config2.SignIncludeBody)
	}
	
	// 测试应用级别配置
	includeBody := true
	app := &AppConfig{
		AppID:           "test-app",
		AppSecret:       "test-secret",
		SignIncludeBody: &includeBody,
	}
	
	if app.SignIncludeBody == nil || *app.SignIncludeBody != true {
		t.Error("应用级别配置应该生效")
	}
	
	t.Log("✅ SignIncludeBody 配置测试通过")
}

// TestBuildSignParamsWithBody 测试构建签名参数（包含请求体）
func TestBuildSignParamsWithBody(t *testing.T) {
	appID := "test-app"
	body := map[string]interface{}{
		"key": "value",
	}
	
	// 测试不包含请求体（默认）
	params1, err := BuildSignParams(appID, body, false)
	if err != nil {
		t.Fatalf("BuildSignParams 失败: %v", err)
	}
	
	if _, exists := params1["requestBody"]; exists {
		t.Error("默认不应该包含 requestBody 参数")
	}
	
	if params1["appId"] != appID {
		t.Error("应该包含 appId 参数")
	}
	
	// 测试包含请求体
	params2, err := BuildSignParams(appID, body, true)
	if err != nil {
		t.Fatalf("BuildSignParams 失败: %v", err)
	}
	
	if _, exists := params2["requestBody"]; !exists {
		t.Error("应该包含 requestBody 参数")
	}
	
	t.Log("✅ BuildSignParams 测试通过")
	t.Logf("  不包含body参数数量: %d", len(params1))
	t.Logf("  包含body参数数量: %d", len(params2))
}

// TestQuickSignWithBody 测试快速签名（包含请求体）
func TestQuickSignWithBody(t *testing.T) {
	appID := "test-app"
	appSecret := "test-secret"
	body := map[string]interface{}{
		"text": "Hello",
	}
	
	// 测试不包含请求体
	params1, sign1, err := QuickSign(appID, appSecret, body, false)
	if err != nil {
		t.Fatalf("QuickSign 失败: %v", err)
	}
	
	// 测试包含请求体
	params2, sign2, err := QuickSign(appID, appSecret, body, true)
	if err != nil {
		t.Fatalf("QuickSign 失败: %v", err)
	}
	
	// 两种方式生成的签名应该不同
	if sign1 == sign2 {
		t.Error("包含和不包含请求体的签名应该不同")
	}
	
	// 验证签名
	if !VerifySign(params1, appSecret, sign1) {
		t.Error("签名1验证失败")
	}
	
	if !VerifySign(params2, appSecret, sign2) {
		t.Error("签名2验证失败")
	}
	
	t.Log("✅ QuickSign 测试通过")
	t.Logf("  不包含body签名: %s", sign1[:16]+"...")
	t.Logf("  包含body签名: %s", sign2[:16]+"...")
}

// TestClientSignIncludeBody 测试客户端签名配置
func TestClientSignIncludeBody(t *testing.T) {
	// 测试默认配置（不包含请求体）
	client1 := NewClient("http://localhost:8089", "test-app", "test-secret")
	if client1.SignIncludeBody != false {
		t.Errorf("客户端默认应该不包含请求体, 实际是 %v", client1.SignIncludeBody)
	}
	
	// 测试启用包含请求体
	client2 := NewClient("http://localhost:8089", "test-app", "test-secret",
		WithSignIncludeBody(true))
	if client2.SignIncludeBody != true {
		t.Errorf("客户端配置应该生效, 期望 true, 实际是 %v", client2.SignIncludeBody)
	}
	
	t.Log("✅ Client SignIncludeBody 配置测试通过")
}

// BenchmarkSignWithBody 性能测试：包含请求体
func BenchmarkSignWithBody(b *testing.B) {
	appID := "test-app"
	appSecret := "test-secret"
	body := map[string]interface{}{
		"text":   "Hello, World!",
		"voice":  "zh-CN",
		"speed":  1.0,
		"volume": 50,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = QuickSign(appID, appSecret, body, true)
	}
}

// BenchmarkSignWithoutBody 性能测试：不包含请求体
func BenchmarkSignWithoutBody(b *testing.B) {
	appID := "test-app"
	appSecret := "test-secret"
	body := map[string]interface{}{
		"text":   "Hello, World!",
		"voice":  "zh-CN",
		"speed":  1.0,
		"volume": 50,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = QuickSign(appID, appSecret, body, false)
	}
}

// TestSignConsistency 测试签名一致性
func TestSignConsistency(t *testing.T) {
	appID := "test-app"
	appSecret := "test-secret"
	body := map[string]interface{}{
		"text": "Hello",
	}
	
	// 多次生成签名（相同的timestamp和nonce应该生成相同的签名）
	// 注意：由于timestamp和nonce每次都不同，这里只测试函数能正常工作
	_, sign1, err1 := QuickSign(appID, appSecret, body, false)
	_, sign2, err2 := QuickSign(appID, appSecret, body, false)
	
	if err1 != nil || err2 != nil {
		t.Fatal("QuickSign 失败")
	}
	
	// 由于timestamp和nonce每次都不同，签名应该不同
	if sign1 == sign2 {
		t.Error("不同的timestamp/nonce应该产生不同的签名")
	}
	
	t.Log("✅ 签名一致性测试通过")
}
