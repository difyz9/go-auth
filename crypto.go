package goauth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CryptoConfig 加解密配置
type CryptoConfig struct {
	Enabled           bool     `json:"enabled" yaml:"enabled"`                       // 是否启用加解密
	SecretKey         string   `json:"secret_key" yaml:"secret_key"`                 // AES密钥
	EnableDecryption  bool     `json:"enable_decryption" yaml:"enable_decryption"`   // 是否启用请求解密
	EnableEncryption  bool     `json:"enable_encryption" yaml:"enable_encryption"`   // 是否启用响应加密
	SkipPaths         []string `json:"skip_paths" yaml:"skip_paths"`                 // 跳过加解密的路径
	ForceEncryption   bool     `json:"force_encryption" yaml:"force_encryption"`     // 强制加密（不需要客户端头）
	EncryptionHeader  string   `json:"encryption_header" yaml:"encryption_header"`   // 加密标识头（默认X-Encrypted）
	ResponseEncHeader string   `json:"response_enc_header" yaml:"response_enc_header"` // 响应加密请求头（默认X-Response-Encrypt）
}

// NewCryptoConfig 创建默认加解密配置
func NewCryptoConfig(secretKey string) *CryptoConfig {
	return &CryptoConfig{
		Enabled:           true,
		SecretKey:         normalizeAESKey(secretKey),
		EnableDecryption:  true,
		EnableEncryption:  true,
		SkipPaths:         []string{},
		ForceEncryption:   false,
		EncryptionHeader:  "X-Encrypted",
		ResponseEncHeader: "X-Response-Encrypt",
	}
}

// EncryptedRequest 加密请求结构
type EncryptedRequest struct {
	Data      string `json:"data"`                // 加密后的数据（Base64）
	Timestamp int64  `json:"timestamp,omitempty"` // 时间戳
}

// EncryptedResponse 加密响应结构
type EncryptedResponse struct {
	Encrypted bool   `json:"encrypted"` // 是否加密
	Data      string `json:"data"`      // 加密后的数据（Base64）
	Timestamp int64  `json:"timestamp"` // 时间戳
}

// CryptoMiddleware 加解密中间件
type CryptoMiddleware struct {
	config *CryptoConfig
	logger Logger
}

// NewCryptoMiddleware 创建加解密中间件
func NewCryptoMiddleware(config *CryptoConfig, logger ...Logger) *CryptoMiddleware {
	var log Logger
	if len(logger) > 0 && logger[0] != nil {
		log = logger[0]
	} else {
		log = &defaultLogger{}
	}

	return &CryptoMiddleware{
		config: config,
		logger: log,
	}
}

// DecryptRequest 请求体解密中间件
func (m *CryptoMiddleware) DecryptRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用解密
		if !m.config.Enabled || !m.config.EnableDecryption {
			c.Next()
			return
		}

		// 检查是否跳过该路径
		if m.shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 只处理有请求体的方法
		if c.Request.Method != http.MethodPost && 
		   c.Request.Method != http.MethodPut && 
		   c.Request.Method != http.MethodPatch {
			c.Next()
			return
		}

		// 检查Content-Type
		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			c.Next()
			return
		}

		// 读取原始请求体
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			m.logger.Error("读取请求体失败", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
			c.Abort()
			return
		}

		// 如果请求体为空，直接继续
		if len(bodyBytes) == 0 {
			c.Next()
			return
		}

		// 检查是否有加密标识头
		isEncrypted := c.GetHeader(m.config.EncryptionHeader) == "1" || 
		               c.GetHeader(m.config.EncryptionHeader) == "true"

		var decryptedData []byte
		if isEncrypted {
			// 解析加密的JSON结构
			var encryptedReq EncryptedRequest
			if err := json.Unmarshal(bodyBytes, &encryptedReq); err != nil {
				m.logger.Error("加密请求格式错误", "error", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "加密请求格式错误"})
				c.Abort()
				return
			}

			// AES解密
			decryptedData, err = AESDecrypt(m.config.SecretKey, encryptedReq.Data)
			if err != nil {
				m.logger.Error("请求体解密失败", "error", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "请求体解密失败"})
				c.Abort()
				return
			}

			// 验证解密后的数据是否为有效JSON
			if !json.Valid(decryptedData) {
				m.logger.Error("解密后的数据不是有效的JSON格式")
				c.JSON(http.StatusBadRequest, gin.H{"error": "解密后的数据格式错误"})
				c.Abort()
				return
			}

			m.logger.Debug("请求体解密成功", "原始长度", len(encryptedReq.Data), "解密后长度", len(decryptedData))
		} else {
			// 未加密的请求，直接使用原始数据
			decryptedData = bodyBytes
		}

		// 将解密后的数据重新设置为请求体
		c.Request.Body = io.NopCloser(bytes.NewBuffer(decryptedData))
		c.Request.ContentLength = int64(len(decryptedData))

		// 在上下文中标记解密信息
		c.Set("crypto_decrypted", isEncrypted)
		c.Set("crypto_original_size", len(bodyBytes))
		c.Set("crypto_decrypted_size", len(decryptedData))

		c.Next()
	}
}

