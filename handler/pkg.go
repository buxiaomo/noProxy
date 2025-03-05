package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

func isHopHeader(header string) bool {
	hopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
	}
	for _, h := range hopHeaders {
		if strings.EqualFold(header, h) {
			return true
		}
	}
	return false
}

func in(target string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, target)
	if index < len(strArray) && strArray[index] == target {
		return true
	}
	return false
}

// 解析Docker Registry的WWW-Authenticate头部
func parseAuthHeader(header string) (string, string, string) {
	if !strings.HasPrefix(header, "Bearer ") {
		return "", "", ""
	}

	params := strings.TrimPrefix(header, "Bearer ")
	fields := strings.Fields(params)

	var realm, service, scope string
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			continue
		}

		value := strings.Trim(parts[1], "\",")
		switch parts[0] {
		case "realm":
			realm = value
		case "service":
			service = value
		case "scope":
			scope = value
		}
	}

	return realm, service, scope
}

// 获取Docker Registry的认证token
func getAuthToken(realm, service, scope string) (string, error) {
	log.Printf("[getAuthToken] 开始获取认证token, realm: %s, service: %s, scope: %s", realm, service, scope)

	if realm == "" {
		log.Printf("[getAuthToken] 认证realm为空")
		return "", fmt.Errorf("认证realm为空")
	}

	// 特殊处理docker.io的认证
	if strings.Contains(realm, "auth.docker.io") {
		log.Printf("[getAuthToken] 检测到docker.io认证，使用特殊处理")
		realm = "https://auth.docker.io/token"
	}

	params := url.Values{}
	if service != "" {
		params.Set("service", service)
	}
	if scope != "" {
		params.Set("scope", scope)
	}

	authURL := fmt.Sprintf("%s?%s", realm, params.Encode())
	log.Printf("[getAuthToken] 发送认证请求: %s", authURL)

	resp, err := http.Get(authURL)
	if err != nil {
		log.Printf("[getAuthToken] 请求认证服务失败: %v", err)
		return "", fmt.Errorf("请求认证服务失败: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("[getAuthToken] 认证服务响应状态码: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		log.Printf("[getAuthToken] 认证服务返回错误状态码: %d", resp.StatusCode)
		return "", fmt.Errorf("认证服务返回错误状态码: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[getAuthToken] 读取认证响应失败: %v", err)
		return "", fmt.Errorf("读取认证响应失败: %v", err)
	}

	var tokenResp struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		log.Printf("[getAuthToken] 解析认证响应失败: %v", err)
		return "", fmt.Errorf("解析认证响应失败: %v", err)
	}

	log.Printf("[getAuthToken] 成功获取认证token")
	return tokenResp.Token, nil
}
