package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

type authRequestBody struct {
	Email    string
	Password string
}

type cliTokenExchangeRequest struct {
	ExchangeCode string `json:"exchange_code"`
	State        string `json:"state"`
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

/*
This function is called when we navigate to the '/api/v1/auth/:provider' endpoint on the server, initiating the auth process
for a provider. It takes a state and callback url passed in as query parameters, and stores them on the server so that we
can use it later to navigate the caller browser back to this url. Finally, it takes the user to the provider's login screen.
*/
func (controller *AuthController) getAuthProvider(c *gin.Context) {
	provider := c.Param("provider")
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "provider", provider))

	cliCallback := strings.TrimSpace(c.Query("cli_callback"))
	cliState := strings.TrimSpace(c.Query("cli_state"))

	if cliCallback != "" || cliState != "" {
		if cliCallback == "" || cliState == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cli_callback and cli_state are required together"})
			return
		}

		if !isAllowedCLICallback(cliCallback) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cli_callback"})
			return
		}

		// create httpOnly cookies to use on subsequent requests
		writeCLIAuthContext(c, cliCallback, cliState)
	}

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

/*
This function is called when we navigate to the '/api/v1/auth/:provider/callback' endpoint on the server. It is called when the user
successfully authenticates. It takes the user's credentials and validates it. It generates an exchange code, then it takes the
callbackURL and state that we stored on the server and builds the url with the exchange code. It saves all of this information on the
server we can do another validation.
*/
func (controller *AuthController) getAuthCallbackFunction(c *gin.Context) {
	provider := c.Param("provider")
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "provider", provider))

	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<h1>Authentication failed</h1><p>Please return to your CLI and try again.</p>"))
		return
	}

	if callbackURL, callbackState, found := readCLIAuthContext(c); found {
		exchangeCode, codeErr := newExchangeCode()
		if codeErr == nil {
			// save all of this information on the server so we can refer back to it later
			pendingCLIExchanges.Save(exchangeCode, callbackState, user.AccessToken, cliExchangeTTL)
			if redirectURL, redirectErr := buildCLIRedirectURL(callbackURL, exchangeCode, callbackState); redirectErr == nil {
				clearCLIAuthContext(c)
				c.Redirect(http.StatusFound, redirectURL)
				return
			}
		}
	}

	clearCLIAuthContext(c)

	html := fmt.Sprintf(`
		<!doctype html>
		<html>
		<head><meta charset="utf-8"><title>Auth successful</title></head>
		<body>
			<h1>Authentication successful</h1>
			<p><strong>Name:</strong> %s</p>
			<p><strong>Email:</strong> %s</p>
			<p><strong>Provider:</strong> %s</p>
			<p>You can close this window and return to the CLI.</p>
		</body>
		</html>
	`, user.Name, user.Email, user.Provider)

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

/*
This function is called when we navigate to the '/api/v1/auth/cli/exchange' endpoint on the server. It is called when the browser
received the exchange code from the server, and sends it back to the server for one last validation check. If the exchange code
that the browser sent matches the one on the server, then we return an authentication token.
*/
func (controller *AuthController) postCLIExchange(c *gin.Context) {
	var requestBody cliTokenExchangeRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	exchangeCode := strings.TrimSpace(requestBody.ExchangeCode)
	state := strings.TrimSpace(requestBody.State)
	if exchangeCode == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exchange_code and state are required"})
		return
	}

	// check if the exchange code and state is stored in the server
	token, found := pendingCLIExchanges.Consume(exchangeCode, state)
	if !found {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired exchange_code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (controller *AuthController) getLogoutProvider(c *gin.Context) {
	gothic.Logout(c.Writer, c.Request)
	c.Writer.Header().Set("Location", "/")
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)
}