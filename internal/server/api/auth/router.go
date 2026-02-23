package auth

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthRouter(router *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	controller := NewController(repo)

	auth := router.Group("/auth")
	{
		auth.POST("/login", controller.postLogin)
		auth.POST("/register", controller.postRegister)
		auth.GET("/:provider", controller.getAuthProvider)
		auth.GET("/:provider/callback", controller.getAuthCallbackFunction)
		auth.POST("/cli/exchange", controller.postCLIExchange)
		auth.GET("/:provider/logout", controller.getLogoutProvider)
	}
}