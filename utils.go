package goauth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

// GenerateSecureNonce 生成加密安全的随机字符串
func GenerateSecureNonce(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// GenerateAppSecret 生成应用密钥
func GenerateAppSecret(length int) (string, error) {
	if length < 16 {
		return "", fmt.Errorf("密钥长度不能少于16位")
	}
	
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	// 使用base64编码并去除特殊字符
	secret := base64.URLEncoding.EncodeToString(bytes)
	secret = strings.ReplaceAll(secret, "-", "a")
	secret = strings.ReplaceAll(secret, "_", "b")
	
	if len(secret) > length {
		secret = secret[:length]
	}
	
	return secret, nil
}

// ValidateAppID 验证应用ID格式
func ValidateAppID(appID string) error {
	if appID == "" {
		return fmt.Errorf("应用ID不能为空")
	}
	if len(appID) < 3 {
		return fmt.Errorf("应用ID长度不能少于3位")
	}
	if len(appID) > 64 {
		return fmt.Errorf("应用ID长度不能超过64位")
	}
	
	// 只允许字母、数字、下划线和连字符
	for _, char := range appID {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return fmt.Errorf("应用ID只能包含字母、数字、下划线和连字符")
		}
	}
	
	return nil
}

// SanitizeIP 清理和标准化IP地址
func SanitizeIP(ip string) string {
	ip = strings.TrimSpace(ip)
	
	// 处理IPv6的简写形式
	if ip == "::1" {
		return "127.0.0.1"
	}
	
	// 去除端口号
	if idx := strings.LastIndex(ip, ":"); idx > 0 {
		// 确保不是IPv6地址
		if strings.Count(ip, ":") == 1 {
			ip = ip[:idx]
		}
	}
	
	return ip
}

// MaskSecret 遮蔽密钥用于日志输出
func MaskSecret(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

// BuildRequestID 生成请求ID
func BuildRequestID() string {
	nonce, _ := GenerateSecureNonce(16)
	return fmt.Sprintf("%d-%s", timeNow().UnixNano(), nonce)
}

// QuickConfig 快速创建单应用配置
func QuickConfig(appID, appSecret string) *Config {
	config := NewConfig()
	config.AddApp(&AppConfig{
		AppID:       appID,
		AppSecret:   appSecret,
		RequireSign: true,
		Enabled:     true,
		RateLimit:   1000,
		IPWhitelist: []string{"*"},
	})
	return config
}

// ConfigBuilder 配置构建器
type ConfigBuilder struct {
	config *Config
}

// NewConfigBuilder 创建配置构建器
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: NewConfig(),
	}
}

// SetTimestampTolerance 设置时间戳容差
func (b *ConfigBuilder) SetTimestampTolerance(seconds int64) *ConfigBuilder {
	b.config.TimestampTolerance = seconds
	return b
}

// SetDefaultRateLimit 设置默认速率限制
func (b *ConfigBuilder) SetDefaultRateLimit(limit int) *ConfigBuilder {
	b.config.DefaultRateLimit = limit
	return b
}

// SetEnableIPCheck 设置是否启用IP检查
func (b *ConfigBuilder) SetEnableIPCheck(enabled bool) *ConfigBuilder {
	b.config.EnableIPCheck = enabled
	return b
}

// AddApp 添加应用
func (b *ConfigBuilder) AddApp(app *AppConfig) *ConfigBuilder {
	b.config.AddApp(app)
	return b
}

// AddSimpleApp 添加简单应用配置
func (b *ConfigBuilder) AddSimpleApp(appID, appSecret, appName string) *ConfigBuilder {
	b.config.AddApp(&AppConfig{
		AppID:       appID,
		AppSecret:   appSecret,
		AppName:     appName,
		RequireSign: true,
		Enabled:     true,
		IPWhitelist: []string{"*"},
	})
	return b
}

// Build 构建配置
func (b *ConfigBuilder) Build() (*Config, error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}

// MustBuild 构建配置（失败则panic）
func (b *ConfigBuilder) MustBuild() *Config {
	config, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("配置构建失败: %v", err))
	}
	return config
}
