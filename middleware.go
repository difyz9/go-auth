package goauth

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// 认证头部字段
	HeaderAppID     = "X-App-Id"
	HeaderTimestamp = "X-Timestamp"
	HeaderNonce     = "X-Nonce"
	HeaderSign      = "X-Sign"

	// 上下文键
	ContextKeyApp   = "goauth_app"
	ContextKeyAppID = "goauth_app_id"
)

// AuthMiddleware API认证中间件
type AuthMiddleware struct {
	config      *Config
	rateLimiter *RateLimiter
	logger      Logger
	jwtService  *JWTService // JWT 服务（可选）
}

// Logger 日志接口
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// defaultLogger 默认日志实现
type defaultLogger struct{}

func (l *defaultLogger) Info(msg string, fields ...interface{})  { fmt.Println("INFO:", msg, fields) }
func (l *defaultLogger) Error(msg string, fields ...interface{}) { fmt.Println("ERROR:", msg, fields) }
func (l *defaultLogger) Debug(msg string, fields ...interface{}) { fmt.Println("DEBUG:", msg, fields) }

// Options 中间件选项
type Options struct {
	Config      *Config
	Logger      Logger
	ErrorHandler ErrorHandler
}

// ErrorHandler 错误处理函数
type ErrorHandler func(c *gin.Context, code int, message string, detail string)

// DefaultErrorHandler 默认错误处理
func DefaultErrorHandler(c *gin.Context, code int, message string, detail string) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": message,
		"detail":  detail,
	})
}

// NewAuthMiddleware 创建新的认证中间件
func NewAuthMiddleware(opts Options) *AuthMiddleware {
	if opts.Config == nil {
		opts.Config = NewConfig()
	}
	if opts.Logger == nil {
		opts.Logger = &defaultLogger{}
	}
	
	return &AuthMiddleware{
		config:      opts.Config,
		rateLimiter: NewRateLimiter(),
		logger:      opts.Logger,
	}
}

// New 创建认证中间件的便捷方法
func New(config *Config) *AuthMiddleware {
	return NewAuthMiddleware(Options{Config: config})
}

// WithConfig 设置配置
func (m *AuthMiddleware) WithConfig(config *Config) *AuthMiddleware {
	m.config = config
	return m
}

// WithLogger 设置日志
func (m *AuthMiddleware) WithLogger(logger Logger) *AuthMiddleware {
	m.logger = logger
	return m
}

// MustLoadYAML 从YAML加载配置（失败则panic）
func MustLoadYAML(filePath string) *Config {
	config := NewConfig()
	if err := config.LoadFromYAML(filePath); err != nil {
		panic(fmt.Sprintf("加载配置文件失败: %v", err))
	}
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("配置验证失败: %v", err))
	}
	return config
}

// MustLoadJSON 从JSON加载配置（失败则panic）
func MustLoadJSON(filePath string) *Config {
	config := NewConfig()
	if err := config.LoadFromJSON(filePath); err != nil {
		panic(fmt.Sprintf("加载配置文件失败: %v", err))
	}
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("配置验证失败: %v", err))
	}
	return config
}

// Authenticate 认证中间件处理函数
func (m *AuthMiddleware) Authenticate(errorHandler ...ErrorHandler) gin.HandlerFunc {
	handler := DefaultErrorHandler
	if len(errorHandler) > 0 && errorHandler[0] != nil {
		handler = errorHandler[0]
	}

	return func(c *gin.Context) {
		// 获取认证参数
		appID := c.GetHeader(HeaderAppID)
		timestamp := c.GetHeader(HeaderTimestamp)
		nonce := c.GetHeader(HeaderNonce)
		sign := c.GetHeader(HeaderSign)

		m.logger.Debug("认证请求", "appId", appID, "timestamp", timestamp, "path", c.Request.URL.Path)

		// 如果Header中没有，尝试从查询参数获取
		if appID == "" {
			appID = c.Query("appId")
			timestamp = c.Query("timestamp")
			nonce = c.Query("nonce")
			sign = c.Query("sign")
		}

		// 验证必要参数
		if appID == "" || timestamp == "" || nonce == "" {
			m.logger.Error("认证参数不完整", "appId", appID)
			handler(c, http.StatusBadRequest, "认证参数不完整", "appId, timestamp, nonce 参数不能为空")
			c.Abort()
			return
		}

		// 获取应用配置
		app, exists := m.config.GetApp(appID)
		if !exists || !app.Enabled {
			m.logger.Error("应用不存在或已禁用", "appId", appID)
			handler(c, http.StatusUnauthorized, "应用不存在或已禁用", "")
			c.Abort()
			return
		}

		// 验证时间戳
		if !ValidateTimestamp(timestamp, m.config.TimestampTolerance) {
			m.logger.Error("时间戳无效", "timestamp", timestamp)
			handler(c, http.StatusBadRequest, "时间戳无效", "时间戳格式错误或超出有效期")
			c.Abort()
			return
		}

		// 验证IP白名单
		if m.config.EnableIPCheck && len(app.IPWhitelist) > 0 {
			if !m.checkIPWhitelist(c.ClientIP(), app.IPWhitelist) {
				m.logger.Error("IP不在白名单中", "ip", c.ClientIP(), "appId", appID)
				handler(c, http.StatusForbidden, "IP地址不在白名单中", "")
				c.Abort()
				return
			}
		}

		// 速率限制检查
		if app.RateLimit > 0 {
			if !m.rateLimiter.Allow(appID, app.RateLimit) {
				m.logger.Error("超过速率限制", "appId", appID)
				handler(c, http.StatusTooManyRequests, "请求过于频繁", "")
				c.Abort()
				return
			}
		}

		// 提取请求参数用于签名验证
		params, err := m.extractRequestParams(c, appID, timestamp, nonce, app)
		if err != nil {
			m.logger.Error("参数提取失败", "error", err)
			handler(c, http.StatusBadRequest, "参数提取失败", err.Error())
			c.Abort()
			return
		}

		// 验证签名
		if app.RequireSign {
			if sign == "" {
				m.logger.Error("签名参数缺失", "appId", appID)
				handler(c, http.StatusBadRequest, "认证参数不完整", "此应用需要提供签名参数 X-Sign")
				c.Abort()
				return
			}

			if !VerifySign(params, app.AppSecret, sign) {
				m.logger.Error("签名验证失败", "appId", appID)
				handler(c, http.StatusUnauthorized, "签名验证失败", "")
				c.Abort()
				return
			}
		}

		m.logger.Info("认证成功", "appId", appID)

		// 将应用信息存储到上下文中
		c.Set(ContextKeyApp, app)
		c.Set(ContextKeyAppID, app.AppID)

		c.Next()
	}
}

