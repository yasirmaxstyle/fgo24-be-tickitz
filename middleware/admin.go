package middleware

import (
	"net/http"
	"noir-backend/utils"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, exists := ctx.Get("role")
		if !exists {
			utils.SendError(ctx, http.StatusUnauthorized, "Status Unauthorized")
			ctx.Abort()
			return
		}

		if role != "admin" {
			utils.SendError(ctx, http.StatusForbidden, "only admin can access")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
