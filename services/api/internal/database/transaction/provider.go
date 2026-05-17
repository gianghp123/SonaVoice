package transaction

import (
	"gorm.io/gorm"

	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	modelgatewayrepo "github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/repositories"
)

type IProvider interface {
	GlobalConfig() repository_interfaces.IGlobalConfigRepository
	Session() repository_interfaces.ISessionRepository
	UserQuota() repository_interfaces.IUserQuotaRepository
}

type gormProvider struct {
	tx               *gorm.DB
	globalConfigRepo repository_interfaces.IGlobalConfigRepository
	sessionRepo      repository_interfaces.ISessionRepository
	userQuotaRepo    repository_interfaces.IUserQuotaRepository
}

func NewGormProvider(tx *gorm.DB) IProvider {
	return &gormProvider{
		tx: tx,
	}
}

func (p *gormProvider) GlobalConfig() repository_interfaces.IGlobalConfigRepository {
	if p.globalConfigRepo == nil {
		p.globalConfigRepo = modelgatewayrepo.NewGlobalConfigRepository(p.tx)
	}
	return p.globalConfigRepo
}

func (p *gormProvider) Session() repository_interfaces.ISessionRepository {
	if p.sessionRepo == nil {
		p.sessionRepo = modelgatewayrepo.NewSessionRepository(p.tx)
	}
	return p.sessionRepo
}

func (p *gormProvider) UserQuota() repository_interfaces.IUserQuotaRepository {
	if p.userQuotaRepo == nil {
		p.userQuotaRepo = modelgatewayrepo.NewUserQuotaRepository(p.tx)
	}
	return p.userQuotaRepo
}
