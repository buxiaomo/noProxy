package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"
)

var config string

func init() {
	flag.StringVar(&config, "config", "./noProxy.yaml", "configuration file path.")
	flag.Parse()

	viper.SetConfigFile(config)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func main() {
	r := gin.Default()
	r.GET("/*url", func(c *gin.Context) {
		DownloadUrl := c.Param("url")
		if DownloadUrl == "/" {
			c.String(http.StatusOK, "Hello World.")
			return
		}
		DownloadUrl = DownloadUrl[1:len(DownloadUrl)]
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
	})

	if viper.GetString("domainName") != "" {
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(viper.GetString("domainName")),
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

func in(target string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, target)
	//index的取值：[0,len(str_array)]
	if index < len(strArray) && strArray[index] == target { //需要注意此处的判断，先判断 &&左侧的条件，如果不满足则结束此处判断，不会再进行右侧的判断
		return true
	}
	return false
}
