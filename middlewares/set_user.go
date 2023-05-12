package middlewares

import (
	"github.com/CPU-commits/USACH.dev-Server/services"
	"github.com/gin-gonic/gin"
)

const USER_ID_KEY = "USER_ID_REQ"

func SetUserID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, exists := services.NewClaimsFromContext(ctx)
		if !exists {
			ctx.Set(USER_ID_KEY, ctx.ClientIP())
		} else {
			ctx.Set(USER_ID_KEY, claims.UserID)
		}

		ctx.Next()
	}
}
