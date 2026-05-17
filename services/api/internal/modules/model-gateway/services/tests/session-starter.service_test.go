package tests

import (
	"context"
	"testing"

	appErrors "github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	svcMocks "github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services/mocks"
	services "github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSessionStarterService_StartOrResume_Success(t *testing.T) {
	quotaSvc := new(svcMocks.SessionQuotaService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)

	webrtcRes := &res.WebRTCConnectionRes{SessionID: "speech-s1"}
	quotaSvc.On("Reserve", mock.Anything, "user-1", 3600).Return(int64(300), (*appErrors.AppError)(nil))
	sessionSvc.On("SetReservation", mock.Anything, "s1", int64(300), int64(3600)).Return((*appErrors.AppError)(nil))
	speechSvc.On("StartConnection", mock.Anything, mock.Anything).Return(webrtcRes, (*appErrors.AppError)(nil))
	sessionSvc.On("SetSpeechSessionID", mock.Anything, "s1", "speech-s1").Return((*appErrors.AppError)(nil))

	svc := services.NewSessionStarterService(quotaSvc, sessionSvc, speechSvc)
	session := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "pending"}

	result, appErr := svc.StartOrResume(context.Background(), session, "user-1", 3600)

	require.Nil(t, appErr)
	assert.Equal(t, "s1", result.ID)
	assert.Equal(t, int64(300), result.MaxDuration)
	assert.Equal(t, "speech-s1", result.WebRTCConnectionRes.SessionID)
}

func TestSessionStarterService_StartOrResume_QuotaExceeded(t *testing.T) {
	quotaSvc := new(svcMocks.SessionQuotaService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)

	quotaSvc.On("Reserve", mock.Anything, "user-1", 3600).Return(int64(0), appErrors.Forbidden("quota exceeded"))

	svc := services.NewSessionStarterService(quotaSvc, sessionSvc, speechSvc)
	session := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "pending"}

	result, appErr := svc.StartOrResume(context.Background(), session, "user-1", 3600)

	require.Nil(t, result)
	assert.Equal(t, 403, appErr.Code)
}

func TestSessionStarterService_StartOrResume_SpeechConnectionFailed(t *testing.T) {
	quotaSvc := new(svcMocks.SessionQuotaService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)

	quotaSvc.On("Reserve", mock.Anything, "user-1", 3600).Return(int64(300), (*appErrors.AppError)(nil))
	sessionSvc.On("SetReservation", mock.Anything, "s1", int64(300), int64(3600)).Return((*appErrors.AppError)(nil))
	speechSvc.On("StartConnection", mock.Anything, mock.Anything).Return((*res.WebRTCConnectionRes)(nil), appErrors.Internal("speech error"))
	quotaSvc.On("ReleaseAll", mock.Anything, "user-1", int64(300)).Return(nil)
	sessionSvc.On("MarkSessionFailed", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
	sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

	svc := services.NewSessionStarterService(quotaSvc, sessionSvc, speechSvc)
	session := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "pending"}

	result, appErr := svc.StartOrResume(context.Background(), session, "user-1", 3600)

	require.Nil(t, result)
	assert.Equal(t, 500, appErr.Code)
}

func TestSessionStarterService_StartOrResume_SetReservationFailed(t *testing.T) {
	quotaSvc := new(svcMocks.SessionQuotaService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)

	quotaSvc.On("Reserve", mock.Anything, "user-1", 3600).Return(int64(300), (*appErrors.AppError)(nil))
	sessionSvc.On("SetReservation", mock.Anything, "s1", int64(300), int64(3600)).Return(appErrors.Internal("db error"))
	quotaSvc.On("ReleaseAll", mock.Anything, "user-1", int64(300)).Return(nil)

	svc := services.NewSessionStarterService(quotaSvc, sessionSvc, speechSvc)
	session := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "pending"}

	result, appErr := svc.StartOrResume(context.Background(), session, "user-1", 3600)

	require.Nil(t, result)
	assert.Equal(t, 500, appErr.Code)
}
