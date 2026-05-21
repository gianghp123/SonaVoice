package transaction

import (
	"gorm.io/gorm"

	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	sessionrepo "github.com/gianghp123/SonaVoice/api/internal/modules/session/repositories"
)

type IProvider interface {
	SessionConfig() repository_interfaces.ISessionConfigRepository
	Session() repository_interfaces.ISessionRepository
	UserQuota() repository_interfaces.IUserQuotaRepository
}

type gormProvider struct {
	tx               *gorm.DB
	sessionConfigRepo repository_interfaces.ISessionConfigRepository
	sessionRepo      repository_interfaces.ISessionRepository
	userQuotaRepo    repository_interfaces.IUserQuotaRepository
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
