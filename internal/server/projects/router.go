package projects

import (
	"github.com/gin-gonic/gin"
)

func ProjectsRouter(router *gin.RouterGroup) {
	router.GET("/projects", getProjects)
}