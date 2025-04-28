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
// 支持精确匹配、通配符匹配（*.example.com）和子域名匹配
func InWhiteList(target string, strArray []string) bool {
	if len(strArray) == 0 {
		log.Printf("[InWhiteList] 白名单为空，目标域名: %s", target)
		return false
	}

	log.Printf("[InWhiteList] 开始检查域名 %s 是否在白名单中", target)
	
	// 1. 精确匹配检查
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, target)
	if index < len(strArray) && strArray[index] == target {
		log.Printf("[InWhiteList] 域名 %s 精确匹配白名单", target)
		return true
	}
	
	// 2. 子域名和通配符匹配
	targetParts := strings.Split(target, ".")
	for _, whiteDomain := range strArray {
		// 处理通配符情况（*.example.com）
		if strings.HasPrefix(whiteDomain, "*.") {
			rootDomain := whiteDomain[2:] // 去掉"*."前缀
			if strings.HasSuffix(target, "."+rootDomain) {
				log.Printf("[InWhiteList] 域名 %s 通配符匹配 %s", target, whiteDomain)
				return true
			}
			continue
		}
		
		// 检查是否为子域名
		whiteParts := strings.Split(whiteDomain, ".")
		if len(whiteParts) <= len(targetParts) {
			matched := true
			// 从后向前匹配，确保是子域名关系
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