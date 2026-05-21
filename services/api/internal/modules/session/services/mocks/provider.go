package mocks

import (
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"github.com/stretchr/testify/mock"
)

type Provider struct {
	mock.Mock
}

func (p *Provider) SessionConfig() repository_interfaces.ISessionConfigRepository {
	args := p.Called()
	return args.Get(0).(repository_interfaces.ISessionConfigRepository)
}

func (p *Provider) Session() repository_interfaces.ISessionRepository {
	args := p.Called()
	return args.Get(0).(repository_interfaces.ISessionRepository)
}

func (p *Provider) UserQuota() repository_interfaces.IUserQuotaRepository {
	args := p.Called()
	return args.Get(0).(repository_interfaces.IUserQuotaRepository)
}
