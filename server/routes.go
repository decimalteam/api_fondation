package server

import (
	"bitbucket.org/decimalteam/api_fondation/logger"
	"github.com/gin-gonic/gin"
)

type Handler struct {
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.Use(logger.LogGinMiddleware())
	return router
}
