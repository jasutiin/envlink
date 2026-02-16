package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
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

func (controller *AuthController) getAuthCallbackFunction(c *gin.Context) {
	provider := c.Param("provider")
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "provider", provider))

	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		fmt.Fprintln(c.Writer, c.Request)
	}

	fmt.Println(user)
}

func (controller *AuthController) getAuthProvider(c *gin.Context) {
	provider := c.Param("provider")
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "provider", provider))

	if _, err := gothic.CompleteUserAuth(c.Writer, c.Request); err == nil {
		token := "token" // this is supposed to be a JWT
		c.JSON(http.StatusOK, gin.H{ "token": token })
	} else {
		gothic.BeginAuthHandler(c.Writer, c.Request)
	}
}

func (controller *AuthController) getLogoutProvider(c *gin.Context) {
	gothic.Logout(c.Writer, c.Request)
	c.Writer.Header().Set("Location", "/")
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)
}