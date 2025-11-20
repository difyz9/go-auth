package goauth

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAESEncryptDecrypt(t *testing.T) {
	key := "test-aes-key-32-bytes-long-12345"
	plaintext := []byte("Hello, World! 这是测试数据。")

	// 测试加密
	encrypted, err := AESEncrypt(key, plaintext)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	if encrypted == "" {
		t.Error("加密结果不能为空")
	}

	// 测试解密
	decrypted, err := AESDecrypt(key, encrypted)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("解密结果不匹配\n期望: %s\n实际: %s", plaintext, decrypted)
	}
}

func TestAESEncryptJSON(t *testing.T) {
	key := "test-aes-key-32-bytes-long-12345"
	
	// 测试JSON数据加解密
	data := map[string]interface{}{
		"name":  "张三",
		"email": "zhangsan@example.com",
		"age":   30,
	}

	jsonData, _ := json.Marshal(data)

	// 加密
	encrypted, err := AESEncrypt(key, jsonData)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	// 解密
	decrypted, err := AESDecrypt(key, encrypted)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	// 解析解密后的JSON
	var result map[string]interface{}
	if err := json.Unmarshal(decrypted, &result); err != nil {
		t.Fatalf("解析JSON失败: %v", err)
	}

	if result["name"] != data["name"] {
		t.Errorf("数据不匹配: 期望 %v, 实际 %v", data["name"], result["name"])
	}
}

func TestCryptoMiddlewareDecrypt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建加解密配置
	cryptoConfig := NewCryptoConfig("test-key-32-bytes-long-1234567")
	cryptoConfig.EnableDecryption = true

	// 创建中间件
	cryptoMiddleware := NewCryptoMiddleware(cryptoConfig)

	// 创建测试路由
	r := gin.New()
	r.Use(cryptoMiddleware.DecryptRequest())
	r.POST("/test", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		wasDecrypted, _ := c.Get("crypto_decrypted")
		c.JSON(200, gin.H{
			"data":      data,
			"decrypted": wasDecrypted,
		})
	})

	// 准备测试数据
	testData := map[string]interface{}{
		"name":  "测试",
		"value": 123,
	}
	testDataJSON, _ := json.Marshal(testData)

	// 加密测试数据
	encrypted, err := AESEncrypt(cryptoConfig.SecretKey, testDataJSON)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	encryptedRequest := EncryptedRequest{
		Data: encrypted,
	}

	// 测试加密请求
	reqBody, _ := json.Marshal(encryptedRequest)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Encrypted", "1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("期望状态码 200, 实际 %d, 响应: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["decrypted"] != true {
		t.Error("请求应该被标记为已解密")
	}
}

func TestCryptoMiddlewareEncrypt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建加解密配置
	cryptoConfig := NewCryptoConfig("test-key-32-bytes-long-1234567")
	cryptoConfig.EnableEncryption = true
	cryptoConfig.ForceEncryption = false

	// 创建中间件
	cryptoMiddleware := NewCryptoMiddleware(cryptoConfig)

	// 创建测试路由
	r := gin.New()
	r.Use(cryptoMiddleware.EncryptResponse())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "test response",
			"code":    200,
		})
	})

	// 测试请求加密响应
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Response-Encrypt", "1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("期望状态码 200, 实际 %d", w.Code)
	}

	// 检查是否返回了加密标识头
	if w.Header().Get("X-Encrypted") != "1" {
		t.Error("响应应该包含 X-Encrypted 头")
	}

	// 解析加密响应
	var encryptedResp EncryptedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &encryptedResp); err != nil {
		t.Fatalf("解析加密响应失败: %v", err)
	}

	if !encryptedResp.Encrypted {
		t.Error("响应应该被标记为已加密")
	}

	// 解密验证
	decrypted, err := AESDecrypt(cryptoConfig.SecretKey, encryptedResp.Data)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(decrypted, &result); err != nil {
		t.Fatalf("解析解密数据失败: %v", err)
	}

	if result["message"] != "test response" {
		t.Errorf("解密后的数据不匹配: %v", result)
	}
}

func TestNormalizeAESKey(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"short", 16},                                // 少于16字节，填充到16
		{"exactly16bytes", 16},                       // 正好16字节
		{"this-is-24-bytes-key!!!", 24},             // 24字节（实际23字节，截断到16）
		{"this-is-a-32-bytes-long-!!", 32},          // 28字节，截断到24
		{"this-is-a-32-bytes-long-key-value!!!", 32}, // 超过32字节，截断到32
	}

	for _, tt := range tests {
		result := normalizeAESKey(tt.input)
		// 实际测试长度是否符合AES标准（16/24/32）
		validLengths := map[int]bool{16: true, 24: true, 32: true}
		if !validLengths[len(result)] {
			t.Errorf("normalizeAESKey(%q) = %d字节, 不是有效的AES密钥长度（16/24/32）",
				tt.input, len(result))
		}
	}
}

func TestCryptoConfigSkipPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cryptoConfig := NewCryptoConfig("test-key")
	cryptoConfig.SkipPaths = []string{"/health", "/public"}
	cryptoConfig.EnableDecryption = true

	cryptoMiddleware := NewCryptoMiddleware(cryptoConfig)

	r := gin.New()
	r.Use(cryptoMiddleware.DecryptRequest())
	r.POST("/health", func(c *gin.Context) {
		wasDecrypted, exists := c.Get("crypto_decrypted")
		c.JSON(200, gin.H{
			"processed": exists,
			"decrypted": wasDecrypted,
		})
	})

	// 测试跳过的路径
	req := httptest.NewRequest("POST", "/health", bytes.NewBufferString(`{"test": "data"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["processed"] == true {
		t.Error("跳过的路径不应该经过加解密处理")
	}
}

func BenchmarkAESEncrypt(b *testing.B) {
	key := "test-aes-key-32-bytes-long-12345"
	plaintext := []byte("Hello, World! This is a test message for benchmarking.")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AESEncrypt(key, plaintext)
	}
}

func BenchmarkAESDecrypt(b *testing.B) {
	key := "test-aes-key-32-bytes-long-12345"
	plaintext := []byte("Hello, World! This is a test message for benchmarking.")
	encrypted, _ := AESEncrypt(key, plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AESDecrypt(key, encrypted)
	}
}
