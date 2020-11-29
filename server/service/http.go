package service

import "github.com/gin-gonic/gin"

type HTTP struct {
	protocol string
	engine   *gin.Engine
}

func NewHTTP() *HTTP {
	return &HTTP{
		protocol: "http",
		engine:   initRouter(),
	}
}



func (h *HTTP)Start(){
	h.engine.Run(":8000")
}