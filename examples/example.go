package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"payment_service/pkg/goauth"
)

// ============================
// 服务端示例
// ============================

func runServer() {
	// 方式1：从配置文件加载
	config := goauth.NewConfig()
	if err := config.LoadFromYAML("goauth_config.yaml"); err != nil {
		fmt.Println("加载配置文件失败，使用内存配置:", err)
		
		// 方式2：直接在代码中配置
		config.AddApp(&goauth.AppConfig{
			AppID:       "test-app-001",
			AppSecret:   "test-secret-key-12345678",
			AppName:     "测试应用",
			RequireSign: true,
			Enabled:     true,
			RateLimit:   100,
			IPWhitelist: []string{"*"}, // 允许所有IP，生产环境请配置具体IP
		})
		
		config.AddApp(&goauth.AppConfig{
			AppID:       "no-sign-app",
			AppSecret:   "no-sign-secret",
			AppName:     "无需签名的应用",
			RequireSign: false, // 不需要签名验证
			Enabled:     true,
			IPWhitelist: []string{"*"},
		})
	}

	// 创建认证中间件
	authMiddleware := goauth.NewAuthMiddleware(goauth.Options{
		Config: config,
		// 可选：自定义日志
		// Logger: myCustomLogger,
	})

	// 创建 Gin 路由
	r := gin.Default()

	// 公开接口（无需认证）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 需要认证的接口
	api := r.Group("/api")
	api.Use(authMiddleware.Authenticate(customErrorHandler))
	{
		api.GET("/users", getUsers)
		api.POST("/orders", createOrder)
		api.GET("/profile", getProfile)
	}

	// 管理接口（自定义错误处理）
	admin := r.Group("/admin")
	admin.Use(authMiddleware.Authenticate(adminErrorHandler))
	{
		admin.GET("/stats", getStats)
		admin.POST("/config", updateConfig)
	}

	fmt.Println("服务器启动在 :8080")
	r.Run(":8080")
}

// 自定义错误处理
func customErrorHandler(c *gin.Context, code int, message string, detail string) {
	c.JSON(code, gin.H{
		"success":   false,
		"code":      code,
		"message":   message,
		"detail":    detail,
		"timestamp": time.Now().Unix(),
	})
}

func adminErrorHandler(c *gin.Context, code int, message string, detail string) {
	c.JSON(code, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
			"detail":  detail,
		},
		"request_id": c.GetHeader("X-Request-Id"),
	})
}

// API 处理函数
func getUsers(c *gin.Context) {
	app, _ := goauth.GetAppFromContext(c)
	appID, _ := goauth.GetAppIDFromContext(c)

	c.JSON(200, gin.H{
		"success": true,
		"app":     app.AppName,
		"app_id":  appID,
		"users": []map[string]interface{}{
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"},
		},
	})
}

func createOrder(c *gin.Context) {
	var order struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	appID, _ := goauth.GetAppIDFromContext(c)

	c.JSON(200, gin.H{
		"success":  true,
		"app_id":   appID,
		"order_id": "ORD" + strconv.FormatInt(time.Now().Unix(), 10),
		"amount":   order.Amount,
	})
}

func getProfile(c *gin.Context) {
	app, _ := goauth.GetAppFromContext(c)
	c.JSON(200, gin.H{
		"success": true,
		"profile": gin.H{
			"app_name":   app.AppName,
			"rate_limit": app.RateLimit,
		},
	})
}

func getStats(c *gin.Context) {
	c.JSON(200, gin.H{
		"success":       true,
		"total_requests": 12345,
		"active_apps":   10,
	})
}

func updateConfig(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"message": "配置更新成功",
	})
}

// ============================
// 客户端示例
// ============================

// Client API 客户端
type Client struct {
	BaseURL   string
	AppID     string
	AppSecret string
	HTTPClient *http.Client
}

// NewClient 创建新的客户端
func NewClient(baseURL, appID, appSecret string) *Client {
	return &Client{
		BaseURL:   baseURL,
		AppID:     appID,
		AppSecret: appSecret,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Request 发送认证请求
func (c *Client) Request(method, path string, body interface{}) ([]byte, error) {
	// 准备请求体
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	// 生成认证参数
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := goauth.GenerateNonce(16)

	// 生成签名参数
	params := map[string]string{
		"appId":     c.AppID,
		"timestamp": timestamp,
		"nonce":     nonce,
	}

	// 如果有请求体，加入签名
	if len(bodyBytes) > 0 {
		params["requestBody"] = string(bodyBytes)
	}

	// 生成签名
	sign := goauth.GenerateSign(params, c.AppSecret)

	// 创建请求
	var reqBody io.Reader
	if len(bodyBytes) > 0 {
		reqBody = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Id", c.AppID)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-Sign", sign)

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败: %d, %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// 客户端使用示例
func clientExample() {
	// 创建客户端
	client := NewClient(
		"http://localhost:8080",
		"test-app-001",
		"test-secret-key-12345678",
	)

	// 示例1：GET 请求
	fmt.Println("\n=== GET /api/users ===")
	respBody, err := client.Request("GET", "/api/users", nil)
	if err != nil {
		fmt.Println("请求失败:", err)
	} else {
		fmt.Println("响应:", string(respBody))
	}

	// 示例2：POST 请求
	fmt.Println("\n=== POST /api/orders ===")
	orderData := map[string]interface{}{
		"user_id": 123,
		"amount":  99.99,
	}
	respBody, err = client.Request("POST", "/api/orders", orderData)
	if err != nil {
		fmt.Println("请求失败:", err)
	} else {
		fmt.Println("响应:", string(respBody))
	}

	// 示例3：GET 请求
	fmt.Println("\n=== GET /api/profile ===")
	respBody, err = client.Request("GET", "/api/profile", nil)
	if err != nil {
		fmt.Println("请求失败:", err)
	} else {
		fmt.Println("响应:", string(respBody))
	}
}

// 简单的客户端调用示例
func simpleClientExample() {
	appID := "test-app-001"
	appSecret := "test-secret-key-12345678"

	// 准备请求参数
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := goauth.GenerateNonce(16)

	// 请求体
	requestBody := map[string]interface{}{
		"user_id": 123,
		"amount":  100.00,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	// 生成签名
	params := map[string]string{
		"appId":       appID,
		"timestamp":   timestamp,
		"nonce":       nonce,
		"requestBody": string(bodyBytes),
	}
	sign := goauth.GenerateSign(params, appSecret)

	// 创建请求
	req, _ := http.NewRequest("POST", "http://localhost:8080/api/orders", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Id", appID)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-Sign", sign)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println("响应状态:", resp.Status)
	fmt.Println("响应内容:", string(respBody))
}

func main() {
	// 启动服务器（在实际使用中，服务器和客户端通常在不同的进程）
	go runServer()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 运行客户端示例
	fmt.Println("\n========== 客户端调用示例 ==========")
	clientExample()

	// 保持程序运行
	select {}
}
