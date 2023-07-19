package controllers

import (
	"fmt"
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/gin-gonic/gin"
)

type AuthController struct{}

func (a *AuthController) CreateUser(c *gin.Context) {
	var user *forms.UserForm
	if err := c.BindJSON(&user); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Message: err.Error(),
		})
		return
	}
	// Create User
	err := usersService.CreateUser(user)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, &res.Response{})
}

func (a *AuthController) ConfirmUser(c *gin.Context) {
	tokenValue := c.DefaultQuery("token", "")
	// Validate token and activate user
	err := usersService.ActivateUser(tokenValue)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}
	// Redirect
	protocol := "http"
	if settingsData.GO_ENV == "prod" {
		protocol += "s"
	}

	c.Redirect(
		http.StatusPermanentRedirect,
		fmt.Sprintf("%s://%s/session?confirm=true", protocol, settingsData.CLIENT_URL),
	)
}

func (a *AuthController) Login(c *gin.Context) {
	var loginForm *forms.LoginForm

	if err := c.BindJSON(&loginForm); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Message: err.Error(),
		})
		return
	}
	// Login
	response, err := authService.Login(loginForm)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &res.Response{
		Data: response,
	})
}

func (a *AuthController) RefreshToken(c *gin.Context) {
	accessToken, err := authService.RefreshToken(c.Request)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &res.Response{
		Data: map[string]interface{}{
			"token": accessToken,
		},
	})
}