// EncryptResponse 响应体加密中间件
func (m *CryptoMiddleware) EncryptResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用加密
		if !m.config.Enabled || !m.config.EnableEncryption {
			c.Next()
			return
		}

		// 检查是否跳过该路径
		if m.shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 检查是否需要加密响应
		needEncrypt := m.config.ForceEncryption || 
		               c.GetHeader(m.config.ResponseEncHeader) == "1" || 
		               c.GetHeader(m.config.ResponseEncHeader) == "true"

		if !needEncrypt {
			c.Next()
			return
		}

		// 创建自定义ResponseWriter来捕获响应数据
		writer := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		// 执行后续处理
		c.Next()

		// 获取响应数据
		responseData := writer.body.Bytes()

		// 如果有错误或响应为空，不加密
		if writer.Status() != http.StatusOK || len(responseData) == 0 {
			writer.ResponseWriter.Write(responseData)
			return
		}

		// 检查Content-Type是否为JSON
		contentType := c.Writer.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			writer.ResponseWriter.Write(responseData)
			return
		}

		// AES加密响应数据
		encryptedData, err := AESEncrypt(m.config.SecretKey, responseData)
		if err != nil {
			m.logger.Error("响应体加密失败", "error", err)
			writer.ResponseWriter.Write(responseData)
			return
		}

		// 创建加密响应结构
		encryptedResp := EncryptedResponse{
			Encrypted: true,
			Data:      encryptedData,
			Timestamp: time.Now().Unix(),
		}

		// 序列化加密响应
		encryptedJSON, err := json.Marshal(encryptedResp)
		if err != nil {
			m.logger.Error("加密响应序列化失败", "error", err)
			writer.ResponseWriter.Write(responseData)
			return
		}

		// 设置加密标识头
		c.Header(m.config.EncryptionHeader, "1")
		c.Header("Content-Length", fmt.Sprintf("%d", len(encryptedJSON)))

		// 写入加密后的响应
		writer.ResponseWriter.Write(encryptedJSON)

		m.logger.Debug("响应体加密成功", "原始长度", len(responseData), "加密后长度", len(encryptedJSON))
	}
}

// shouldSkipPath 检查是否应该跳过该路径
func (m *CryptoMiddleware) shouldSkipPath(path string) bool {
	for _, skipPath := range m.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// responseBodyWriter 自定义ResponseWriter，用于捕获响应数据
type responseBodyWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *responseBodyWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

func (w *responseBodyWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseBodyWriter) Status() int {
	if w.status != 0 {
		return w.status
	}
	return w.ResponseWriter.Status()
}

// AES加解密函数

// AESEncrypt AES加密（使用CBC模式）
func AESEncrypt(key string, plaintext []byte) (string, error) {
	// 确保密钥长度正确
	keyBytes := []byte(normalizeAESKey(key))

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("创建AES cipher失败: %w", err)
	}

	// PKCS7填充
	plaintext = pkcs7Padding(plaintext, aes.BlockSize)

	// 生成随机IV
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("生成IV失败: %w", err)
	}

	// CBC模式加密
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	// Base64编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt AES解密（使用CBC模式）
func AESDecrypt(key string, ciphertext string) ([]byte, error) {
	// 确保密钥长度正确
	keyBytes := []byte(normalizeAESKey(key))

	// Base64解码
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("Base64解码失败: %w", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("创建AES cipher失败: %w", err)
	}

	if len(ciphertextBytes) < aes.BlockSize {
		return nil, fmt.Errorf("密文长度太短")
	}

	// 提取IV
	iv := ciphertextBytes[:aes.BlockSize]
	ciphertextBytes = ciphertextBytes[aes.BlockSize:]

	if len(ciphertextBytes)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("密文长度不是块大小的倍数")
	}

	// CBC模式解密
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertextBytes, ciphertextBytes)

	// 去除PKCS7填充
	plaintext, err := pkcs7Unpadding(ciphertextBytes)
	if err != nil {
		return nil, fmt.Errorf("去除填充失败: %w", err)
	}

	return plaintext, nil
}

// pkcs7Padding PKCS7填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// pkcs7Unpadding 去除PKCS7填充
func pkcs7Unpadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("数据为空")
	}

	padding := int(data[length-1])
	if padding > length || padding > aes.BlockSize {
		return nil, fmt.Errorf("填充值无效")
	}

	return data[:length-padding], nil
}

// normalizeAESKey 标准化AES密钥长度（16、24或32字节）
func normalizeAESKey(key string) string {
	keyBytes := []byte(key)
	keyLen := len(keyBytes)

	// AES-256（32字节）
	if keyLen >= 32 {
		return string(keyBytes[:32])
	}
	// AES-192（24字节）
	if keyLen >= 24 {
		return string(keyBytes[:24])
	}
	// AES-128（16字节）
	if keyLen >= 16 {
		return string(keyBytes[:16])
	}

	// 如果密钥长度不足16字节，用0填充
	paddedKey := make([]byte, 16)
	copy(paddedKey, keyBytes)
	return string(paddedKey)
}

// WithCrypto 为认证中间件添加加解密功能的便捷方法
func (m *AuthMiddleware) WithCrypto(cryptoConfig *CryptoConfig) *AuthMiddleware {
	if cryptoConfig == nil {
		return m
	}
	// 可以在这里存储crypto配置供后续使用
	return m
}
