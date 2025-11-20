package goauth

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// AppConfig 应用配置
type AppConfig struct {
	AppID         string   `json:"app_id" yaml:"app_id"`                   // 应用ID
	AppSecret     string   `json:"app_secret" yaml:"app_secret"`           // 应用密钥
	AppName       string   `json:"app_name" yaml:"app_name"`               // 应用名称
	RequireSign   bool     `json:"require_sign" yaml:"require_sign"`       // 是否需要签名验证
	IPWhitelist   []string `json:"ip_whitelist" yaml:"ip_whitelist"`      // IP白名单
	AllowedRoutes []string `json:"allowed_routes" yaml:"allowed_routes"`  // 允许访问的路由
	RateLimit     int      `json:"rate_limit" yaml:"rate_limit"`          // 速率限制(次/分钟)
	Enabled       bool     `json:"enabled" yaml:"enabled"`                // 是否启用
}

// Config 认证配置管理器
type Config struct {
	Apps              map[string]*AppConfig `json:"apps" yaml:"apps"`                                   // 应用配置映射
	TimestampTolerance int64                `json:"timestamp_tolerance" yaml:"timestamp_tolerance"`     // 时间戳容差(秒)
	DefaultRateLimit  int                   `json:"default_rate_limit" yaml:"default_rate_limit"`       // 默认速率限制
	EnableIPCheck     bool                  `json:"enable_ip_check" yaml:"enable_ip_check"`             // 是否启用IP检查
	mu                sync.RWMutex          `json:"-" yaml:"-"`
}

// ConfigOption 配置选项函数
type ConfigOption func(*Config)

// WithTimestampTolerance 设置时间戳容差
func WithTimestampTolerance(seconds int64) ConfigOption {
	return func(c *Config) {
		c.TimestampTolerance = seconds
	}
}

// WithDefaultRateLimit 设置默认速率限制
func WithDefaultRateLimit(limit int) ConfigOption {
	return func(c *Config) {
		c.DefaultRateLimit = limit
	}
}

// WithIPCheck 设置是否启用IP检查
func WithIPCheck(enabled bool) ConfigOption {
	return func(c *Config) {
		c.EnableIPCheck = enabled
	}
}

// NewConfig 创建新的配置管理器
func NewConfig(opts ...ConfigOption) *Config {
	c := &Config{
		Apps:               make(map[string]*AppConfig),
		TimestampTolerance: 300,  // 默认5分钟
		DefaultRateLimit:   1000,
		EnableIPCheck:      true,
	}
	
	// 应用配置选项
	for _, opt := range opts {
		opt(c)
	}
	
	return c
}

// AddApp 添加应用配置
func (c *Config) AddApp(app *AppConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if app.RateLimit == 0 {
		app.RateLimit = c.DefaultRateLimit
	}
	c.Apps[app.AppID] = app
}

// GetApp 获取应用配置
func (c *Config) GetApp(appID string) (*AppConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	app, exists := c.Apps[appID]
	return app, exists
}

// RemoveApp 移除应用配置
func (c *Config) RemoveApp(appID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.Apps, appID)
}

// LoadFromYAML 从YAML文件加载配置
func (c *Config) LoadFromYAML(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("解析YAML配置失败: %w", err)
	}
	
	return nil
}

// LoadFromJSON 从JSON文件加载配置
func (c *Config) LoadFromJSON(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("解析JSON配置失败: %w", err)
	}
	
	return nil
}

// SaveToYAML 保存配置到YAML文件
func (c *Config) SaveToYAML(filePath string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("序列化YAML失败: %w", err)
	}
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	
	return nil
}

// SaveToJSON 保存配置到JSON文件
func (c *Config) SaveToJSON(filePath string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化JSON失败: %w", err)
	}
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	
	return nil
}

// Validate 验证配置是否有效
func (c *Config) Validate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.TimestampTolerance <= 0 {
		return fmt.Errorf("时间戳容差必须大于0")
	}
	
	if c.DefaultRateLimit <= 0 {
		return fmt.Errorf("默认速率限制必须大于0")
	}
	
	for appID, app := range c.Apps {
		if app.AppID == "" {
			return fmt.Errorf("应用ID不能为空")
		}
		if app.AppID != appID {
			return fmt.Errorf("应用ID不匹配: %s != %s", app.AppID, appID)
		}
		if app.AppSecret == "" {
			return fmt.Errorf("应用[%s]的密钥不能为空", appID)
		}
		if len(app.AppSecret) < 16 {
			return fmt.Errorf("应用[%s]的密钥长度不能少于16位", appID)
		}
		if app.RateLimit < 0 {
			return fmt.Errorf("应用[%s]的速率限制不能为负数", appID)
		}
	}
	
	return nil
}

// GetApps 获取所有应用配置（副本）
func (c *Config) GetApps() map[string]*AppConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	apps := make(map[string]*AppConfig, len(c.Apps))
	for k, v := range c.Apps {
		appCopy := *v
		apps[k] = &appCopy
	}
	return apps
}

// GetEnabledAppCount 获取启用的应用数量
func (c *Config) GetEnabledAppCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	count := 0
	for _, app := range c.Apps {
		if app.Enabled {
			count++
		}
	}
	return count
}
