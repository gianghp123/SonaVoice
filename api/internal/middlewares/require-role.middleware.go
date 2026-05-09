package middlewares

import (
	"net/http"
	"slices"

	"github.com/gianghp123/SonaVoice/api/internal/core"
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	appErr "github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"github.com/gin-gonic/gin"
)

func RequireRole(roles ...enums.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := zapLogger.S()
		role := utils.GetCtx[enums.UserRole](c, core.RoleKey)
		if slices.Contains(roles, role) {
			c.Next()
		}

		logger.Warnw("Unauthorized request",
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"ip", c.ClientIP(),
			"role", role,
		)
		c.AbortWithStatusJSON(http.StatusForbidden, appErr.Forbidden())

		return
	}
}
