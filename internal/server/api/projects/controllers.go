package projects

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type projectsRequestBody struct {
	UserId string
}

func getProjects(c *gin.Context) {
	var requestBody projectsRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		fmt.Printf("error!")
	}

	fmt.Println(requestBody.UserId)
	c.IndentedJSON(http.StatusOK, requestBody)
}