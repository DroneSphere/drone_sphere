package entity

import (
	"time"

	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/model/ro"
	"github.com/jinzhu/copier"
)

type Drone struct {
	ID          uint   `json:"id"`
	SN          string `json:"sn"`          // 序列号
	Callsign    string `json:"callsign"`    // 呼号
	Description string `json:"description"` // 描述
	Type        int    `json:"type"`        // 类型（已废弃，保留向后兼容）
	SubType     int    `json:"sub_type"`    // 子类型（已废弃，保留向后兼容）

	// 型号信息 - 关联字段
	DroneModelID uint              `json:"drone_model_id"` // 无人机型号ID
	DroneModel   po.DroneModel     `json:"drone_model"`    // 无人机型号
	VariationID  uint              `json:"variation_id"`   // 变体ID
	Variation    po.DroneVariation `json:"variation"`      // 无人机变体配置

	// 状态信息
	Status    string `json:"status"` // 在线状态
	CreatedAt time.Time
	UpdatedAt time.Time
	dto.DroneMessageProperty
}

// NewDrone 从持久化对象和实时对象创建无人机实体
func NewDrone(po *po.Drone, rt *ro.Drone) *Drone {
	var d = &Drone{}
	if err := copier.Copy(d, po); err != nil {
		panic(err)
	}

	// 实体中添加型号信息
	d.DroneModelID = po.DroneModelID
	d.VariationID = po.VariationID

	// 填充实时状态信息
	if rt != nil {
		d.Status = rt.Status
		d.DroneMessageProperty = rt.DroneMessageProperty
	}
	return d
}

// StatusText 获取状态文本描述
func (d *Drone) StatusText() string {
	var statusMap = map[string]string{
		ro.DroneStatusOffline: "离线",
		ro.DroneStatusOnline:  "在线",
		ro.DroneStatusUnknown: "未知",
	}
	return statusMap[d.Status]
}

// GetModelName 获取无人机型号名称
func (d *Drone) GetModelName() string {
	if d.DroneModel.ID > 0 {
		return d.DroneModel.Name
	}
	return "未知型号"
}

// GetConfigSummary 获取无人机配置摘要
func (d *Drone) GetConfigSummary() string {
	if d.Variation.ID > 0 {
		return d.Variation.GetSummary()
	}
	return d.GetModelName()
}

// IsRTKAvailable 是否支持RTK
func (d *Drone) IsRTKAvailable() bool {
	// 优先从变体信息判断
	if d.Variation.ID > 0 {
		return d.Variation.SupportsRTK()
	}

	// 从型号信息判断
	if d.DroneModel.ID > 0 {
		return d.DroneModel.IsRTKAvailable
	}

	// 无法确定时假设不支持
	return false
}

// IsThermalAvailable 是否支持热成像
func (d *Drone) IsThermalAvailable() bool {
	// 优先从变体信息判断
	if d.Variation.ID > 0 {
		return d.Variation.SupportsThermal()
	}

	// 从无人机型号的云台判断（需要进一步实现）
	// 无法确定时假设不支持
	return false
}
