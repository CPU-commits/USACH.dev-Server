package middlewares

import (
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/services"
	"github.com/gin-gonic/gin"
)

func RolesMiddleware(roles []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, _ := services.NewClaimsFromContext(ctx)
		for _, rol := range roles {
			if rol == claims.Role {
				ctx.Next()
				return
			}
		}
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, &res.Response{
			Message: "Unauthorized role",
		})
	}
}
