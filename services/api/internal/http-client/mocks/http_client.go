package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/stretchr/testify/mock"
)

type HttpClient struct {
	mock.Mock
}

func (m *HttpClient) Do(
	ctx context.Context,
	method string,
	url string,
	headers map[string]string,
	body any,
) ([]byte, int, *errors.AppError) {
	args := m.Called(ctx, method, url, headers, body)
	respBody, _ := args.Get(0).([]byte)
	return respBody, args.Int(1), func() *errors.AppError {
		if args.Get(2) == nil {
			return nil
		}
		return args.Get(2).(*errors.AppError)
	}()
}
