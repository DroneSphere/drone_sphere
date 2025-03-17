package entity

import (
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/pkg/misc"
)

type DroneModel struct {
	misc.BaseModel
	// 型号名称
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// 领域
	Domain int `json:"domain"`
	// 主类型
	Type int `json:"type"`
	// 自类型
	SubType int `json:"sub_type"`
	// 对应的网关描述
	GatewayDescription string `json:"gateway_description,omitempty"`
	// 对应的网关ID
	GatewayID uint `json:"gateway_id"`
	// 对应的网关名称
	GatewayName string `json:"gateway_name"`
	// 可搭载云台
	Gimbals []po.GimbalModel `json:"gimbals,omitempty" gorm:"-"`
}

func NewDroneModelFromPO(drone po.DroneModel, gateway po.GatewayModel) *DroneModel {
	return &DroneModel{
		BaseModel:          drone.BaseModel,
		Name:               drone.Name,
		Description:        drone.Description,
		Domain:             drone.Domain,
		Type:               drone.Type,
		SubType:            drone.SubType,
		GatewayDescription: gateway.Description,
		GatewayID:          drone.GatewayID,
		GatewayName:        gateway.Name,
		Gimbals:            drone.Gimbals,
	}
}
