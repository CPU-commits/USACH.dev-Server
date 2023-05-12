package controllers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/middlewares"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/services"
	"github.com/gin-gonic/gin"
)

type RepositoryController struct{}

func (r *RepositoryController) GetRepository(c *gin.Context) {
	username := c.Param("username")
	repositoryName := c.Param("repository")

	idUserREQ, exists := c.Get(middlewares.USER_ID_KEY)
	if !exists {
		c.AbortWithStatusJSON(http.StatusInternalServerError, &res.Response{
			Message: "unavailable",
		})
		return
	}
	// Get repository
	repository, like, err := repoService.GetRepository(
		username,
		repositoryName,
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
			"repository": repository,
			"like":       like,
		},
	})
}

func (r *RepositoryController) GetRepositories(c *gin.Context) {
	// Page has 20 elements
	page := c.DefaultQuery("page", "0")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Message: "page must be a int",
		})
		return
	}
	// Return total of elements?
	total := c.DefaultQuery("total", "false")

	search := c.DefaultQuery("search", "")

	claims, _ := services.NewClaimsFromContext(c)

	repositories, totalElements, errRes := repoService.GetRepositories(
		claims.UserID,
		search,
		pageNumber,
		total == "true",
	)
	if errRes != nil {
		c.AbortWithStatusJSON(errRes.StatusCode, &res.Response{
			Message: errRes.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &res.Response{
		Data: map[string]interface{}{
			"repositories": repositories,
			"total":        totalElements,
		},
	})
}

func (r *RepositoryController) GetUserRepositories(c *gin.Context) {
	username := c.Param("username")
	claims, _ := services.NewClaimsFromContext(c)

	repositories, err := repoService.GetUserRepositories(username, claims.UserID)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &res.Response{
		Data: map[string]interface{}{
			"repositories": repositories,
		},
	})
}

func (r *RepositoryController) DownloadRepository(c *gin.Context) {
	repository := c.Param("repository")
	child := c.DefaultQuery("child", "")

	c.Stream(func(w io.Writer) bool {
		var err *res.ErrorRes
		var fileName string

		fileName, err = repoService.DownloadRepository(repository, child, w)
		if err != nil {
			c.AbortWithStatusJSON(err.StatusCode, &res.Response{
				Message: err.Err.Error(),
			})
			return false
		}
		c.Writer.Header().Set(
			"Content-Type",
			"application/octet-stream",
		)
		c.Writer.Header().Set(
			"Content-Disposition",
			fmt.Sprintf("attachment; filename='%s.zip'", fileName),
		)

		return false
	})
}

func (r *RepositoryController) UploadRepository(c *gin.Context) {
	var repository *forms.RepositoryForm
	if err := c.BindJSON(&repository); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Message: err.Error(),
		})
		return
	}
	claims, _ := services.NewClaimsFromContext(c)
	err := repoService.UploadRepository(repository, claims.UserID)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, &res.Response{})
}

func (r *RepositoryController) ToggleLike(c *gin.Context) {
	idRepository := c.Param("repository")

	var plus *forms.LikeForm
	if err := c.BindJSON(&plus); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Message: err.Error(),
		})
		return
	}
	claims, _ := services.NewClaimsFromContext(c)

	err := repoService.ToggleLike(claims.UserID, idRepository, *plus.Plus)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &res.Response{})
}

func (r *RepositoryController) UpdateRepository(c *gin.Context) {
	idRepository := c.Param("repository")

	var repository *forms.UpdateRepositoryForm
	if err := c.BindJSON(&repository); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Message: err.Error(),
		})
		return
	}
	claims, _ := services.NewClaimsFromContext(c)
	err := repoService.UpdateRepository(idRepository, claims.UserID, repository)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, &res.Response{})
}

func (r *RepositoryController) AddLink(c *gin.Context) {
	idRepository := c.Param("repository")

	var link *forms.LinkForm
	if err := c.BindJSON(&link); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Message: err.Error(),
		})
		return
	}

	claims, _ := services.NewClaimsFromContext(c)
	// Create link
	idLink, err := linkService.AddLink(idRepository, claims.UserID, link)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, &res.Response{
		Data: map[string]interface{}{
			"id_link": idLink.Hex(),
		},
	})
}

func (r *RepositoryController) DeleteLink(c *gin.Context) {
	idRepository := c.Param("repository")
	idLink := c.Param("link")

	claims, _ := services.NewClaimsFromContext(c)
	// Delete link
	err := linkService.DeleteLink(idRepository, idLink, claims.UserID)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &res.Response{})
}

func (r *RepositoryController) DeleteRepository(c *gin.Context) {
	idRepository := c.Param("repository")

	claims, _ := services.NewClaimsFromContext(c)
	// Delete repository
	err := repoService.DeleteRepository(idRepository, claims.UserID)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, &res.Response{})
}
