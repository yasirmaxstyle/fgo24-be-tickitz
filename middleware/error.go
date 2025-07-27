package middleware

import (
	"log"
	"net/http"
	"noir-backend/utils"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Printf("Panic recovered: %s", err)
			utils.SendError(c, http.StatusInternalServerError, "Internal server error")
		}
		c.Abort()
	})
}
