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
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsIng1YyI6WyJNSUlFRmpDQ0F2NmdBd0lCQWdJVUlvQW42a0k2MUs3bTAwNVhqcXVpKzRDTzVUb3dEUVlKS29aSWh2Y05BUUVMQlFBd2dZWXhDekFKQmdOVkJBWVRBbFZUTVJNd0VRWURWUVFJRXdwRFlXeHBabTl5Ym1saE1SSXdFQVlEVlFRSEV3bFFZV3h2SUVGc2RHOHhGVEFUQmdOVkJBb1RERVJ2WTJ0bGNpd2dTVzVqTGpFVU1CSUdBMVVFQ3hNTFJXNW5hVzVsWlhKcGJtY3hJVEFmQmdOVkJBTVRHRVJ2WTJ0bGNpd2dTVzVqTGlCRmJtY2dVbTl2ZENCRFFUQWVGdzB5TkRBNU1qUXlNalUxTURCYUZ3MHlOVEE1TWpReU1qVTFNREJhTUlHRk1Rc3dDUVlEVlFRR0V3SlZVekVUTUJFR0ExVUVDQk1LUTJGc2FXWnZjbTVwWVRFU01CQUdBMVVFQnhNSlVHRnNieUJCYkhSdk1SVXdFd1lEVlFRS0V3eEViMk5yWlhJc0lFbHVZeTR4RkRBU0JnTlZCQXNUQzBWdVoybHVaV1Z5YVc1bk1TQXdIZ1lEVlFRREV4ZEViMk5yWlhJc0lFbHVZeTRnUlc1bklFcFhWQ0JEUVRDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTGRWRDVxNlJudkdETUxPVysrR1MxWENwR2FRRHd0V3FIS2tLYlM5cVlJMXdCallKWEJ6U2MweTBJK0swU0lVd2pqNGJJT3ZpNXNyOGhJajdReGhrY1ppTlU1OEE5NW5BeGVFS3lMaU9QU0tZK3Y5VnZadmNNT2NwVW1xZ1BxWkhoeTVuMW8xbGxmek92dTd5SDc4a1FyT0lTMTZ3RFVVZm8yRkxPaERDaElsbCtYa2VlbFB6c0tiRWo3ZGJqdXV6RGxIODlWaE4yenNWNFV3c244UVpGVTB4V00wb3R2d0lQN3k0UDZGWDBuUFJuTVQyMlRIajVIWVJ3SUFVdm1FN0l3YlZVQ2wvM1hPaGhwbGNJZFQxREZGOUJUMHJOUC93ZTBWMklId1RHdVdTVENWb3M2b3R5ekk3a1hEdGZzeWRjU2Q5TklpSXZITHFYamJPVGtidWVjQ0F3RUFBYU43TUhrd0RnWURWUjBQQVFIL0JBUURBZ0dtTUJNR0ExVWRKUVFNTUFvR0NDc0dBUVVGQndNQk1CSUdBMVVkRXdFQi93UUlNQVlCQWY4Q0FRQXdIUVlEVlIwT0JCWUVGSmZWdXV4Vko3UXh1amlMNExZajFjQjEzbWhjTUI4R0ExVWRJd1FZTUJhQUZDNjBVUE5lQmtvZ1kyMnRYUGNCTUhGdkczQ3NNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUJBUUFiQkdlZHZVVzhOVWp1VXJWbDlrWmMybDRDbjhJbDFzeVBVTDNYVXdQSHprcy9iUFJ4S1loeFlIODdOb1NwdDZJT3ZPS0k3ZCthQmoyM1lldTdDWGltTWxMUWl4UGhwQ0J0dC92Vmx1UXNJbVZXZXBJWCtraENienNGemtNbUNpbHo1OXVxOURiaGg3VUw1NjQxUjdFQ2pZc0h0Y2RKeURXRWFqMXFEVFoyOUUwY1UvWmhmbmsrVFVOTExkNjYxNldCREQ3TDlSNkgzK3VkRXBRRDFlcXYzU1YwczY3R2ZVT3l0RXhzMVRja3U4aUJCdnJLbnhZa3BZMVZDbW5UMUxSaFo4K283YU94YjR4ZHByMFR6ZnBqN3BidEhWQnQwSGNUUlpIdG54MkhCN3pzWXRFZUl3eVE3bGhhMVJ4eDJNQmR0R2tBREFLUk9RRnpmMEJubm91TSJdfQ.eyJhY2Nlc3MiOltdLCJhdWQiOiIiLCJleHAiOjE3NDExNTY0NjQsImh0dHBzOi8vYXV0aC5kb2NrZXIuaW8iOnsicGxhbl9uYW1lIjoiZnJlZSIsInVzZXJuYW1lIjoiYnV4aWFvbW8ifSwiaWF0IjoxNzQxMTU2MTY0LCJpc3MiOiJhdXRoLmRvY2tlci5pbyIsImp0aSI6ImRja3JfanRpX2JyaUJXZE9NcmlMT3NPNHhmd3lscVVQcnZqRT0iLCJuYmYiOjE3NDExNTU4NjQsInN1YiI6IjM3MjE0ZjE0LTYyMTYtNDEyMi05YTFiLWNjMTE1ZGZhZGVhYiJ9.miMNT8kniWsp-1sb1yA9iwmpuGJywaL_DQSG_4BJNZFbs84bBfkByLjzO6ibpIpsaCgJaegAucETvz-2zmv9sIq73gkY3jyJYUOYBRcQbINxy4B7PKuMMe9G1YCbViUuy1JVy9RNl_ELKTtKLS-4PHrKUEn0YMi3D7o06eSi9KbWzUJsE5bGIJWEQt6LnHAGsTS8c-Y-_kOoyxwzN6KJe4o0YAGbxveMhZMiOZ3ZdYsQScqKuqam9LfsbGVeuZ-fMEOnzikNUxErJ5gt5DoTrQZdtC0BcJBnnD21V32kea6GPLEOhepmGk6TnLoiF2DK5yVaKapcpvGCoQpc5yJ48g")
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

	var Authorization string
	// 处理Docker Registry的认证
	if resp.StatusCode == http.StatusUnauthorized {
		authHeader := resp.Header.Get("WWW-Authenticate")
		log.Printf("[DockerHandler] 收到认证头: %s", authHeader)

		if authHeader != "" {
			// 解析认证信息
			realm, service, scope := parseAuthHeader(authHeader)
			// 获取认证token
			Authorization, err := getAuthToken(realm, service, scope)
			log.Printf("[DockerHandler] 获取到的token: %s", Authorization)

			if err != nil {
				log.Printf("[DockerHandler] 获取认证token失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "认证失败"})
				return
			}
			// 使用token重新发送请求
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", Authorization))
			resp, err = client.Do(req)
			if err != nil {
				log.Printf("[DockerHandler] 使用token重新请求失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "认证请求失败"})
				return
			}
			defer resp.Body.Close()
		}
	}
	resp.Header.Set("Authorization", fmt.Sprintf("Bearer %s", Authorization))
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
