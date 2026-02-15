package push

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PushRequestBody struct {
	ProjectId string
	Content string
}

func Push(c *gin.Context) {
	var requestBody PushRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		fmt.Printf("error!")
	}

	fmt.Println(requestBody.ProjectId)
	fmt.Println(requestBody.Content)
	c.IndentedJSON(http.StatusOK, requestBody)
}