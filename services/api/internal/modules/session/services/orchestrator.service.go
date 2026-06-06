package services

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/domain"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"gorm.io/gorm"
)

type IOrchestratorService interface {
	CreateSession(ctx context.Context) (*res.CreateSessionRes, *errors.AppError)
	GetSession(ctx context.Context, sessionID string) (*res.SessionRes, *errors.AppError)
	ListSessions(ctx context.Context, q req.SessionListQuery) (*response.PaginatedResult[*res.SessionListItemRes], *errors.AppError)
	StartConnection(ctx context.Context, sessionID string) (*res.WebRTCConnectionRes, *errors.AppError)
	FinalizeSession(ctx context.Context, reqBody *req.FinalizeSessionReq) *errors.AppError
	CancelSession(ctx context.Context, sessionID string) *errors.AppError
	ProxyOffer(ctx context.Context, sessionId string, method string, body []byte) ([]byte, int, *errors.AppError)
}

type orchestratorService struct {
	configService      ISessionConfigService
	sessionService     ISessionService
	speechService      ISpeechProxyService
	startConnectionSvc IStartConnectionService
	quotaService       IQuotaService
	uow                transaction.UnitOfWork
}

func NewOrchestratorService(
	configService ISessionConfigService,
	sessionService ISessionService,
	speechService ISpeechProxyService,
	startConnectionSvc IStartConnectionService,
	quotaService IQuotaService,
	uow transaction.UnitOfWork,
) IOrchestratorService {
	return &orchestratorService{
		configService:      configService,
		sessionService:     sessionService,
		speechService:      speechService,
		startConnectionSvc: startConnectionSvc,
		quotaService:       quotaService,
		uow:                uow,
	}
}

func (s *orchestratorService) GetSession(ctx context.Context, sessionID string) (*res.SessionRes, *errors.AppError) {
	logger := zapLogger.S()
	logger.Debugw("Get session", "sessionId", sessionID)

	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	model, appErr := s.sessionService.Get(ctx, sessionID)
	if appErr != nil {
		return nil, appErr
	}

	if model == nil {
		return nil, errors.NotFound("session not found")
	}

	if appErr := utils.EnforceOwnership(model.UserID, requesterID); appErr != nil {
		return nil, appErr
	}

	return &res.SessionRes{
		ID:        model.ID,
		UserID:    model.UserID,
		Status:    model.Status,
		CreatedAt: model.CreatedAt,
	}, nil
}

func (s *orchestratorService) ListSessions(ctx context.Context, q req.SessionListQuery) (*response.PaginatedResult[*res.SessionListItemRes], *errors.AppError) {
	logger := zapLogger.S()
	logger.Debugw("List sessions", "page", q.Page, "limit", q.Limit)

	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
	status := enums.SessionStatusInactive
	q.UserID = &requesterID
	q.Status = &status

	result, appErr := s.sessionService.List(ctx, q)
	if appErr != nil {
		return nil, appErr
	}

	var items []*res.SessionListItemRes
	err := utils.MapToDTOs(result.Data, &items)
	if err != nil {
		logger.Errorw("Failed to map sessions to DTO", "error", err)
		return nil, errors.Internal()
	}

	return &response.PaginatedResult[*res.SessionListItemRes]{Data: items, Meta: result.Meta}, nil
}

