package auth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type authRequestBody struct {
	Email    string
	Password string
}

type AuthController struct {
	repo AuthRepository
}

func NewController(repo AuthRepository) *AuthController {
	return &AuthController{repo: repo}
}

func (controller *AuthController) postLogin(c *gin.Context) {
	var requestBody authRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if _, err := controller.repo.Login(requestBody.Email, requestBody.Password); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	fmt.Println(requestBody.Email)
	fmt.Println(requestBody.Password)
	c.IndentedJSON(http.StatusOK, requestBody)
}

func (controller *AuthController) postRegister(c *gin.Context) {
	var requestBody authRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if _, err := controller.repo.Register(requestBody.Email, requestBody.Password); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
		return
	}

	fmt.Println(requestBody.Email)
	fmt.Println(requestBody.Password)
	c.IndentedJSON(http.StatusOK, requestBody)
}