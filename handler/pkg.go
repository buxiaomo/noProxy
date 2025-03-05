package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	if realm == "" {
		return "", fmt.Errorf("认证realm为空")
	}

	// 特殊处理docker.io的认证
	if strings.Contains(realm, "auth.docker.io") {
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
	resp, err := http.Get(authURL)
	if err != nil {
		return "", fmt.Errorf("请求认证服务失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("认证服务返回错误状态码: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取认证响应失败: %v", err)
	}

	var tokenResp struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("解析认证响应失败: %v", err)
	}

	return tokenResp.Token, nil
}
