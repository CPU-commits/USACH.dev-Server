package controllers

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/services"
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

func (*UserController) GetAvatar(c *gin.Context) {
	username := c.Param("username")

	c.Stream(func(w io.Writer) bool {
		err := usersService.GetAvatar(username, w)
		if err != nil {
			c.AbortWithStatusJSON(err.StatusCode, &res.Response{
				Message: err.Err.Error(),
			})
			return false
		}

		return false
	})
}

func (*UserController) UpdateProfile(c *gin.Context) {
	claims, _ := services.NewClaimsFromContext(c)

	var profile *forms.ProfileForm
	if c.PostForm("description") != "" {
		if err := c.ShouldBind(&profile); err != nil {
			fmt.Printf("err: %v\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
				Message: err.Error(),
			})
			return
		}
	}
	// Form
	avatar, err := c.FormFile("avatar")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			c.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
	}
	// Update profile
	errRes := usersService.UpdateProfile(claims.UserID, profile, avatar)
	if errRes != nil {
		c.AbortWithStatusJSON(errRes.StatusCode, &res.Response{
			Message: errRes.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &res.Response{})
}
