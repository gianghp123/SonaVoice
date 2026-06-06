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
	"github.com/gianghp123/SonaVoice/api/internal/database"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/domain"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"gorm.io/gorm"
)

type ISessionSevice interface {
	CreateSession(ctx context.Context) (*res.CreateSessionRes, *errors.AppError)
	GetSession(ctx context.Context, sessionID string) (*models.Session, *errors.AppError)
	ListSessions(ctx context.Context, q req.SessionListQuery) (*response.PaginatedResult[*models.Session], *errors.AppError)
	StartConnection(ctx context.Context, sessionID string) (*res.WebRTCConnectionRes, *errors.AppError)
	FinalizeSession(ctx context.Context, reqBody *req.FinalizeSessionReq) *errors.AppError
	CancelSession(ctx context.Context, sessionID string) *errors.AppError
	ProxyOffer(ctx context.Context, sessionId string, method string, body []byte) ([]byte, int, *errors.AppError)
}

type sessionService struct {
	sessionRepo        repository_interfaces.ISessionRepository
	configService      ISessionConfigService
	speechService      ISpeechProxyService
	startConnectionSvc IStartConnectionService
	quotaRepo          repository_interfaces.IUserQuotaRepository
	uow                transaction.UnitOfWork
	userProfileRepo    repository_interfaces.IUserProfileRepository
}

func NewSessionService(
	sessionRepo repository_interfaces.ISessionRepository,
	configService ISessionConfigService,
	speechService ISpeechProxyService,
	startConnectionSvc IStartConnectionService,
	quotaRepo repository_interfaces.IUserQuotaRepository,
	uow transaction.UnitOfWork,
	userProfileRepo repository_interfaces.IUserProfileRepository,
) ISessionSevice {
	return &sessionService{
		sessionRepo:        sessionRepo,
		configService:      configService,
		speechService:      speechService,
		startConnectionSvc: startConnectionSvc,
		quotaRepo:          quotaRepo,
		uow:                uow,
		userProfileRepo:    userProfileRepo,
	}
}

func (s *sessionService) GetSession(ctx context.Context, sessionID string) (*models.Session, *errors.AppError) {
	logger := zapLogger.S()
	logger.Debugw("Get session", "sessionId", sessionID)

	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		logger.Errorw("Failed to get session", "error", err)
		return nil, errors.MapRepoError(err)
	}

	if appErr := utils.EnforceOwnership(session.UserID, requesterID); appErr != nil {
		return nil, appErr
	}

	return session, nil

}

func (s *sessionService) ListSessions(ctx context.Context, q req.SessionListQuery) (*response.PaginatedResult[*models.Session], *errors.AppError) {
	logger := zapLogger.S()
	logger.Debugw("List sessions", "page", q.Page, "limit", q.Limit)

	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
	status := enums.SessionStatusInactive
	q.UserID = &requesterID
	q.Status = &status

	dbQuery := database.NewQuery().
		SetPage(q.Page).
		SetLimit(q.Limit).
		SetOrderBy("created_at DESC")

	if q.UserID != nil {
		dbQuery.SetFilter("user_id", *q.UserID)
	}

	if q.Status != nil {
		dbQuery.SetFilter("status", *q.Status)
	}

	result, err := s.sessionRepo.List(ctx, dbQuery)
	if err != nil {
		logger.Errorw("Failed to list sessions", "error", err)
		return nil, errors.MapRepoError(err)
	}

	return &response.PaginatedResult[*models.Session]{Data: result.Data, Meta: result.Meta}, nil
}

func (s *sessionService) CreateSession(ctx context.Context) (*res.CreateSessionRes, *errors.AppError) {
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	model, appErr := s.configService.Get(ctx)
	if appErr != nil {
		return nil, appErr
	}
	if model == nil {
		return nil, errors.Internal("global config missing")
	}
	configPayload, err := utils.ParseJSON[dtos.ConfigPayload](model.Config)
	if err != nil {
		return nil, errors.Internal()
	}
	dailyLimit := int64(configPayload.Limits.User.DailyVoiceSeconds)

	quotaDate := utils.QuotaDate()
	remaining, err := s.quotaRepo.GetOrCreate(ctx, requesterID, "voice", quotaDate, dailyLimit)
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

func (s *sessionService) StartConnection(ctx context.Context, sessionID string) (*res.WebRTCConnectionRes, *errors.AppError) {
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	session, appErr := s.GetSession(ctx, sessionID)
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

	connReq := &req.StartConnectionReq{
		EnableDefaultIceServers: true,
		Body: req.StartConnectionBody{
			UserID:      requesterID,
			SessionID:   session.ID,
			MaxDuration: session.MaxDuration,
		},
	}

	profile, err := s.userProfileRepo.GetByUserID(ctx, requesterID)
	if err != nil {
		zapLogger.S().Warnw("Failed to fetch user profile, continuing without it", "userId", requesterID, "error", err)
	} else if profile != nil {
		connReq.Body.UserProfile = buildUserProfileDTO(profile)
	}

	return s.startConnectionSvc.Start(ctx, session, connReq)
}

func buildUserProfileDTO(profile *models.UserProfile) *req.UserProfileDTO {
	dto := &req.UserProfileDTO{
		DisplayName:  profile.DisplayName,
		EnglishLevel: profile.EnglishLevel,
	}

	if len(profile.Preferences) > 0 {
		prefs, err := utils.ParseJSON[req.UserPreferencesPayload](profile.Preferences)
		if err == nil {
			dto.Preferences = &req.UserPreferencesDTO{
				NativeLanguage:       prefs.NativeLanguage,
				ImprovementGoals:     prefs.ImprovementGoals,
				Topics:               prefs.Topics,
				CustomTopics:         prefs.CustomTopics,
				LearningReason:       prefs.LearningReason,
				CustomLearningReason: prefs.CustomLearningReason,
			}
		}
	}

	return dto
}

func (s *sessionService) ProxyOffer(ctx context.Context, sessionId string, method string, body []byte) ([]byte, int, *errors.AppError) {
	logger := zapLogger.S()
	logger.Debug("Proxying offer to speech engine")

	if sessionId == "" {
		return nil, 0, errors.BadRequest("missing session id")
	}

	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	session, appErr := s.GetSession(ctx, sessionId)
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

func (s *sessionService) FinalizeSession(ctx context.Context, reqBody *req.FinalizeSessionReq) *errors.AppError {
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

func (s *sessionService) CancelSession(ctx context.Context, sessionID string) *errors.AppError {
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

func (s *sessionService) cancelStalePendingSession(ctx context.Context, userID string) error {
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
