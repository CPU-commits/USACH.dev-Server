package controllers

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/services"
	"github.com/gin-gonic/gin"
)

type DiscussionController struct{}

func (d *DiscussionController) GetDiscussion(c *gin.Context) {
	// Params
	idDiscussion := c.Param("discussion")

	claims, _ := services.NewClaimsFromContext(c)
	// Discussion
	discussion, err := discussionService.GetDiscussion(
		idDiscussion,
		claims.UserID,
	)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &res.Response{
		Data: map[string]interface{}{
			"discussion": discussion,
		},
	})
}

func (d *DiscussionController) GetDiscussions(c *gin.Context) {
	// Query params
	page := c.DefaultQuery("page", "0")
	sortBy := c.DefaultQuery("sortBy", "created_at")
	repository := c.DefaultQuery("repository", "")
	getTotal := c.DefaultQuery("count", "false")
	search := c.DefaultQuery("search", "")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Message: err.Error(),
		})
		return
	}
	// User
	claims, _ := services.NewClaimsFromContext(c)
	// Get discussions
	discussions, totalData, errRes := discussionService.GetDiscussions(
		pageNumber,
		sortBy,
		repository,
		claims.UserID,
		search,
		getTotal == "true",
	)
	if errRes != nil {
		c.AbortWithStatusJSON(errRes.StatusCode, &res.Response{
			Message: errRes.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &res.Response{
		Data: map[string]interface{}{
			"discussions": discussions,
			"total":       totalData,
		},
	})
}

func (*DiscussionController) HasAccess(c *gin.Context) {
	idDiscussion := c.Param("discussion")
	claims, _ := services.NewClaimsFromContext(c)
	// Get access
	err := discussionService.HasAccess(
		idDiscussion,
		claims.UserID,
	)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &res.Response{})
}

func (d *DiscussionController) GetImage(c *gin.Context) {
	discussion := c.Param("discussion")
	image := c.Param("image")

	c.Stream(func(w io.Writer) bool {
		err := discussionService.GetImage(
			discussion,
			image,
			w,
		)
		if err != nil {
			c.AbortWithStatusJSON(err.StatusCode, &res.Response{
				Message: err.Err.Error(),
			})
			return false
		}

		return false
	})
}

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
			Message: errRes.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, &res.Response{})
}

func (d *DiscussionController) ReactDiscussion(c *gin.Context) {
	var reaction *forms.ReactionForm
	if err := c.BindJSON(&reaction); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Message: err.Error(),
		})
		return
	}

	idDiscussion := c.Param("discussion")

	claims, _ := services.NewClaimsFromContext(c)
	// React
	err := discussionService.React(
		idDiscussion,
		claims.UserID,
		reaction.Reaction,
	)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, &res.Response{})
}

func (d *DiscussionController) DeleteReaction(c *gin.Context) {
	idDiscussion := c.Param("discussion")

	claims, _ := services.NewClaimsFromContext(c)
	// Delete reaction
	err := discussionService.DeleteReaction(idDiscussion, claims.UserID)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &res.Response{})
}
