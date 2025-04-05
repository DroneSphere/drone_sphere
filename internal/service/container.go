package service

import "log/slog"

// Container 服务容器，用于依赖注入
type Container struct {
	User    UserSvc
	Drone   DroneSvc
	Area    AreaSvc
	Wayline WaylineSvc
	Job     JobSvc
	Model   ModelSvc
	Gateway GatewaySvc
	Result  ResultSvc // 添加结果服务
	l       *slog.Logger
}

// NewContainer 创建服务容器
func NewContainer(
	user UserSvc,
	drone DroneSvc,
	area AreaSvc,
	wayline WaylineSvc,
	job JobSvc,
	model ModelSvc,
	gateway GatewaySvc,
	result ResultSvc, // 添加结果服务
	l *slog.Logger,
) *Container {
	return &Container{
		User:    user,
		Drone:   drone,
		Area:    area,
		Wayline: wayline,
		Job:     job,
		Model:   model,
		Gateway: gateway,
		Result:  result, // 添加结果服务
		l:       l,
	}
}