func (s *orchestratorService) CreateSession(ctx context.Context) (*res.CreateSessionRes, *errors.AppError) {
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	model, appErr := s.configService.Get(ctx)
	if appErr != nil {
		return nil, appErr
	}
	if model == nil {
		return nil, errors.Internal("global config missing")
	}
	configPayload, err := domain.ParseSessionConfig(model.Config)
	if err != nil {
		return nil, errors.Internal()
	}
	dailyLimit := int64(configPayload.Limits.User.DailyVoiceSeconds)

	remaining, err := s.quotaService.CheckRemaining(ctx, requesterID, dailyLimit)
	if err != nil {
		zapLogger.S().Errorw("Failed to check remaining quota", "userId", requesterID, "error", err)
		return nil, errors.Internal()
	}
	if remaining <= 0 {
		return nil, errors.Forbidden("quota exceeded")
	}

	if err := s.cancelStalePendingSession(ctx, requesterID); err != nil {
		sentry.CaptureException(err)
		zapLogger.S().Errorw("Failed to cancel stale pending session", "userId", requesterID, "error", err)
		return nil, errors.Internal()
	}

	var session *models.Session

	if err := s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		sessionRepo := p.Session()

		quoteDate := utils.QuotaDate()

		session = &models.Session{
			UserID:      requesterID,
			Status:      enums.SessionStatusPending,
			QuotaDate:   &quoteDate,
			MaxDuration: remaining,
		}

		if err := sessionRepo.Create(ctx, session); err != nil {
			zapLogger.S().Errorw("Failed to create session", "error", err)
			return err
		}

		return nil
	}); err != nil {
		return nil, errors.MapRepoError(err)
	}

	return &res.CreateSessionRes{
		ID:                  session.ID,
		MaxDuration:         0,
		WebRTCConnectionRes: nil,
	}, nil
}

func (s *orchestratorService) StartConnection(ctx context.Context, sessionID string) (*res.WebRTCConnectionRes, *errors.AppError) {
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	session, appErr := s.sessionService.Get(ctx, sessionID)
	if appErr != nil {
		return nil, appErr
	}

	if appErr := utils.EnforceOwnership(session.UserID, requesterID); appErr != nil {
		return nil, appErr
	}

	if session.SpeechSessionID != "" && session.SpeechStartResponse != nil {
		var cached res.WebRTCConnectionRes
		if err := json.Unmarshal(session.SpeechStartResponse, &cached); err != nil {
			zapLogger.S().Errorw("Failed to deserialize cached speech start response", "sessionId", sessionID, "error", err)
			return nil, errors.Internal()
		}
		return &cached, nil
	}

	domainSession := domain.NewSessionFromModel(session)
	if appErr := domainSession.CanBeStarted(); appErr != nil {
		return nil, appErr
	}

	return s.startConnectionSvc.Start(ctx, session, requesterID)
}

func (s *orchestratorService) ProxyOffer(ctx context.Context, sessionId string, method string, body []byte) ([]byte, int, *errors.AppError) {
	logger := zapLogger.S()
	logger.Debug("Proxying offer to speech engine")

	if sessionId == "" {
		return nil, 0, errors.BadRequest("missing session id")
	}

	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	session, appErr := s.sessionService.Get(ctx, sessionId)
	if appErr != nil {
		logger.Errorw("Failed to get app session", "sessionId", sessionId, "error", appErr)
		return nil, 0, appErr
	}
	if appErr := utils.EnforceOwnership(session.UserID, requesterID); appErr != nil {
		logger.Errorw("Ownership enforcement failed", "sessionId", sessionId, "error", appErr)
		return nil, 0, appErr
	}

	speechSessionId := session.SpeechSessionID
	if speechSessionId == "" {
		return nil, 0, errors.BadRequest("session has not started a speech connection")
	}

	if method == http.MethodPatch {
		speechCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		return s.speechService.ProxyOffer(speechCtx, speechSessionId, method, body)
	}

	var responseBody []byte
	var statusCode int
	var proxyErr *errors.AppError

	err := s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		sessionRepo := p.Session()

		sess, err := sessionRepo.GetForUpdate(ctx, session.ID)
		if err != nil {
			return err
		}

		wasPending := sess.Status == enums.SessionStatusPending

		speechCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		responseBody, statusCode, appErr = s.speechService.ProxyOffer(speechCtx, speechSessionId, method, body)
		if appErr != nil {
			logger.Errorw("Speech engine error during offer", "sessionId", sess.ID, "speechSessionId", speechSessionId, "error", appErr)
			if wasPending {
				if err := sessionRepo.SetSessionFailed(ctx, sess.ID); err != nil {
					logger.Errorw("Failed to mark session failed", "sessionId", sess.ID, "error", err)
					return err
				}
			}
			proxyErr = appErr
			return nil
		}

		if wasPending {
			if err := sessionRepo.SetSessionActive(ctx, sess.ID, utils.NowUTC()); err != nil {
				logger.Errorw("Failed to mark session active", "sessionId", sess.ID, "error", err)
				if err := sessionRepo.SetSessionFailed(ctx, sess.ID); err != nil {
					logger.Errorw("Failed to mark session failed after activation error", "sessionId", sess.ID, "error", err)
					return err
				}
				proxyErr = errors.Internal("failed to activate session")
				return nil
			}
		}

		return nil
	})

	if err != nil {
		logger.Errorw("ProxyOffer transaction failed", "speechSessionId", speechSessionId, "error", err)
		return nil, 0, errors.MapRepoError(err)
	}

	if proxyErr != nil {
		return responseBody, statusCode, proxyErr
	}

	return responseBody, statusCode, nil
}

