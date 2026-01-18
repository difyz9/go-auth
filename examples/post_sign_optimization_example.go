package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/difyz9/go-auth"
)

func main() {
	// ========== 示例1：基础配置（默认不包含请求体） ==========
	fmt.Println("=== 示例1：基础配置 ===")
	
	config := goauth.NewConfig()
	
	// 添加应用
	config.AddApp(&goauth.AppConfig{
		AppID:       "test-app-001",
		AppSecret:   "tmcf5m6qcm6k9hrp3sy8rhgafu00ttph",
		AppName:     "测试应用",
		Enabled:     true,
		RequireSign: true,
	})
	
	// 创建中间件
	middleware := goauth.NewAuthMiddleware(goauth.Options{
		Config: config,
	})
	
	// 创建Gin路由
	r := gin.Default()
	r.Use(middleware.Authenticate())
	
	// POST接口（不验证请求体）
	r.POST("/api/test", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "POST请求成功",
			"data":    req,
		})
	})
	
	fmt.Println("服务器启动在 :8089")
	
	// ========== 客户端示例（不包含请求体） ==========
	go func() {
		// 等待服务器启动
		// time.Sleep(time.Second)
		
		client := goauth.NewClient(
			"http://localhost:8089",
			"test-app-001",
			"tmcf5m6qcm6k9hrp3sy8rhgafu00ttph",
			goauth.WithDebug(true), // 启用调试
		)
		
		// POST请求
		var result map[string]interface{}
		err := client.PostJSON("/api/test", map[string]interface{}{
			"user_id": 123,
			"amount":  99.99,
			"text":    "Hello, World!",
		}, &result)
		
		if err != nil {
			log.Printf("请求失败: %v", err)
		} else {
			log.Printf("请求成功: %+v", result)
		}
	}()
	
	r.Run(":8089")
}

// exampleHighSecurity 之前的高安全模式示例已移除，因为不再支持请求体签名
		
		c.JSON(http.StatusOK, gin.H{
			"message": "支付成功",
			"amount":  req.Amount,
		})
	})
	
	// 客户端（高安全模式）
	secureClient := goauth.NewClient(
		"http://localhost:8089",
		"secure-app",
		"secure-secret",
		goauth.WithSignIncludeBody(true), // 启用请求体签名
		goauth.WithDebug(true),
	)
	
	var result map[string]interface{}
	err := secureClient.PostJSON("/api/payment", map[string]interface{}{
		"amount":  1000.00,
		"card_no": "1234-5678-9012-3456",
	}, &result)
	
	if err != nil {
		log.Printf("支付请求失败: %v", err)
	} else {
		log.Printf("支付成功: %+v", result)
	}
}

// ========== 示例3：从配置文件加载 ==========
func exampleLoadFromFile() {
	fmt.Println("=== 示例3：从配置文件加载 ===")
	
	// 从YAML加载配置
	config := goauth.NewConfig()
	if err := config.LoadFromYAML("goauth_config_optimized.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	
	middleware := goauth.NewAuthMiddleware(goauth.Options{
		Config: config,
	})
	
	r := gin.Default()
	r.Use(middleware.Authenticate())
	
	// API接口
	r.POST("/api/v1/tts", func(c *gin.Context) {
		var req struct {
			Text  string `json:"text"`
			Voice string `json:"voice"`
		}
		
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "TTS请求成功",
			"text":    req.Text,
			"voice":   req.Voice,
		})
	})
	
	r.Run(":8089")
}

// ========== 示例4：签名调试 ==========
func exampleSignDebug() {
	fmt.Println("=== 示例4：签名调试 ===")
	
	// 创建客户端
	client := goauth.NewClient(
		"http://localhost:8089",
		"test-app-001",
		"tmcf5m6qcm6k9hrp3sy8rhgafu00ttph",
		goauth.WithSignIncludeBody(false), // 不包含请求体
		goauth.WithDebug(true),
	)
	
	// 调试签名生成
	requestBody := map[string]interface{}{
		"text":  "Hello",
		"voice": "zh-CN",
	}
	
	client.DebugSign(requestBody)
	
	// 输出示例：
	// === 签名调试信息 ===
	// AppID:     test-app-001
	// AppSecret: tmcf5m6qcm6k9hrp3sy8rhgafu00ttph
	// Timestamp: 1732089600
	// Nonce:     abc123xyz456
	//
	// 签名参数:
	//   appId: test-app-001
	//   timestamp: 1732089600
	//   nonce: abc123xyz456
	//
	// 生成的签名: 1a2b3c4d5e6f...
	// ==================
}

// ========== 示例5：混合使用 ==========
func exampleMixedUsage() {
	fmt.Println("=== 示例5：混合使用 ===")
	
	config := goauth.NewConfig(
		goauth.WithSignIncludeBody(false), // 全局默认false
	)
	
	// 普通应用
	config.AddApp(&goauth.AppConfig{
		AppID:       "app1",
		AppSecret:   "secret1",
		Enabled:     true,
		RequireSign: true,
	})
	
	// 高安全应用
	includeBody := true
	config.AddApp(&goauth.AppConfig{
		AppID:           "app2",
		AppSecret:       "secret2",
		Enabled:         true,
		RequireSign:     true,
		SignIncludeBody: &includeBody,
	})
	
	middleware := goauth.NewAuthMiddleware(goauth.Options{
		Config: config,
	})
	
	r := gin.Default()
	r.Use(middleware.Authenticate())
	
	// 普通接口
	r.POST("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "普通接口，签名不包含请求体"})
	})
	
	// 敏感接口
	r.POST("/api/payment", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "敏感接口，签名包含请求体"})
	})
	
	// 客户端1（不包含请求体）
	client1 := goauth.NewClient(
		"http://localhost:8089",
		"app1",
		"secret1",
		goauth.WithSignIncludeBody(false),
	)
	
	client1.PostJSON("/api/data", map[string]interface{}{
		"key": "value",
	}, nil)
	
	// 客户端2（包含请求体）
	client2 := goauth.NewClient(
		"http://localhost:8089",
		"app2",
		"secret2",
		goauth.WithSignIncludeBody(true),
	)
	
	client2.PostJSON("/api/payment", map[string]interface{}{
		"amount": 100.00,
	}, nil)
}
