package auth

import (
	"context"
	"errors"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
)

type Actor struct {
	UserID string
	Role   enums.UserRole
}

func ActorFromContext(ctx context.Context) (*Actor, error) {
	userID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
	if userID == "" {
		return nil, errors.New("unauthorized")
	}
	role := utils.GetCtx[enums.UserRole](ctx, enums.ContextKeyUserRole)
	return &Actor{UserID: userID, Role: role}, nil
}
