package push

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type pushRequestBody struct {
	ProjectId string
	Content string
}

func postPush(c *gin.Context) {
	var requestBody pushRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		fmt.Printf("error!")
	}

	fmt.Println(requestBody.ProjectId)
	fmt.Println(requestBody.Content)
	c.IndentedJSON(http.StatusOK, requestBody)
}