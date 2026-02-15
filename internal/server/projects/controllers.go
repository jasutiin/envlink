package projects

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProjectsRequestBody struct {
	UserId string
}

func Projects(c *gin.Context) {
	var requestBody ProjectsRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		fmt.Printf("error!")
	}

	fmt.Println(requestBody.UserId)
	c.IndentedJSON(http.StatusOK, requestBody)
}