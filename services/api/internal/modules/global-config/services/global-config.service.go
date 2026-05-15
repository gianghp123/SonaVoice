package services

import (
	"context"
	"encoding/json"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/gianghp123/SonaVoice/api/internal/modules/global-config/dtos"
	"github.com/gianghp123/SonaVoice/api/internal/modules/global-config/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/global-config/repositories"
	"gorm.io/datatypes"
)

type IGlobalConfigService interface {
	Get(ctx context.Context) (*res.GlobalConfigRes, *errors.AppError)
	Update(ctx context.Context, cfg *dtos.GlobalConfig) (*res.GlobalConfigRes, *errors.AppError)
}

type globalConfigService struct {
	repo repositories.IGlobalConfigRepository
}

func NewGlobalConfigService(repo repositories.IGlobalConfigRepository) IGlobalConfigService {
	return &globalConfigService{
		repo: repo,
	}
}

func (s *globalConfigService) Get(ctx context.Context) (*res.GlobalConfigRes, *errors.AppError) {
	logger := zapLogger.S()

	model, err := s.repo.Get(ctx)
	if err != nil {
		logger.Errorw("Failed to get global config", "error", err)
		return nil, errors.Internal("failed to get global config")
	}

	result, appErr := mapModelToDto(model)
	if appErr != nil {
		return nil, appErr
	}

	return result, nil
}

func (s *globalConfigService) Update(ctx context.Context, cfg *dtos.GlobalConfig) (*res.GlobalConfigRes, *errors.AppError) {
	logger := zapLogger.S()

	jsonBytes, err := json.Marshal(cfg)
	if err != nil {
		logger.Errorw("Failed to marshal global config", "error", err)
		return nil, errors.Internal("failed to marshal global config")
	}

	model, err := s.repo.Get(ctx)
	if err != nil {
		logger.Errorw("Failed to get global config for update", "error", err)
		return nil, errors.Internal("failed to get global config")
	}

	model.Config = datatypes.JSON(jsonBytes)

	if err := s.repo.Save(ctx, model); err != nil {
		logger.Errorw("Failed to save global config", "error", err)
		return nil, errors.Internal("failed to save global config")
	}

	result, appErr := mapModelToDto(model)
	if appErr != nil {
		return nil, appErr
	}

	return result, nil
}

func mapModelToDto(model *models.GlobalConfig) (*res.GlobalConfigRes, *errors.AppError) {
	logger := zapLogger.S()

	if model == nil || len(model.Config) == 0 {
		return &res.GlobalConfigRes{}, nil
	}

	var result res.GlobalConfigRes
	if err := json.Unmarshal(model.Config, &result); err != nil {
		logger.Errorw("Failed to unmarshal global config", "error", err)
		return nil, errors.Internal("failed to unmarshal global config")
	}

	return &result, nil
}
