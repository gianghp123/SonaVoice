package auth

import (
	"errors"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
)

var ErrForbidden = errors.New("forbidden")

func CanPerform(actor *Actor, resourceOwnerID string) error {
	if actor.Role == enums.UserRoleAdmin {
		return nil
	}
	if actor.UserID == resourceOwnerID {
		return nil
	}
	return ErrForbidden
}
