package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/difyz9/go-auth"
)

func main() {
	// 配置
	baseURL := "http://localhost:8080"
	appID := "test-app"
	appSecret := "test-secret-key-123456"
	aesKey := "my-aes-key-32-bytes-long-123456"

	fmt.Println("========== GoAuth 加解密客户端示例 ==========\n")

	// 示例1：发送未加密的请求
	fmt.Println("【示例1】发送未加密的请求")
	sendPlainRequest(baseURL, appID, appSecret)

	time.Sleep(1 * time.Second)

	// 示例2：发送加密的请求
	fmt.Println("\n【示例2】发送加密的请求")
	sendEncryptedRequest(baseURL, appID, appSecret, aesKey)

	time.Sleep(1 * time.Second)

	// 示例3：请求加密的响应
	fmt.Println("\n【示例3】请求加密的响应")
	requestEncryptedResponse(baseURL, appID, appSecret, aesKey)

	time.Sleep(1 * time.Second)

	// 示例4：同时加密请求和响应
	fmt.Println("\n【示例4】同时加密请求和响应")
	fullEncryptedCommunication(baseURL, appID, appSecret, aesKey)
}

// 示例1：发送未加密的请求
func sendPlainRequest(baseURL, appID, appSecret string) {
	client := goauth.NewClient(baseURL, appID, appSecret)

	userData := map[string]interface{}{
		"name":  "张三",
		"email": "zhangsan@example.com",
		"age":   30,
	}

	var result map[string]interface{}
	err := client.PostJSON("/api/users", userData, &result)
	if err != nil {
		fmt.Println("❌ 请求失败:", err)
		return
	}

	fmt.Println("✅ 请求成功")
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(resultJSON))
}

// 示例2：发送加密的请求
func sendEncryptedRequest(baseURL, appID, appSecret, aesKey string) {
	// 准备请求数据
	userData := map[string]interface{}{
		"name":  "李四",
		"email": "lisi@example.com",
		"age":   25,
	}

	// 序列化为JSON
	userDataJSON, _ := json.Marshal(userData)

	// AES加密
	encryptedData, err := goauth.AESEncrypt(aesKey, userDataJSON)
	if err != nil {
		fmt.Println("❌ 加密失败:", err)
		return
	}

	// 构建加密请求结构
	encryptedRequest := goauth.EncryptedRequest{
		Data:      encryptedData,
		Timestamp: time.Now().Unix(),
	}

	// 生成签名（注意：签名是对加密后的数据进行的）
	encReqJSON, _ := json.Marshal(encryptedRequest)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := goauth.GenerateNonce(16)

	params := map[string]string{
		"appId":       appID,
		"timestamp":   timestamp,
		"nonce":       nonce,
		"requestBody": string(encReqJSON),
	}
	sign := goauth.GenerateSign(params, appSecret)

	// 发送请求
	client := goauth.NewClient(baseURL, appID, appSecret)
	
	// 添加加密标识头
	headers := map[string]string{
		"X-Encrypted": "1",
	}

	resp, err := client.Post("/api/users", encryptedRequest, headers)
	if err != nil {
		fmt.Println("❌ 请求失败:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("❌ 解析响应失败:", err)
		return
	}

	fmt.Println("✅ 加密请求成功")
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(resultJSON))
}

// 示例3：请求加密的响应
func requestEncryptedResponse(baseURL, appID, appSecret, aesKey string) {
	client := goauth.NewClient(baseURL, appID, appSecret)

	// 添加请求加密响应的头
	headers := map[string]string{
		"X-Response-Encrypt": "1",
	}

	resp, err := client.Get("/api/users/123", headers)
	if err != nil {
		fmt.Println("❌ 请求失败:", err)
		return
	}
	defer resp.Body.Close()

	// 检查是否返回了加密响应
	if resp.Header.Get("X-Encrypted") == "1" {
		fmt.Println("✅ 收到加密响应")

		// 解析加密响应
		var encryptedResp goauth.EncryptedResponse
		if err := json.NewDecoder(resp.Body).Decode(&encryptedResp); err != nil {
			fmt.Println("❌ 解析加密响应失败:", err)
			return
		}

		// AES解密
		decryptedData, err := goauth.AESDecrypt(aesKey, encryptedResp.Data)
		if err != nil {
			fmt.Println("❌ 解密失败:", err)
			return
		}

		// 解析解密后的数据
		var result map[string]interface{}
		if err := json.Unmarshal(decryptedData, &result); err != nil {
			fmt.Println("❌ 解析解密数据失败:", err)
			return
		}

		fmt.Println("✅ 解密成功")
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(resultJSON))
	} else {
		fmt.Println("⚠️  收到未加密响应")
	}
}