func (s *orchestratorService) FinalizeSession(ctx context.Context, reqBody *req.FinalizeSessionReq) *errors.AppError {
	logger := zapLogger.S()

	if reqBody == nil {
		return errors.BadRequest("request body is required")
	}

	sessionId := reqBody.SessionID

	logger.Debugw("Finalizing session", "sessionId", sessionId, "actualUsage", reqBody.ActualUsage)

	if sessionId == "" {
		return errors.BadRequest("sessionId is required")
	}
	if reqBody.ActualUsage < 0 {
		return errors.BadRequest("actualUsage cannot be negative")
	}

	err := s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		sessionRepo := p.Session()

		session, err := sessionRepo.GetForUpdate(ctx, sessionId)
		if err != nil {
			return err
		}

		domainSession := domain.NewSessionFromModel(session)
		if appErr := domainSession.CanBeFinalized(); appErr != nil {
			return appErr
		}

		actualUsage := min(int64(reqBody.ActualUsage), domainSession.MaxDuration)

		if actualUsage > 0 {
			quotaRepo := p.UserQuota()
			if err := quotaRepo.Deduct(ctx, session.UserID, "voice", *session.QuotaDate, actualUsage); err != nil {
				logger.Errorw("Failed to deduct quota", "sessionId", sessionId, "error", err)
				return err
			}
		}

		if err := sessionRepo.SetActualUsage(ctx, session.ID, actualUsage); err != nil {
			logger.Errorw("Failed to set actual usage", "sessionId", sessionId, "error", err)
			return err
		}

		if session.Status != enums.SessionStatusInactive {
			return sessionRepo.SetSessionInactive(ctx, session.ID, utils.NowUTC())
		}

		return nil
	})

	if err != nil {
		return errors.MapRepoError(err)
	}
	return nil
}

func (s *orchestratorService) CancelSession(ctx context.Context, sessionID string) *errors.AppError {
	logger := zapLogger.S()

	if sessionID == "" {
		return errors.BadRequest("sessionId is required")
	}

	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	err := s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		sessionRepo := p.Session()

		session, err := sessionRepo.GetForUpdate(ctx, sessionID)
		if err != nil {
			return err
		}

		if appErr := utils.EnforceOwnership(session.UserID, requesterID); appErr != nil {
			return appErr
		}

		domainSession := domain.NewSessionFromModel(session)
		if appErr := domainSession.CanBeCancelled(); appErr != nil {
			return appErr
		}

		return sessionRepo.SetSessionInactive(ctx, session.ID, utils.NowUTC())
	})

	if err != nil {
		logger.Errorw("Failed to cancel session", "sessionId", sessionID, "error", err)
		return errors.MapRepoError(err)
	}
	return nil
}

func (s *orchestratorService) cancelStalePendingSession(ctx context.Context, userID string) error {
	return s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		sessionRepo := p.Session()

		staleSession, err := sessionRepo.GetPendingByUserIDForUpdate(ctx, userID)
		if err != nil {
			if stdErrors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		if !utils.IsStale(staleSession.CreatedAt, 30*time.Second) {
			return nil
		}

		return sessionRepo.SetSessionFailed(ctx, staleSession.ID)
	})
}
