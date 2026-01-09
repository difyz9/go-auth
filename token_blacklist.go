package goauth

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// TokenBlacklist Token 黑名单接口
type TokenBlacklist interface {
	// Add 添加 Token 到黑名单
	Add(tokenHash string, expiry time.Time) error
	
	// IsBlacklisted 检查 Token 是否在黑名单中
	IsBlacklisted(tokenHash string) bool
	
	// Cleanup 清理过期的黑名单记录
	Cleanup() error
	
	// Count 获取黑名单中的 Token 数量
	Count() int
}

// HashToken 计算 Token 的哈希值
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// InMemoryBlacklist 基于内存的黑名单实现
type InMemoryBlacklist struct {
	tokens         map[string]time.Time
	mu             sync.RWMutex
	cleanupTicker  *time.Ticker
	stopCleanup    chan struct{}
	cleanupRunning bool
}

// NewInMemoryBlacklist 创建内存黑名单
func NewInMemoryBlacklist() *InMemoryBlacklist {
	return &InMemoryBlacklist{
		tokens:      make(map[string]time.Time),
		stopCleanup: make(chan struct{}),
	}
}

// NewInMemoryBlacklistWithCleanup 创建带自动清理的内存黑名单
func NewInMemoryBlacklistWithCleanup(interval time.Duration) *InMemoryBlacklist {
	bl := NewInMemoryBlacklist()
	bl.StartAutoCleanup(interval)
	return bl
}

// Add 添加 Token 到黑名单
func (b *InMemoryBlacklist) Add(tokenHash string, expiry time.Time) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.tokens[tokenHash] = expiry
	return nil
}

// IsBlacklisted 检查 Token 是否在黑名单中
func (b *InMemoryBlacklist) IsBlacklisted(tokenHash string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	expiry, exists := b.tokens[tokenHash]
	if !exists {
		return false
	}
	
	// 如果已过期，不在黑名单中
	if time.Now().After(expiry) {
		return false
	}
	
	return true
}

// Cleanup 清理过期的黑名单记录
func (b *InMemoryBlacklist) Cleanup() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	now := time.Now()
	for hash, expiry := range b.tokens {
		if now.After(expiry) {
			delete(b.tokens, hash)
		}
	}
	
	return nil
}

// Count 获取黑名单中的 Token 数量
func (b *InMemoryBlacklist) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	return len(b.tokens)
}

// StartAutoCleanup 启动自动清理
func (b *InMemoryBlacklist) StartAutoCleanup(interval time.Duration) {
	b.mu.Lock()
	if b.cleanupRunning {
		b.mu.Unlock()
		return
	}
	b.cleanupRunning = true
	b.cleanupTicker = time.NewTicker(interval)
	b.mu.Unlock()
	
	go func() {
		for {
			select {
			case <-b.cleanupTicker.C:
				b.Cleanup()
			case <-b.stopCleanup:
				return
			}
		}
	}()
}

// StopAutoCleanup 停止自动清理
func (b *InMemoryBlacklist) StopAutoCleanup() {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if !b.cleanupRunning {
		return
	}
	
	b.cleanupRunning = false
	if b.cleanupTicker != nil {
		b.cleanupTicker.Stop()
	}
	close(b.stopCleanup)
}

// Clear 清空黑名单
func (b *InMemoryBlacklist) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.tokens = make(map[string]time.Time)
}

// GetExpiry 获取 Token 的过期时间
func (b *InMemoryBlacklist) GetExpiry(tokenHash string) (time.Time, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	expiry, exists := b.tokens[tokenHash]
	return expiry, exists
}

// TokenBlacklistConfig Token 黑名单配置
type TokenBlacklistConfig struct {
	Type            string        `yaml:"type" json:"type"`                         // blacklist 类型: memory, redis
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"` // 清理间隔
	
	// Redis 配置（如果使用 Redis）
	Redis RedisConfig `yaml:"redis" json:"redis"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `yaml:"addr" json:"addr"`         // Redis 地址
	Password string `yaml:"password" json:"password"` // Redis 密码
	DB       int    `yaml:"db" json:"db"`             // Redis 数据库
}

// DefaultTokenBlacklistConfig 默认黑名单配置
func DefaultTokenBlacklistConfig() TokenBlacklistConfig {
	return TokenBlacklistConfig{
		Type:            "memory",
		CleanupInterval: 1 * time.Hour,
	}
}
