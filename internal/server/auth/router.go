package auth

import (
	"github.com/gin-gonic/gin"
)

func AuthRouter(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", Login)
		auth.POST("/register", Register)
	}
}