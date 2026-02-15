package pull

import (
	"github.com/gin-gonic/gin"
)

func PullRouter(router *gin.RouterGroup) {
	router.POST("/pull", Pull)
}