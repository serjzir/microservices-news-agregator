package middlewares
import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.New().String()
		c.ClientIP()
		c.Set("request_id", id)
		c.Header("request_id", id)
		c.Next()
	}
}