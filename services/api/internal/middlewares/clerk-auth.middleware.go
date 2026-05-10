package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/gianghp123/SonaVoice/api/internal/configs"
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	appErr "github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gin-gonic/gin"
)

type ClerkMetadata struct {
	Role     enums.UserRole `json:"role"`
	Username string         `json:"username"`
}

func customClaimsConstructor(ctx context.Context) any {
	return &ClerkMetadata{}
}

func withCustomClaims(params *clerkhttp.AuthorizationParams) error {
	params.VerifyParams.CustomClaimsConstructor = customClaimsConstructor
	return nil
}

func ClerkAuth() gin.HandlerFunc {
	logger := zapLogger.S()
	clerkCfg := configs.Load().ClerkAuth
	clerk.SetKey(clerkCfg.ClerkSecret)

	return func(c *gin.Context) {
		handler := clerkhttp.WithHeaderAuthorization(withCustomClaims)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := clerk.SessionClaimsFromContext(r.Context())
			if !ok {
				authHeader := r.Header.Get("Authorization")
				reason := fmt.Sprintf("missing or invalid clerk session, auth header: %s", authHeader)
				if authHeader == "" {
					reason = "missing clerk authorization header"
				}

				logger.Warnw("Unauthorized request",
					"path", r.URL.Path,
					"method", r.Method,
					"ip", c.ClientIP(),
					"reason", reason,
				)
				c.AbortWithStatusJSON(http.StatusUnauthorized, appErr.Unauthorized())
				return
			}

			userID := claims.Subject
			role := enums.UserRoleUser // Default

			if customClaims, ok := claims.Custom.(*ClerkMetadata); ok && customClaims.Role != "" {
				role = customClaims.Role
			}

			ctx := context.WithValue(r.Context(), string(enums.ContextKeyUserID), userID)
			ctx = context.WithValue(ctx, string(enums.ContextKeyUserRole), role)

			// Replace the request context with our new one
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		}))

		handler.ServeHTTP(c.Writer, c.Request)

		if c.IsAborted() {
			return
		}
	}
}
