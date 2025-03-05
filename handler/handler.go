package handler

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func ProxyHandler(c *gin.Context) {
	DownloadUrl := c.Param("url")
	DownloadUrl = DownloadUrl[1:len(DownloadUrl)]
	log.Printf("[ProxyHandler] 收到下载请求: %s, 客户端IP: %s", DownloadUrl, c.ClientIP())

	targetURL, err := url.Parse(DownloadUrl)
	if err != nil {
		log.Printf("[ProxyHandler] URL解析失败: %v, URL: %s", err, DownloadUrl)
		c.String(http.StatusBadRequest, "无效的URL格式")
		return
	}

	if in(targetURL.Host, viper.GetStringSlice("whiteList")) == false {
		log.Printf("[ProxyHandler] 域名不在白名单中: %s", targetURL.Host)
		c.String(http.StatusForbidden, "目标域名不在白名单中")
		return
	}

	if targetURL.Scheme != "http" && targetURL.Scheme != "https" {
		log.Printf("[ProxyHandler] 不支持的协议: %s", targetURL.Scheme)
		c.String(http.StatusBadRequest, "URL必须是http或https协议")
		return
	}

	client := &http.Client{
		Timeout: 30 * time.Minute,
	}
	resp, err := client.Get(DownloadUrl)
	if err != nil {
		log.Printf("[ProxyHandler] 请求目标URL失败: %v", err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("获取资源失败: %v", err))
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		if isHopHeader(k) {
			continue
		}
		for _, v := range vv {
			c.Header(k, v)
		}
	}

	if c.GetHeader("Content-Disposition") == "" {
		filename := path.Base(targetURL.Path)
		if filename == "" || filename == "." {
			filename = "download"
		}
		params := map[string]string{"filename": filename}
		disposition := mime.FormatMediaType("attachment", params)
		c.Header("Content-Disposition", disposition)
	}

	c.Status(resp.StatusCode)
	copied, err := io.Copy(c.Writer, resp.Body)
	if err != nil {
		log.Printf("[ProxyHandler] 复制响应内容失败: %v", err)
	} else {
		log.Printf("[ProxyHandler] 成功代理请求: %s, 传输大小: %d bytes", DownloadUrl, copied)
	}
}

func DockerHandler(c *gin.Context) {
	originalURL := c.Param("proxyPath")
	targetURL := fmt.Sprintf("https://%s", originalURL[1:len(originalURL)])
	log.Printf("[DockerHandler] 收到Docker镜像请求: %s, 客户端IP: %s", targetURL, c.ClientIP())

	t, err := url.Parse(targetURL)
	if err != nil {
		log.Printf("[DockerHandler] URL解析失败: %v, URL: %s", err, targetURL)
		c.String(http.StatusBadRequest, "无效的URL格式")
		return
	}

	if in(t.Host, viper.GetStringSlice("whiteList")) == false {
		log.Printf("[DockerHandler] 域名不在白名单中: %s", t.Host)
		c.String(http.StatusForbidden, "目标域名不在白名单中")
		return
	}

	cleanPath := strings.TrimPrefix(targetURL, fmt.Sprintf("https://%s", t.Host))
	proxyURL := fmt.Sprintf("https://%s/v2%s", t.Host, cleanPath)
	proxyURL = strings.ReplaceAll(proxyURL, "docker.io", "registry-1.docker.io")

	req, err := http.NewRequest(c.Request.Method, proxyURL, nil)
	if err != nil {
		log.Printf("[DockerHandler] 创建请求失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败"})
		return
	}

	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[DockerHandler] 请求上游服务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "请求上游服务失败"})
		return
	}
	defer resp.Body.Close()

	// 处理Docker Registry的认证
	if resp.StatusCode == http.StatusUnauthorized {
		authHeader := resp.Header.Get("WWW-Authenticate")
		log.Printf("[DockerHandler] 收到认证头: %s", authHeader)

		if authHeader != "" {
			// 解析认证信息
			realm, service, scope := parseAuthHeader(authHeader)
			// 获取认证token
			token, err := getAuthToken(realm, service, scope)
			log.Printf("[DockerHandler] 获取到的token: %s", token)
			if err != nil {
				log.Printf("[DockerHandler] 获取认证token失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "认证失败"})
				return
			}
			// 使用token重新发送请求
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			resp, err = client.Do(req)
			if err != nil {
				log.Printf("[DockerHandler] 使用token重新请求失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "认证请求失败"})
				return
			}
			defer resp.Body.Close()
		}
	}

	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}

	c.Status(resp.StatusCode)

	if c.Request.Method != http.MethodHead {
		copied, err := io.Copy(c.Writer, resp.Body)
		log.Printf("[DockerHandler] resp.Header.Authorization: %s", resp.Header.Get("Authorization"))
		if err != nil {
			log.Printf("[DockerHandler] 复制响应内容失败: %v", err)
		} else {
			log.Printf("[DockerHandler] 成功代理请求: %s, 传输大小: %d bytes", proxyURL, copied)
		}
	}
}
