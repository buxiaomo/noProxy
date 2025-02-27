package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

func main() {
	domain := flag.String("domain", "", "domain name")
	flag.Parse()

	r := gin.Default()
	r.GET("/*url", func(c *gin.Context) {
		DownloadUrl := c.Param("url")
		DownloadUrl = DownloadUrl[1:len(DownloadUrl)]
		targetURL, err := url.Parse(DownloadUrl)
		if err != nil {
			c.String(http.StatusOK, "Parse url failed.")
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
	})

	if domain != nil {
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(*domain),
			Cache:      autocert.DirCache("/data/.cache"),
		}

		log.Fatal(autotls.RunWithManager(r, &m))
	} else {
		r.Run()
	}
}

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
