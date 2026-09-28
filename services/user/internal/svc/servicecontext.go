package svc

import "demo-shop/services/user/internal/config"

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(config config.Config) *ServiceContext {
	return &ServiceContext{
		Config: config,
	}
}
