package entity

import (
	"fmt"
	"time"
)

type Drone struct {
	ID           uint      `json:"id" gorm:"primary_key"`
	SN           string    `json:"sn"`
	Domain       string    `json:"domain"`
	Type         int       `json:"type"`
	SubType      int       `json:"sub_type"`
	OnlineStatus bool      `json:"status" gorm:"-"`
	LastLoginAt  time.Time `json:"last_login_at"`
}

func (d *Drone) StatusText() string {
	var statusMap = map[bool]string{
		false: "离线",
		true:  "在线",
	}
	if _, ok := statusMap[d.OnlineStatus]; !ok {
		return "未知"
	}
	return statusMap[d.OnlineStatus]
}

// ProductIdentifier 产品标识符
// 产品标识符由领域、类型、子类型组成, 例如 0-89-0
func (d *Drone) ProductIdentifier() string {
	t := "%s-%d-%d"
	return fmt.Sprintf(t, d.Domain, d.Type, d.SubType)
}

// ProductType 无人机的型号名称
func (d *Drone) ProductType() string {
	var productMap = map[string]string{
		"0-77-0": "Mavic 3E",
		"0-77-1": "Mavic 3T",
	}
	if _, ok := productMap[d.ProductIdentifier()]; !ok {
		return "未知"
	}
	return productMap[d.ProductIdentifier()]
}

// IsRTKAvailable 是否支持RTK
func (d *Drone) IsRTKAvailable() bool {
	return false
}

// IsThermalAvailable 是否支持热成像
func (d *Drone) IsThermalAvailable() bool {
	return false
}
