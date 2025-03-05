package utils

import (
	"log"
	"sort"
	"strings"
)

// IsHopHeader 检查是否为逐跳头部字段
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

// InWhiteList 检查目标域名是否在白名单中
func InWhiteList(target string, strArray []string) bool {
	if len(strArray) == 0 {
		log.Printf("[InWhiteList] 白名单为空，目标域名: %s", target)
		return false
	}

	log.Printf("[InWhiteList] 开始检查域名 %s 是否在白名单中", target)
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, target)
	result := index < len(strArray) && strArray[index] == target
	log.Printf("[InWhiteList] 域名 %s 白名单检查结果: %v", target, result)
	return result
}