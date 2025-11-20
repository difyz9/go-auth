package main

import (
	"fmt"
	"time"

	"github.com/difyz9/go-auth"
	"github.com/gin-gonic/gin"
)

func main() {
	// 方式1：最简单的配置方式
	simpleExample()
	
	// 方式2：使用配置构建器
	// builderExample()
	
	// 方式3：从文件加载
	// fileExample()
}

// ====================
// 最简单的使用方式
// ====================
func simpleExample() {
	// 1. 创建配置（一行代码）
	config := goauth.QuickConfig("test-app", "test-secret-key-12345678")
	
	// 2. 创建中间件并启动服务
	r := gin.Default()
	r.Use(goauth.New(config).Authenticate())
	
	r.GET("/api/users", func(c *gin.Context) {
		c.JSON(200, gin.H{"users": []string{"Alice", "Bob"}})
	})
	
	fmt.Println("服务器启动在 :8080")
	
	// 3. 在另一个goroutine中测试客户端
	go testClient()
	
	r.Run(":8080")
}

// ====================
// 使用配置构建器
// ====================
func builderExample() {
	// 使用链式调用构建配置
	config := goauth.NewConfigBuilder().
		SetTimestampTolerance(600).
		SetDefaultRateLimit(2000).
		SetEnableIPCheck(false).
		AddSimpleApp("app-001", "secret-001", "应用1").
		AddSimpleApp("app-002", "secret-002", "应用2").
		MustBuild()
	
	r := gin.Default()
	
	// 使用链式调用创建中间件
	auth := goauth.New(config).
		WithLogger(&customLogger{})
	
	r.Use(auth.Authenticate())
	
	r.GET("/api/data", handleData)
	
	r.Run(":8080")
}

// ====================
// 从配置文件加载
// ====================
func fileExample() {
	// 从文件加载配置（自动验证）
	config := goauth.MustLoadYAML("goauth_config.yaml")
	
	r := gin.Default()
	r.Use(goauth.New(config).Authenticate())
	
	r.GET("/api/info", handleInfo)
	
	r.Run(":8080")
}

// ====================
// 客户端调用示例
// ====================
func testClient() {
	// 等待服务器启动
	time.Sleep(1 * time.Second)
	
	// 方式1：使用便捷客户端（推荐）
	client := goauth.NewClient(
		"http://localhost:8080",
		"test-app",
		"test-secret-key-12345678",
		goauth.WithDebug(true), // 启用调试模式
	)
	
	// GET 请求
	var result map[string]interface{}
	if err := client.GetJSON("/api/users", &result); err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	
	fmt.Println("\n✅ 请求成功:")
	fmt.Printf("响应: %+v\n", result)
	
	// POST 请求
	orderData := map[string]interface{}{
		"user_id": 123,
		"amount":  99.99,
	}
	
	var orderResult map[string]interface{}
	if err := client.PostJSON("/api/orders", orderData, &orderResult); err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	
	fmt.Println("\n✅ 订单创建成功:")
	fmt.Printf("响应: %+v\n", orderResult)
	
	// 方式2：手动构建签名（用于学习和调试）
	manualSignExample()
}

// 手动签名示例
func manualSignExample() {
	appID := "test-app"
	appSecret := "test-secret-key-12345678"
	
	// 使用便捷方法快速生成签名
	body := map[string]interface{}{
		"user_id": 123,
		"amount":  100.00,
	}
	
	params, sign, err := goauth.QuickSign(appID, appSecret, body)
	if err != nil {
		fmt.Println("签名生成失败:", err)
		return
	}
	
	fmt.Println("\n=== 签名信息 ===")
	fmt.Printf("AppID:     %s\n", params["appId"])
	fmt.Printf("Timestamp: %s\n", params["timestamp"])
	fmt.Printf("Nonce:     %s\n", params["nonce"])
	fmt.Printf("Sign:      %s\n", sign)
	fmt.Println("================")
}

// ====================
// 高级示例：自定义错误处理
// ====================
func advancedExample() {
	config := goauth.QuickConfig("test-app", "test-secret")
	
	r := gin.Default()
	
	// 自定义错误处理
	customErrorHandler := func(c *gin.Context, code int, message string, detail string) {
		requestID := goauth.BuildRequestID()
		
		// 使用统一的错误响应格式
		err := goauth.NewAuthError(
			goauth.ErrorCode(message),
			message,
			detail,
			code,
		)
		
		response := goauth.NewErrorResponse(err, requestID)
		c.JSON(code, response)
	}
	
	r.Use(goauth.New(config).Authenticate(customErrorHandler))
	
	r.GET("/api/secure", func(c *gin.Context) {
		// 使用统一的成功响应格式
		requestID := c.GetHeader("X-Request-ID")
		response := goauth.NewSuccessResponse(
			gin.H{"message": "Success"},
			requestID,
		)
		c.JSON(200, response)
	})
	
	r.Run(":8080")
}

// ====================
// 辅助函数
// ====================
func handleData(c *gin.Context) {
	app, _ := goauth.GetAppFromContext(c)
	c.JSON(200, gin.H{
		"app":  app.AppName,
		"data": []int{1, 2, 3, 4, 5},
	})
}

func handleInfo(c *gin.Context) {
	appID, _ := goauth.GetAppIDFromContext(c)
	c.JSON(200, gin.H{
		"app_id": appID,
		"info":   "系统运行正常",
	})
}

// 自定义日志器
type customLogger struct{}

func (l *customLogger) Info(msg string, fields ...interface{}) {
	fmt.Printf("[INFO] %s %v\n", msg, fields)
}

func (l *customLogger) Error(msg string, fields ...interface{}) {
	fmt.Printf("[ERROR] %s %v\n", msg, fields)
}

func (l *customLogger) Debug(msg string, fields ...interface{}) {
	fmt.Printf("[DEBUG] %s %v\n", msg, fields)
}

// ====================
// 工具函数使用示例
// ====================
func utilsExample() {
	// 生成安全的应用密钥
	secret, _ := goauth.GenerateAppSecret(32)
	fmt.Println("生成的密钥:", secret)
	
	// 遮蔽密钥用于日志
	masked := goauth.MaskSecret(secret)
	fmt.Println("遮蔽后:", masked)
	
	// 验证应用ID
	if err := goauth.ValidateAppID("my-app-123"); err != nil {
		fmt.Println("应用ID无效:", err)
	} else {
		fmt.Println("应用ID有效")
	}
	
	// 生成请求ID
	requestID := goauth.BuildRequestID()
	fmt.Println("请求ID:", requestID)
}
