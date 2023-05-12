package controllers

import (
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/services"
	"github.com/gin-gonic/gin"
)

type DiscussionController struct{}

func (d *DiscussionController) UploadDiscussion(c *gin.Context) {
	var discussion *forms.DiscussionForm

	if err := c.ShouldBind(&discussion); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Message: err.Error(),
		})
		return
	}
	// Extract file
	var image *multipart.FileHeader

	var err error
	image, err = c.FormFile("image")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Message: "Error con archivo",
		})
		return
	}
	// JWT
	claims, _ := services.NewClaimsFromContext(c)

	errRes := discussionService.UploadDiscussion(discussion, claims.UserID, image)
	if errRes != nil {
		c.AbortWithStatusJSON(errRes.StatusCode, &res.Response{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, &res.Response{})
}
