package service

import (
	"log/slog"

	"github.com/dronesphere/internal/repo"
)

type (
	// GatewaySvc 网关设备服务接口
	GatewaySvc interface {
		Repo() repo.GatewayRepo
	}

	// GatewayImpl 网关设备服务实现
	GatewayImpl struct {
		repo repo.GatewayRepo
		l    *slog.Logger
	}
)

// NewGatewayImpl 创建网关设备服务实例
func NewGatewayImpl(repo repo.GatewayRepo, l *slog.Logger) GatewaySvc {
	return &GatewayImpl{
		repo: repo,
		l:    l,
	}
}

// Repo 获取仓储实例
func (s *GatewayImpl) Repo() repo.GatewayRepo {
	return s.repo
}
