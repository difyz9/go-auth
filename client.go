package goauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Client API认证客户端
type Client struct {
	BaseURL    string
	AppID      string
	AppSecret  string
	HTTPClient *http.Client
	Debug      bool // 调试模式
}

// ClientOption 客户端配置选项
type ClientOption func(*Client)

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.HTTPClient.Timeout = timeout
	}
}

// WithDebug 启用调试模式
func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		c.Debug = debug
	}
}

// WithHTTPClient 自定义HTTP客户端
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// NewClient 创建新的API客户端
func NewClient(baseURL, appID, appSecret string, opts ...ClientOption) *Client {
	client := &Client{
		BaseURL:   baseURL,
		AppID:     appID,
		AppSecret: appSecret,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Debug: false,
	}
	
	// 应用选项
	for _, opt := range opts {
		opt(client)
	}
	
	return client
}

// Request 发送认证请求
func (c *Client) Request(method, path string, body interface{}, headers ...map[string]string) (*http.Response, error) {
	// 准备请求体
	var bodyBytes []byte
	var err error
	
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
	}
	
	// 生成认证参数
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := GenerateNonce(16)
	
	// 构建签名参数
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
	sign := GenerateSign(params, c.AppSecret)
	
	// 调试输出
	if c.Debug {
		fmt.Printf("[GoAuth Client] Request: %s %s\n", method, path)
		fmt.Printf("[GoAuth Client] AppID: %s\n", c.AppID)
		fmt.Printf("[GoAuth Client] Timestamp: %s\n", timestamp)
		fmt.Printf("[GoAuth Client] Nonce: %s\n", nonce)
		fmt.Printf("[GoAuth Client] Sign: %s\n", sign)
		if len(bodyBytes) > 0 {
			fmt.Printf("[GoAuth Client] Body: %s\n", string(bodyBytes))
		}
	}
	
	// 创建请求
	var reqBody io.Reader
	if len(bodyBytes) > 0 {
		reqBody = bytes.NewBuffer(bodyBytes)
	}
	
	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	
	// 设置默认请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(HeaderAppID, c.AppID)
	req.Header.Set(HeaderTimestamp, timestamp)
	req.Header.Set(HeaderNonce, nonce)
	req.Header.Set(HeaderSign, sign)
	
	// 设置自定义请求头
	if len(headers) > 0 {
		for _, h := range headers {
			for k, v := range h {
				req.Header.Set(k, v)
			}
		}
	}
	
	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	
	return resp, nil
}

// Get 发送GET请求
func (c *Client) Get(path string, headers ...map[string]string) (*http.Response, error) {
	return c.Request(http.MethodGet, path, nil, headers...)
}

// Post 发送POST请求
func (c *Client) Post(path string, body interface{}, headers ...map[string]string) (*http.Response, error) {
	return c.Request(http.MethodPost, path, body, headers...)
}

// Put 发送PUT请求
func (c *Client) Put(path string, body interface{}, headers ...map[string]string) (*http.Response, error) {
	return c.Request(http.MethodPut, path, body, headers...)
}

// Delete 发送DELETE请求
func (c *Client) Delete(path string, headers ...map[string]string) (*http.Response, error) {
	return c.Request(http.MethodDelete, path, nil, headers...)
}

// DoJSON 发送请求并解析JSON响应
func (c *Client) DoJSON(method, path string, body interface{}, result interface{}) error {
	resp, err := c.Request(method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	
	if c.Debug {
		fmt.Printf("[GoAuth Client] Response Status: %d\n", resp.StatusCode)
		fmt.Printf("[GoAuth Client] Response Body: %s\n", string(respBody))
	}
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("请求失败: %d, %s", resp.StatusCode, string(respBody))
	}
	
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}
	
	return nil
}

// GetJSON 发送GET请求并解析JSON响应
func (c *Client) GetJSON(path string, result interface{}) error {
	return c.DoJSON(http.MethodGet, path, nil, result)
}

// PostJSON 发送POST请求并解析JSON响应
func (c *Client) PostJSON(path string, body interface{}, result interface{}) error {
	return c.DoJSON(http.MethodPost, path, body, result)
}

// PutJSON 发送PUT请求并解析JSON响应
func (c *Client) PutJSON(path string, body interface{}, result interface{}) error {
	return c.DoJSON(http.MethodPut, path, body, result)
}

// DeleteJSON 发送DELETE请求并解析JSON响应
func (c *Client) DeleteJSON(path string, result interface{}) error {
	return c.DoJSON(http.MethodDelete, path, nil, result)
}

// DebugSign 调试签名生成（用于排查签名问题）
func (c *Client) DebugSign(body interface{}) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := GenerateNonce(16)
	
	params := map[string]string{
		"appId":     c.AppID,
		"timestamp": timestamp,
		"nonce":     nonce,
	}
	
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		if len(bodyBytes) > 0 {
			params["requestBody"] = string(bodyBytes)
		}
	}
	
	sign := GenerateSign(params, c.AppSecret)
	
	fmt.Println("=== 签名调试信息 ===")
	fmt.Printf("AppID:     %s\n", c.AppID)
	fmt.Printf("AppSecret: %s\n", c.AppSecret)
	fmt.Printf("Timestamp: %s\n", timestamp)
	fmt.Printf("Nonce:     %s\n", nonce)
	fmt.Println("\n签名参数:")
	for k, v := range params {
		if k == "requestBody" && len(v) > 100 {
			fmt.Printf("  %s: %s...\n", k, v[:100])
		} else {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}
	fmt.Printf("\n生成的签名: %s\n", sign)
	fmt.Println("==================")
}
