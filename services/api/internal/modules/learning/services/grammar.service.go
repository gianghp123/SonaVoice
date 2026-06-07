package services

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
)

type IGrammarService interface {
	List(ctx context.Context) *errors.AppError
	Create(ctx context.Context) *errors.AppError
}
