package service

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Version(ctx *gin.Context){
	ctx.JSON(http.StatusOK,gin.H{
		"gomq" : "v0.1",
		"gomqctl" : "v0.1",
	})
}
