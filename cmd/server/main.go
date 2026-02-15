package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/login", login)
	router.GET("/register", register)
	router.GET("/store", store)
	router.GET("/push", push)
	router.GET("/pull", pull)
	router.GET("/projects", projects)
	router.Run("localhost:8080")
};

type AuthRequestBody struct {
	Email string
	Password string
}

func login(c *gin.Context) {
	var requestBody AuthRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		fmt.Printf("error!")
	}

	fmt.Println(requestBody.Email)
	fmt.Println(requestBody.Password)
	c.IndentedJSON(http.StatusOK, requestBody)
}

func register(c *gin.Context) {
	var requestBody AuthRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		fmt.Printf("error!")
	}

	fmt.Println(requestBody.Email)
	fmt.Println(requestBody.Password)
	c.IndentedJSON(http.StatusOK, requestBody)
}

func store(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "store")
}

func push(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "push")
}

func pull(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "pull")
}

func projects(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "projects")
}