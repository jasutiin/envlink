package push

import (
	"github.com/gin-gonic/gin"
)

func PushRouter(router *gin.RouterGroup) {
	router.POST("/push", postPush)
}