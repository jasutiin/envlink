package auth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type authRequestBody struct {
	Email string
	Password string
}

func postLogin(c *gin.Context) {
	var requestBody authRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		fmt.Printf("error!")
	}

	fmt.Println(requestBody.Email)
	fmt.Println(requestBody.Password)
	c.IndentedJSON(http.StatusOK, requestBody)
}

func postRegister(c *gin.Context) {
	var requestBody authRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		fmt.Printf("error!")
	}

	fmt.Println(requestBody.Email)
	fmt.Println(requestBody.Password)
	c.IndentedJSON(http.StatusOK, requestBody)
}