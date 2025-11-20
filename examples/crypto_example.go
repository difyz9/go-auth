package main

import (
	"fmt"
	"time"

	"github.com/difyz9/go-auth"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 1. 创建认证配置
	config := goauth.QuickConfig("test-app", "test-secret-key-123456")

	// 2. 创建加解密配置
	cryptoConfig := goauth.NewCryptoConfig("my-aes-key-32-bytes-long-123456")
	cryptoConfig.EnableDecryption = true  // 启用请求解密
	cryptoConfig.EnableEncryption = true  // 启用响应加密
	cryptoConfig.ForceEncryption = false  // 不强制加密（由客户端控制）
	cryptoConfig.SkipPaths = []string{"/health", "/public"} // 跳过某些路径

	// 3. 创建加解密中间件
	cryptoMiddleware := goauth.NewCryptoMiddleware(cryptoConfig)

	// 4. 应用中间件（顺序很重要）
	// 先解密请求 -> 认证 -> 业务逻辑 -> 加密响应
	r.Use(cryptoMiddleware.DecryptRequest())       // 解密请求体
	r.Use(goauth.New(config).Authenticate())        // 认证
	r.Use(cryptoMiddleware.EncryptResponse())       // 加密响应体

	// 公开接口（无需认证和加密）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 需要认证和加解密的接口
	r.POST("/api/users", createUser)
	r.GET("/api/users/:id", getUser)
	r.POST("/api/orders", createOrder)

	fmt.Println("服务器启动在 :8080")
	fmt.Println("使用示例:")
	fmt.Println("1. 发送未加密请求（正常JSON）")
	fmt.Println("2. 发送加密请求（添加 X-Encrypted: 1 头）")
	fmt.Println("3. 请求加密响应（添加 X-Response-Encrypt: 1 头）")
	
	r.Run(":8080")
}

// 业务处理函数

func createUser(c *gin.Context) {
	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": "参数错误", "detail": err.Error()})
		return
	}

	// 检查是否经过解密
	wasDecrypted, _ := c.Get("crypto_decrypted")

	app, _ := goauth.GetAppFromContext(c)

	c.JSON(200, gin.H{
		"success":       true,
		"message":       "用户创建成功",
		"user":          user,
		"app":           app.AppName,
		"was_decrypted": wasDecrypted,
		"timestamp":     time.Now().Unix(),
	})
}

func getUser(c *gin.Context) {
	userID := c.Param("id")
	app, _ := goauth.GetAppFromContext(c)

	c.JSON(200, gin.H{
		"success": true,
		"user": gin.H{
			"id":    userID,
			"name":  "张三",
			"email": "zhangsan@example.com",
			"age":   30,
		},
		"app":       app.AppName,
		"timestamp": time.Now().Unix(),
	})
}

func createOrder(c *gin.Context) {
	var order struct {
		UserID   int     `json:"user_id"`
		Amount   float64 `json:"amount"`
		Products []struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Quantity int    `json:"quantity"`
		} `json:"products"`
	}

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	// 检查是否经过解密
	wasDecrypted, _ := c.Get("crypto_decrypted")
	originalSize, _ := c.Get("crypto_original_size")
	decryptedSize, _ := c.Get("crypto_decrypted_size")

	c.JSON(200, gin.H{
		"success":        true,
		"message":        "订单创建成功",
		"order_id":       fmt.Sprintf("ORD-%d", time.Now().Unix()),
		"order":          order,
		"was_decrypted":  wasDecrypted,
		"original_size":  originalSize,
		"decrypted_size": decryptedSize,
		"timestamp":      time.Now().Unix(),
	})
}
