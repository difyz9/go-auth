package goauth_test

import (
	"testing"
	"time"

	"github.com/difyz9/go-auth"
	"github.com/stretchr/testify/assert"
)

func TestJWTService(t *testing.T) {
	config := goauth.JWTConfig{
		SecretKey:     "test-secret-key-must-be-at-least-32-chars-long",
		Issuer:        "test-issuer",
		AccessExpiry:  24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	jwtService := goauth.NewJWTService(config)
	assert.NotNil(t, jwtService)

	t.Run("GenerateAccessToken", func(t *testing.T) {
		token, err := jwtService.GenerateAccessToken(1, "testuser", "admin", "test-app", nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// 验证 Token
		claims, err := jwtService.ValidateAccessToken(token)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), claims.UserID)
		assert.Equal(t, "testuser", claims.Username)
		assert.Equal(t, "admin", claims.Role)
		assert.Equal(t, "test-app", claims.AppID)
	})

	t.Run("GenerateRefreshToken", func(t *testing.T) {
		token, err := jwtService.GenerateRefreshToken(1, "testuser")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// 验证 Token
		claims, err := jwtService.ValidateRefreshToken(token)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), claims.UserID)
		assert.Equal(t, "testuser", claims.Username)
	})

	t.Run("GenerateTokenPair", func(t *testing.T) {
		tokenPair, err := jwtService.GenerateTokenPair(1, "testuser", "admin", "test-app", nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenPair.AccessToken)
		assert.NotEmpty(t, tokenPair.RefreshToken)
		assert.Equal(t, "Bearer", tokenPair.TokenType)
	})

	t.Run("RefreshAccessToken", func(t *testing.T) {
		// 生成 Refresh Token
		refreshToken, err := jwtService.GenerateRefreshToken(1, "testuser")
		assert.NoError(t, err)

		// 刷新 Access Token
		newAccessToken, err := jwtService.RefreshAccessToken(refreshToken, "user", "test-app", nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, newAccessToken)

		// 验证新的 Access Token
		claims, err := jwtService.ValidateAccessToken(newAccessToken)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), claims.UserID)
		assert.Equal(t, "user", claims.Role)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		_, err := jwtService.ParseToken("invalid-token")
		assert.Error(t, err)
	})

	t.Run("RevokeToken", func(t *testing.T) {
		token, err := jwtService.GenerateAccessToken(1, "testuser", "admin", "test-app", nil)
		assert.NoError(t, err)

		// 撤销 Token
		err = jwtService.RevokeToken(token)
		assert.NoError(t, err)

		// 验证被撤销的 Token
		_, err = jwtService.ParseToken(token)
		assert.Equal(t, goauth.ErrTokenRevoked, err)
	})

	t.Run("CustomClaims", func(t *testing.T) {
		custom := map[string]interface{}{
			"department": "engineering",
			"level":      5,
		}

		token, err := jwtService.GenerateAccessToken(1, "testuser", "admin", "test-app", custom)
		assert.NoError(t, err)

		claims, err := jwtService.ValidateAccessToken(token)
		assert.NoError(t, err)
		assert.Equal(t, "engineering", claims.Custom["department"])
		assert.Equal(t, float64(5), claims.Custom["level"])
	})
}

func TestTokenBlacklist(t *testing.T) {
	t.Run("InMemoryBlacklist", func(t *testing.T) {
		bl := goauth.NewInMemoryBlacklist()
		
		tokenHash := goauth.HashToken("test-token")
		expiry := time.Now().Add(1 * time.Hour)

		// 添加到黑名单
		err := bl.Add(tokenHash, expiry)
		assert.NoError(t, err)
		assert.Equal(t, 1, bl.Count())

		// 检查是否在黑名单中
		assert.True(t, bl.IsBlacklisted(tokenHash))

		// 检查不存在的 Token
		assert.False(t, bl.IsBlacklisted("non-existent"))

		// 添加过期的 Token
		expiredHash := goauth.HashToken("expired-token")
		err = bl.Add(expiredHash, time.Now().Add(-1*time.Hour))
		assert.NoError(t, err)

		// 过期的 Token 不应该在黑名单中
		assert.False(t, bl.IsBlacklisted(expiredHash))

		// 清理
		err = bl.Cleanup()
		assert.NoError(t, err)
		assert.Equal(t, 1, bl.Count()) // 只剩下未过期的
	})

	t.Run("AutoCleanup", func(t *testing.T) {
		bl := goauth.NewInMemoryBlacklistWithCleanup(100 * time.Millisecond)
		defer bl.StopAutoCleanup()

		// 添加过期的 Token
		expiredHash := goauth.HashToken("expired-token")
		err := bl.Add(expiredHash, time.Now().Add(-1*time.Hour))
		assert.NoError(t, err)

		// 等待自动清理
		time.Sleep(200 * time.Millisecond)

		// 应该已经被清理
		assert.Equal(t, 0, bl.Count())
	})
}
