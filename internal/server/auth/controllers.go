package auth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthRequestBody struct {
	Email string
	Password string
}

func Login(c *gin.Context) {
	var requestBody AuthRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		fmt.Printf("error!")
	}

	fmt.Println(requestBody.Email)
	fmt.Println(requestBody.Password)
	c.IndentedJSON(http.StatusOK, requestBody)
}

func Register(c *gin.Context) {
	var requestBody AuthRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		fmt.Printf("error!")
	}

	fmt.Println(requestBody.Email)
	fmt.Println(requestBody.Password)
	c.IndentedJSON(http.StatusOK, requestBody)
}