package service

import (
	"github.com/gin-gonic/gin"
)

func initRouter() *gin.Engine{
	router := gin.Default()
	router.GET("/version", Version)

	return router
}
