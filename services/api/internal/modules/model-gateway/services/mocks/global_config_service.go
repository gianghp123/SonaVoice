package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/stretchr/testify/mock"
)

type GlobalConfigService struct {
	mock.Mock
}

func (m *GlobalConfigService) Get(ctx context.Context) (*res.GlobalConfigRes, *errors.AppError) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, func() *errors.AppError {
			if args.Get(1) == nil {
				return nil
			}
			return args.Get(1).(*errors.AppError)
		}()
	}
	return args.Get(0).(*dtos.GlobalConfig), func() *errors.AppError {
		if args.Get(1) == nil {
			return nil
		}
		return args.Get(1).(*errors.AppError)
	}()
}

func (m *GlobalConfigService) Update(ctx context.Context, cfg *dtos.GlobalConfig) (*res.GlobalConfigRes, *errors.AppError) {
	args := m.Called(ctx, cfg)
	if args.Get(0) == nil {
		return nil, func() *errors.AppError {
			if args.Get(1) == nil {
				return nil
			}
			return args.Get(1).(*errors.AppError)
		}()
	}
	return args.Get(0).(*dtos.GlobalConfig), func() *errors.AppError {
		if args.Get(1) == nil {
			return nil
		}
		return args.Get(1).(*errors.AppError)
	}()
}
