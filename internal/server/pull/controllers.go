package pull

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PullRequestBody struct {
	ProjectId string
}

func Pull(c *gin.Context) {
	var requestBody PullRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		fmt.Printf("error!")
	}

	fmt.Println(requestBody.ProjectId)
	c.IndentedJSON(http.StatusOK, requestBody)
}