package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/stretchr/testify/mock"
)

type SpeechProxyService struct {
	mock.Mock
}

func (m *SpeechProxyService) StartConnection(ctx context.Context, body map[string]interface{}) (*res.WebRTCConnectionRes, *errors.AppError) {
	args := m.Called(ctx, body)
	if args.Get(0) == nil {
		return nil, func() *errors.AppError {
			if args.Get(1) == nil {
				return nil
			}
			return args.Get(1).(*errors.AppError)
		}()
	}
	return args.Get(0).(*res.WebRTCConnectionRes), func() *errors.AppError {
		if args.Get(1) == nil {
			return nil
		}
		return args.Get(1).(*errors.AppError)
	}()
}

func (m *SpeechProxyService) ProxyOffer(ctx context.Context, speechSessionID, method string, body []byte) ([]byte, int, *errors.AppError) {
	args := m.Called(ctx, speechSessionID, method, body)
	respBody, _ := args.Get(0).([]byte)
	return respBody, args.Int(1), func() *errors.AppError {
		if args.Get(2) == nil {
			return nil
		}
		return args.Get(2).(*errors.AppError)
	}()
}
