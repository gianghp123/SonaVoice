package services

import (
	"context"
	"encoding/json"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos"
	"gorm.io/datatypes"
)

type ISessionConfigService interface {
	Get(ctx context.Context) (*models.SessionConfig, *errors.AppError)
	Update(ctx context.Context, cfg *dtos.SessionConfig) (*models.SessionConfig, *errors.AppError)
}

type sessionConfigService struct {
	repo repository_interfaces.ISessionConfigRepository
}

func NewSessionConfigService(repo repository_interfaces.ISessionConfigRepository) ISessionConfigService {
	return &sessionConfigService{
		repo: repo,
	}
}

func (s *sessionConfigService) Get(ctx context.Context) (*models.SessionConfig, *errors.AppError) {
	logger := zapLogger.S()

	model, err := s.repo.Get(ctx)
	if err != nil {
		logger.Errorw("Failed to get global config", "error", err)
		return nil, errors.Internal()
	}

	return model, nil
}

func (s *sessionConfigService) Update(ctx context.Context, cfg *dtos.SessionConfig) (*models.SessionConfig, *errors.AppError) {
	logger := zapLogger.S()

	if cfg == nil {
		return nil, errors.BadRequest("global config is required")
	}

	jsonBytes, err := json.Marshal(cfg.Config)
	if err != nil {
		logger.Errorw("Failed to marshal global config", "error", err)
		return nil, errors.Internal()
	}

	model, err := s.repo.Get(ctx)
	if err != nil {
		logger.Errorw("Failed to get global config for update", "error", err)
		return nil, errors.Internal()
	}

	model.Config = datatypes.JSON(jsonBytes)

	if err := s.repo.Save(ctx, model); err != nil {
		logger.Errorw("Failed to save global config", "error", err)
		return nil, errors.Internal()
	}

	return model, nil
}


