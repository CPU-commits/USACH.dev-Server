package controllers

import (
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/services"
	"github.com/gin-gonic/gin"
)

type CommentController struct{}

func (*CommentController) GetComments(c *gin.Context) {
	idDiscussion := c.Param("discussion")

	claims, _ := services.NewClaimsFromContext(c)
	// Get discussion
	comments, err := commentService.GetComments(idDiscussion, claims.UserID)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res.Response{
		Data: map[string]interface{}{
			"comments": comments,
		},
	})
}

func (*CommentController) Comment(c *gin.Context) {
	idDiscussion := c.Param("discussion")
	replyComment := c.DefaultQuery("reply", "")

	var comment *forms.CommentForm
	if err := c.BindJSON(&comment); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Message: err.Error(),
		})
		return
	}

	claims, _ := services.NewClaimsFromContext(c)
	// Upload
	insertedId, err := commentService.Comment(comment.Comment, idDiscussion, claims.UserID, replyComment)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, &res.Response{
		Data: map[string]interface{}{
			"inserted_id": insertedId,
		},
	})
}
