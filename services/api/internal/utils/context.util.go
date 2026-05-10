package utils

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
)

func GetCtx[T any](ctx context.Context, key enums.ContextKey) T {
	v, _ := ctx.Value(key).(T)
	return v
}
