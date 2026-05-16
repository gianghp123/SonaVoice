package transaction

import (
	"context"

	"gorm.io/gorm"
)

type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context, p IProvider) error) error
}

type gormUnitOfWork struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) UnitOfWork {
	return &gormUnitOfWork{db: db}
}

func (u *gormUnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, p IProvider) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		provider := NewGormProvider(tx)
		return fn(ctx, provider)
	})
}