// checkIPWhitelist 检查IP是否在白名单中
func (m *AuthMiddleware) checkIPWhitelist(clientIP string, whitelist []string) bool {
	if len(whitelist) == 0 {
		return true
	}

	// 处理本地回环地址
	if clientIP == "" || clientIP == "::1" || strings.HasPrefix(clientIP, "127.") {
		for _, ip := range whitelist {
			if ip == "*" || ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
				return true
			}
		}
		return false
	}

	// 检查白名单
	for _, ip := range whitelist {
		ip = strings.TrimSpace(ip)
		if ip == "*" || ip == clientIP {
			return true
		}
		// 简单的通配符支持
		if strings.HasSuffix(ip, "*") {
			prefix := strings.TrimSuffix(ip, "*")
			if strings.HasPrefix(clientIP, prefix) {
				return true
			}
		}
	}

	return false
}

// extractRequestParams 提取请求参数用于签名验证
func (m *AuthMiddleware) extractRequestParams(c *gin.Context, appID, timestamp, nonce string, app *AppConfig) (map[string]string, error) {
	params := map[string]string{
		"appId":     appID,
		"timestamp": timestamp,
		"nonce":     nonce,
	}

	// 签名不包含请求体，但需要保留body供后续处理
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "application/json") {
			bodyBytes, err := c.GetRawData()
			if err != nil {
				return nil, err
			}
			if len(bodyBytes) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
	}

	// 注意：URL 查询参数不参与签名计算
	// 这样设计是为了简化客户端使用，降低集成难度
	// 认证安全性由 appId + timestamp + nonce + HMAC-SHA256 保证

	return params, nil
}

// GetAppFromContext 从上下文中获取应用信息
func GetAppFromContext(c *gin.Context) (*AppConfig, bool) {
	if app, exists := c.Get(ContextKeyApp); exists {
		if apiApp, ok := app.(*AppConfig); ok {
			return apiApp, true
		}
	}
	return nil, false
}

// GetAppIDFromContext 从上下文中获取应用ID
func GetAppIDFromContext(c *gin.Context) (string, bool) {
	if appID, exists := c.Get(ContextKeyAppID); exists {
		if id, ok := appID.(string); ok {
			return id, true
		}
	}
	return "", false
}

// RateLimiter 速率限制器
type RateLimiter struct {
	counters map[string]*counter
	mu       sync.RWMutex
}

type counter struct {
	count     int
	resetTime time.Time
	mu        sync.Mutex
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		counters: make(map[string]*counter),
	}
	
	// 启动清理协程
	go rl.cleanup()
	
	return rl
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(appID string, limit int) bool {
	rl.mu.Lock()
	c, exists := rl.counters[appID]
	if !exists {
		c = &counter{
			count:     0,
			resetTime: time.Now().Add(time.Minute),
		}
		rl.counters[appID] = c
	}
	rl.mu.Unlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if now.After(c.resetTime) {
		c.count = 0
		c.resetTime = now.Add(time.Minute)
	}

	if c.count >= limit {
		return false
	}

	c.count++
	return true
}

// cleanup 定期清理过期的计数器
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for appID, c := range rl.counters {
			c.mu.Lock()
			if now.After(c.resetTime.Add(5 * time.Minute)) {
				delete(rl.counters, appID)
			}
			c.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}
