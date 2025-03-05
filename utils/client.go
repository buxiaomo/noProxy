package utils

import (
	"net"
	"net/http"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

var (
	defaultMaxRetries = 3
	defaultRetryDelay = time.Second
	defaultCacheSize  = 1000

	// 全局HTTP客户端实例
	httpClient *http.Client
	// 全局缓存实例
	cache *lru.Cache[string, []byte]
	// 初始化锁
	initOnce sync.Once
)

// InitHTTPClient 初始化HTTP客户端
func InitHTTPClient() {
	initOnce.Do(func() {
		// 创建自定义传输层
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		// 创建HTTP客户端
		httpClient = &http.Client{
			Transport: transport,
			Timeout:   30 * time.Minute,
		}

		// 初始化缓存
		var err error
		cache, err = lru.New[string, []byte](defaultCacheSize)
		if err != nil {
			panic(err)
		}
	})
}

// GetHTTPClient 获取HTTP客户端实例
func GetHTTPClient() *http.Client {
	if httpClient == nil {
		InitHTTPClient()
	}
	return httpClient
}

// GetCache 获取缓存实例
func GetCache() *lru.Cache[string, []byte] {
	if cache == nil {
		InitHTTPClient()
	}
	return cache
}

// RetryableHTTPGet 执行可重试的HTTP GET请求
func RetryableHTTPGet(url string, maxRetries int) (*http.Response, error) {
	if maxRetries <= 0 {
		maxRetries = defaultMaxRetries
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		resp, err := GetHTTPClient().Get(url)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if i < maxRetries-1 {
			time.Sleep(defaultRetryDelay)
		}
	}

	return nil, lastErr
}
