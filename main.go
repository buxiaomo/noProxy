package main

import (
	"flag"
	"fmt"
	"log"
	"noProxy/handler"
	"time"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
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

	// 添加健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	r.Any("/v2/*proxyPath", handler.DockerHandler)
	r.GET("/d/*url", handler.ProxyHandler)

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
