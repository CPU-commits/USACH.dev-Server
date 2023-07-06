package middlewares

import (
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RepoAccess(maybe bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var idRepository string

		username := ctx.Param("username")
		repository := ctx.Param("repository")
		if ctx.DefaultQuery("repository", "") != "" {
			idRepository = ctx.Query("repository")
		} else if username != "" {
			idObjRepository, errRes := repoService.GetRepositoryId(username, repository)
			if errRes != nil {
				return
			}
			idRepository = idObjRepository.Hex()
		} else {
			idRepository = repository
		}
		// Maybe = true
		if maybe && repository == "" {
			ctx.Next()
			return
		}
		// Object
		idObjRepository, err := primitive.ObjectIDFromHex(idRepository)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
				Message: "Indique un idObject Mongo valido",
			})
			return
		}
		// Repo access
		claims, _ := services.NewClaimsFromContext(ctx)
		// Evaluate
		hasAccess, errRes := repoService.HasRepoAccess(
			idObjRepository,
			claims.UserID,
		)
		if errRes != nil {
			ctx.AbortWithStatusJSON(errRes.StatusCode, &res.Response{
				Message: errRes.Err.Error(),
			})
			return
		}
		if !hasAccess {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, &res.Response{
				Message: "no tienes acceso a este repositorio",
			})
			return
		}
		ctx.Next()
	}
}
