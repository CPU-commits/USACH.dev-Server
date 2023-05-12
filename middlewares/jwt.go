package middlewares

import (
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/services"
	"github.com/gin-gonic/gin"
)

func JWTMiddleware(isPublic bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := services.VerifyToken(ctx.Request, "access")
		if !isPublic && err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, res.Response{
				Message: err.Error(),
			})
			return
		} else if isPublic && err != nil {
			ctx.Next()
			return
		}
		if !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, res.Response{
				Message: "Unauthorized",
			})
			return
		} else if !token.Valid && !isPublic {
			ctx.Next()
			return
		}
		metadata, err := services.ExtractTokenMetadata(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
				Message: err.Error(),
			})
			return
		}
		ctx.Set("user", metadata)
		ctx.Next()
	}
}
