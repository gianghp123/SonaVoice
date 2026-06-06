package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	"github.com/stretchr/testify/mock"
)

type UnitOfWork struct {
	mock.Mock
	provider transaction.IProvider
}

func (m *UnitOfWork) SetProvider(p transaction.IProvider) {
	m.provider = p
}

func (m *UnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, p transaction.IProvider) error) error {
	args := m.Called(ctx, fn)
	if m.provider != nil {
		if err := fn(ctx, m.provider); err != nil {
			return err
		}
	}
	return args.Error(0)
}