// 示例4：同时加密请求和响应
func fullEncryptedCommunication(baseURL, appID, appSecret, aesKey string) {
	// 准备订单数据
	orderData := map[string]interface{}{
		"user_id": 123,
		"amount":  999.99,
		"products": []map[string]interface{}{
			{
				"id":       1,
				"name":     "商品A",
				"quantity": 2,
			},
			{
				"id":       2,
				"name":     "商品B",
				"quantity": 1,
			},
		},
	}

	// 序列化为JSON
	orderDataJSON, _ := json.Marshal(orderData)
	fmt.Printf("📤 原始数据大小: %d 字节\n", len(orderDataJSON))

	// AES加密请求
	encryptedData, err := goauth.AESEncrypt(aesKey, orderDataJSON)
	if err != nil {
		fmt.Println("❌ 加密失败:", err)
		return
	}
	fmt.Printf("🔐 加密后数据大小: %d 字节\n", len(encryptedData))

	// 构建加密请求
	encryptedRequest := goauth.EncryptedRequest{
		Data:      encryptedData,
		Timestamp: time.Now().Unix(),
	}

	// 发送请求（签名由客户端SDK自动处理）
	client := goauth.NewClient(baseURL, appID, appSecret)
	
	headers := map[string]string{
		"X-Encrypted":        "1", // 标识请求已加密
		"X-Response-Encrypt": "1", // 请求加密响应
	}

	resp, err := client.Post("/api/orders", encryptedRequest, headers)
	if err != nil {
		fmt.Println("❌ 请求失败:", err)
		return
	}
	defer resp.Body.Close()

	// 处理加密响应
	if resp.Header.Get("X-Encrypted") == "1" {
		fmt.Println("✅ 收到加密响应")

		// 解析加密响应
		var encryptedResp goauth.EncryptedResponse
		if err := json.NewDecoder(resp.Body).Decode(&encryptedResp); err != nil {
			fmt.Println("❌ 解析加密响应失败:", err)
			return
		}

		fmt.Printf("📥 加密响应大小: %d 字节\n", len(encryptedResp.Data))

		// AES解密
		decryptedData, err := goauth.AESDecrypt(aesKey, encryptedResp.Data)
		if err != nil {
			fmt.Println("❌ 解密失败:", err)
			return
		}

		fmt.Printf("🔓 解密后数据大小: %d 字节\n", len(decryptedData))

		// 解析解密后的数据
		var result map[string]interface{}
		if err := json.Unmarshal(decryptedData, &result); err != nil {
			fmt.Println("❌ 解析解密数据失败:", err)
			return
		}

		fmt.Println("✅ 完整加密通信成功！")
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(resultJSON))
	} else {
		fmt.Println("⚠️  收到未加密响应")
	}
}

// 便捷的加密客户端封装
type CryptoClient struct {
	*goauth.Client
	AESKey string
}

// NewCryptoClient 创建支持加解密的客户端
func NewCryptoClient(baseURL, appID, appSecret, aesKey string) *CryptoClient {
	return &CryptoClient{
		Client: goauth.NewClient(baseURL, appID, appSecret),
		AESKey: aesKey,
	}
}

// PostEncrypted 发送加密的POST请求并解密响应
func (c *CryptoClient) PostEncrypted(path string, data interface{}) (map[string]interface{}, error) {
	// 序列化数据
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// 加密
	encryptedData, err := goauth.AESEncrypt(c.AESKey, dataJSON)
	if err != nil {
		return nil, err
	}

	// 构建加密请求
	encReq := goauth.EncryptedRequest{
		Data:      encryptedData,
		Timestamp: time.Now().Unix(),
	}

	// 发送请求
	headers := map[string]string{
		"X-Encrypted":        "1",
		"X-Response-Encrypt": "1",
	}

	resp, err := c.Client.Post(path, encReq, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析加密响应
	var encResp goauth.EncryptedResponse
	if err := json.NewDecoder(resp.Body).Decode(&encResp); err != nil {
		return nil, err
	}

	// 解密
	decryptedData, err := goauth.AESDecrypt(c.AESKey, encResp.Data)
	if err != nil {
		return nil, err
	}

	// 解析结果
	var result map[string]interface{}
	if err := json.Unmarshal(decryptedData, &result); err != nil {
		return nil, err
	}

	return result, nil
}
