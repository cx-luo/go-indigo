// Package main provides a REST API service for go-indigo chemistry toolkit
// Mimics the EPAM Indigo Python Service API using Gin framework
// coding=utf-8
// @Project : go-indigo
// @Time    : 2026/03/16
// @Author  : chengxiang.luo
// @Email   : chengxiang.luo@foxmail.com
// @File    : main.go
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/cx-luo/go-indigo/service/docs"
)

//	@title						Go-Indigo Service API
//	@version					1.0.0
//	@description				RESTful API service for go-indigo chemistry toolkit. Compatible with EPAM Indigo Service API.
//	@description				Provides endpoints for aromatization, dearomatization, format conversion, property calculation, 2D cleanup, structure rendering, and validation.
//
//	@contact.name				chengxiang.luo
//	@contact.email				chengxiang.luo@foxmail.com
//
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//
//	@host						localhost:8080
//	@BasePath					/
//	@schemes					http https
//
//	@externalDocs.description	EPAM Indigo Service Documentation
//	@externalDocs.url			https://lifescience.opensource.epam.com/indigo/service/index.html

func main() {
	port := flag.Int("port", 8080, "server port")
	poolSize := flag.Int("pool", 4, "indigo session pool size")
	mode := flag.String("mode", "release", "gin mode: debug, release, test")
	flag.Parse()

	gin.SetMode(*mode)

	svc := NewIndigoService(*poolSize)
	r := setupRouter(svc)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("go-indigo service starting on %s (pool=%d, mode=%s)", addr, *poolSize, *mode)
	log.Printf("Swagger UI: http://localhost:%d/swagger/index.html", *port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func setupRouter(svc *IndigoService) *gin.Engine {
	r := gin.Default()

	r.Use(corsMiddleware())

	v2 := r.Group("/v2/indigo")
	{
		v2.GET("/info", svc.handleInfo)
		v2.POST("/aromatize", svc.handleAromatize)
		v2.POST("/dearomatize", svc.handleDearomatize)
		v2.POST("/calculate", svc.handleCalculate)
		v2.POST("/convert", svc.handleConvert)
		v2.POST("/clean", svc.handleClean)
		v2.POST("/render", svc.handleRender)
		v2.POST("/check", svc.handleCheck)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
