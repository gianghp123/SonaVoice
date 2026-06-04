package middlewares

import (
	"net/http"

	appErr "github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"github.com/gin-gonic/gin"
)

func InternalSecretMiddleware() gin.HandlerFunc {
	expectedSecret := utils.GetEnv("INTERNAL_SECRET", "")

	return func(c *gin.Context) {
		logger := zapLogger.S()

		secret := c.GetHeader("X-Internal-Secret")
		if secret != expectedSecret {
			logger.Warnw("Invalid internal secret",
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"ip", c.ClientIP(),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, appErr.Unauthorized())
			return
		}

		c.Next()
	}
}
