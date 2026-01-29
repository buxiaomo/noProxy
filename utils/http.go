package utils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

func IsHopHeader(header string) bool {
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

func InWhiteList(target string, strArray []string) bool {
	if len(strArray) == 0 {
		log.Printf("[InWhiteList] 白名单为空，目标域名: %s", target)
		return false
	}

	log.Printf("[InWhiteList] 开始检查域名 %s 是否在白名单中", target)

	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, target)
	if index < len(strArray) && strArray[index] == target {
		log.Printf("[InWhiteList] 域名 %s 精确匹配白名单", target)
		return true
	}

	targetParts := strings.Split(target, ".")
	for _, whiteDomain := range strArray {
		if strings.HasPrefix(whiteDomain, "*.") {
			rootDomain := whiteDomain[2:]
			if strings.HasSuffix(target, "."+rootDomain) {
				log.Printf("[InWhiteList] 域名 %s 通配符匹配 %s", target, whiteDomain)
				return true
			}
			continue
		}

		whiteParts := strings.Split(whiteDomain, ".")
		if len(whiteParts) <= len(targetParts) {
			matched := true
			for i := 1; i <= len(whiteParts); i++ {
				if targetParts[len(targetParts)-i] != whiteParts[len(whiteParts)-i] {
					matched = false
					break
				}
			}
			if matched {
				log.Printf("[InWhiteList] 域名 %s 是 %s 的子域名", target, whiteDomain)
				return true
			}
		}
	}

	log.Printf("[InWhiteList] 域名 %s 不在白名单中", target)
	return false
}

type ipAPIResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	CountryCode string `json:"countryCode"`
}

var (
	ipRegionCache     *lru.Cache[string, bool]
	ipRegionCacheOnce sync.Once
)

func initIPRegionCache() {
	ipRegionCacheOnce.Do(func() {
		cache, err := lru.New[string, bool](10000)
		if err != nil {
			log.Printf("[IsChinaIP] 初始化IP缓存失败: %v", err)
			return
		}
		ipRegionCache = cache
	})
}

func IsChinaIP(ip string) bool {
	if ip == "" {
		return true
	}

	initIPRegionCache()
	if ipRegionCache != nil {
		if v, ok := ipRegionCache.Get(ip); ok {
			return v
		}
	}

	isChina := queryIPIsChina(ip)
	if ipRegionCache != nil {
		ipRegionCache.Add(ip, isChina)
	}
	return isChina
}

func queryIPIsChina(ip string) bool {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	url := "http://ip-api.com/json/" + ip + "?fields=status,message,countryCode"
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("[IsChinaIP] 查询IP归属地失败: %v", err)
		return true
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[IsChinaIP] 读取IP归属地响应失败: %v", err)
		return true
	}

	var data ipAPIResponse
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("[IsChinaIP] 解析IP归属地响应失败: %v", err)
		return true
	}

	if data.Status != "success" {
		log.Printf("[IsChinaIP] IP归属地查询失败: %s", data.Message)
		return true
	}

	return strings.EqualFold(data.CountryCode, "CN")
}
