package transaction

import (
	"gorm.io/gorm"

	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	messagerepo "github.com/gianghp123/SonaVoice/api/internal/modules/message/repositories"
	sessionrepo "github.com/gianghp123/SonaVoice/api/internal/modules/session/repositories"
	userprofilerepo "github.com/gianghp123/SonaVoice/api/internal/modules/user_profile/repositories"
)

type IProvider interface {
	SessionConfig() repository_interfaces.ISessionConfigRepository
	Session() repository_interfaces.ISessionRepository
	UserQuota() repository_interfaces.IUserQuotaRepository
	Message() repository_interfaces.IMessageRepository
	UserProfile() repository_interfaces.IUserProfileRepository
}

type gormProvider struct {
	tx               *gorm.DB
	sessionConfigRepo repository_interfaces.ISessionConfigRepository
	sessionRepo      repository_interfaces.ISessionRepository
	userQuotaRepo    repository_interfaces.IUserQuotaRepository
	messageRepo      repository_interfaces.IMessageRepository
	userProfileRepo  repository_interfaces.IUserProfileRepository
}

func NewGormProvider(tx *gorm.DB) IProvider {
	return &gormProvider{
		tx: tx,
	}
}

func (p *gormProvider) SessionConfig() repository_interfaces.ISessionConfigRepository {
	if p.sessionConfigRepo == nil {
		p.sessionConfigRepo = sessionrepo.NewSessionConfigRepository(p.tx)
	}
	return p.sessionConfigRepo
}

func (p *gormProvider) Session() repository_interfaces.ISessionRepository {
	if p.sessionRepo == nil {
		p.sessionRepo = sessionrepo.NewSessionRepository(p.tx)
	}
	return p.sessionRepo
}

func (p *gormProvider) UserQuota() repository_interfaces.IUserQuotaRepository {
	if p.userQuotaRepo == nil {
		p.userQuotaRepo = sessionrepo.NewUserQuotaRepository(p.tx)
	}
	return p.userQuotaRepo
}

func (p *gormProvider) Message() repository_interfaces.IMessageRepository {
	if p.messageRepo == nil {
		p.messageRepo = messagerepo.NewMessageRepository(p.tx)
	}
	return p.messageRepo
}

func (p *gormProvider) UserProfile() repository_interfaces.IUserProfileRepository {
	if p.userProfileRepo == nil {
		p.userProfileRepo = userprofilerepo.NewUserProfileRepository(p.tx)
	}
	return p.userProfileRepo
}
