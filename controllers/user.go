package controllers

import (
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/gin-gonic/gin"
)

type UserController struct{}

func (u *UserController) GetUser(c *gin.Context) {
	// Params
	profileQuery := c.DefaultQuery("profile", "false")

	profile := profileQuery == "true"
	username := c.Param("idUser")

	user, err := usersService.GetUser(username, profile)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &res.Response{
		Data: map[string]interface{}{
			"user": user,
		},
	})
}
