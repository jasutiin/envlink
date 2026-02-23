package pull

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type pullRequestBody struct {
	ProjectId string
}

func postPull(c *gin.Context) {
	var requestBody pullRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		fmt.Printf("error!")
	}

	fmt.Println(requestBody.ProjectId)
	c.IndentedJSON(http.StatusOK, requestBody)
}