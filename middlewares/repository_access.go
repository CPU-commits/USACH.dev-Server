package middlewares

import (
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/services"
	"github.com/CPU-commits/USACH.dev-Server/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RepoAccess() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var idRepository string

		username := ctx.Param("username")
		repository := ctx.Param("repository")
		if username != "" {
			idObjRepository, errRes := repoService.GetRepositoryId(username, repository)
			if errRes != nil {
				return
			}
			idRepository = idObjRepository.Hex()
		} else {
			idRepository = repository
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
		repoAccess, errRes := repoService.GetRepoAccess(idObjRepository)
		if errRes != nil {
			ctx.AbortWithStatusJSON(errRes.StatusCode, &res.Response{
				Message: errRes.Err.Error(),
			})
			return
		}
		if repoAccess["access"] == "private-group" {
			claims, exists := services.NewClaimsFromContext(ctx)
			if !exists {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, &res.Response{
					Message: "no tienes acceso a este repositorio",
				})
				return
			}
			// Evaluate
			idObjUser, err := primitive.ObjectIDFromHex(claims.UserID)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
					Message: err.Error(),
				})
				return
			}

			hasAccess, err := utils.AnyMatch(
				repoAccess["custom_access"],
				func(x interface{}) bool {
					return x.(primitive.ObjectID) != idObjUser
				},
			)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, &res.Response{
					Message: err.Error(),
				})
				return
			}
			if !hasAccess {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, &res.Response{
					Message: "no tienes acceso a este repositorio",
				})
				return
			}
		}
		ctx.Next()
	}
}
