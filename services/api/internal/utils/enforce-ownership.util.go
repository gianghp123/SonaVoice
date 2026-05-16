package utils

import (
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
)

func EnforceOwnership(ownerID, requesterID string) *errors.AppError {
	if ownerID != requesterID {
		return errors.Forbidden()
	}
	return nil
}
