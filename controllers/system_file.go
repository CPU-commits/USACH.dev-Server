package controllers

import (
	"mime/multipart"
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/middlewares"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/services"
	"github.com/gin-gonic/gin"
)

type SystemFileController struct{}

func (s *SystemFileController) GetFolder(c *gin.Context) {
	// Route
	username := c.Param("username")
	repositoryName := c.Param("repository")
	idFolder := c.Param("folder")
	// User
	idUserREQ, exists := c.Get(middlewares.USER_ID_KEY)
	if !exists {
		c.AbortWithStatusJSON(http.StatusInternalServerError, &res.Response{
			Message: "unavailable",
		})
		return
	}

	// Get folder
	folder, repository, err := systemFileService.GetFolder(
		username,
		repositoryName,
		idFolder,
		idUserREQ.(string),
	)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &res.Response{
		Data: map[string]interface{}{
			"folder":     folder,
			"repository": repository,
		},
	})
}

func (s *SystemFileController) NewRepoElement(c *gin.Context) {
	// Route
	repository := c.Param("idRepository")
	parent := c.DefaultQuery("parent", "")
	// JWT
	claims, _ := services.NewClaimsFromContext(c)

	var element *forms.SystemFileForm
	if err := c.ShouldBind(&element); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Message: "Bad request",
		})
		return
	}
	var file *multipart.FileHeader
	if !*element.IsDirectory {
		// Extract file
		var err error
		file, err = c.FormFile("file")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
				Message: "Error con archivo",
			})
			return
		}

		if file == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
				Message: "No ha envíado ningún archivo",
			})
			return
		}
	} else {
		file = nil
	}
	// Upload element
	response, err := systemFileService.NewRepoElement(
		element,
		repository,
		parent,
		claims.UserID,
		file,
	)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, &res.Response{
		Data: response,
	})
}

func (s *SystemFileController) DeleteElement(c *gin.Context) {
	repository := c.Param("repository")
	element := c.Param("element")

	claims, _ := services.NewClaimsFromContext(c)
	// Delete element
	err := systemFileService.DeleteElement(repository, element, claims.UserID)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &res.Response{})
}
