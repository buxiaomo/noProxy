package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

func ProxyHandler(c *gin.Context) {
	DownloadUrl := c.Param("url")
	if DownloadUrl == "/d/" {
		c.String(http.StatusOK, "Hello World.")
		return
	}
	DownloadUrl = DownloadUrl[1:len(DownloadUrl)]
	log.Println("DownloadUrl:", DownloadUrl)
	targetURL, err := url.Parse(DownloadUrl)
	if err != nil {
		c.String(http.StatusOK, "Parse url failed.")
		return
	}

	if in(targetURL.Host, viper.GetStringSlice("whiteList")) == false {
		c.String(http.StatusOK, "target domain name is not in whiteList.")
		return
	}

	if targetURL.Scheme != "http" && targetURL.Scheme != "https" {
		c.String(http.StatusBadRequest, "URL must be http or https")
		return
	}

	client := &http.Client{
		Timeout: 30 * time.Minute, // 长超时以支持大文件下载
	}
	resp, err := client.Get(DownloadUrl)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch URL: %v", err))
		return
	}
	defer resp.Body.Close()

	// 过滤并复制响应头
	for k, vv := range resp.Header {
		if isHopHeader(k) {
			continue
		}
		for _, v := range vv {
			c.Header(k, v)
			//w.Header().Add(k, v)
		}
	}
	// 设置默认的Content-Disposition
	if c.GetHeader("Content-Disposition") == "" {
		//if w.Header().Get("Content-Disposition") == "" {
		filename := path.Base(targetURL.Path)
		if filename == "" || filename == "." {
			filename = "download"
		}
		params := map[string]string{"filename": filename}
		disposition := mime.FormatMediaType("attachment", params)
		c.Header("Content-Disposition", disposition)
		//w.Header().Set("Content-Disposition", disposition)
	}
	c.Status(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

func DockerHandler(c *gin.Context) {
	originalURL := c.Param("proxyPath")
	//log.Printf("Received request for: %s", originalURL)
	targetURL := fmt.Sprintf("https://%s", originalURL[1:len(originalURL)])
	t, err := url.Parse(targetURL)
	if err != nil {
		//log.Printf("Error parsing URL: %s", targetURL)
		c.String(http.StatusOK, err.Error())
		return
	}

	//log.Printf("target url: %s", targetURL)

	// 解析路径，去除 "/k8s.gcr.io" 前缀，使其符合 Docker Registry 规范
	cleanPath := strings.TrimPrefix(targetURL, fmt.Sprintf("https://%s", t.Host))

	proxyURL := fmt.Sprintf("https://%s/v2%s", t.Host, cleanPath)
	proxyURL = strings.ReplaceAll(proxyURL, "docker.io", "registry-1.docker.io")

	//log.Printf("Proxying request to: %s", proxyURL)
	// 创建 HTTP 请求，支持 HEAD/GET 方法
	req, err := http.NewRequest(c.Request.Method, proxyURL, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// 复制原请求的 Header
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching from upstream: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch from upstream"})
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}

	// 设置状态码
	c.Status(resp.StatusCode)

	// HEAD 请求不需要返回 body
	if c.Request.Method != http.MethodHead {
		io.Copy(c.Writer, resp.Body)
	}

	//log.Printf("Successfully proxied request for: %s", originalURL)
}
