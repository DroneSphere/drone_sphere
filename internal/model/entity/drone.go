package entity

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dronesphere/internal/model/po"
)

type Drone struct {
	ID       uint   `json:"id"`
	Callsign string `json:"callsign"` // 呼号
	SN       string `json:"sn"`
	Domain   string `json:"domain"`
	Type     int    `json:"type"`
	SubType  int    `json:"sub_type"`
	po.RTDrone
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
		"0-99-0": "Mavic 3E",
		"0-99-1": "Mavic 3T",
	}
	if _, ok := productMap[d.ProductIdentifier()]; !ok {
		return "未知"
	}
	return productMap[d.ProductIdentifier()]
}

// IsRTKAvailable 是否支持RTK
func (d *Drone) IsRTKAvailable() bool {
	// 主型号支持RTK的集合
	supportTypes := mapset.NewSet[int]()
	supportTypes.Add(99)
	if supportTypes.Contains(d.Type) {
		return true
	}
	// 主型号不全都支持，但部分子型号支持RTK
	type SubType struct {
		Type    int
		SubType int
	}
	supportSubTypes := mapset.NewSet[SubType]()
	supportSubTypes.Add(SubType{Type: 77, SubType: -1})
	if supportSubTypes.Contains(SubType{Type: d.Type, SubType: d.SubType}) {
		return true
	}
	// 其他情况均不支持RTK
	return false
}

// IsThermalAvailable 是否支持热成像
func (d *Drone) IsThermalAvailable() bool {
	// 主型号支持热成像的集合
	supportTypes := mapset.NewSet[int]()
	supportTypes.Add(-1)
	if supportTypes.Contains(d.Type) {
		return true
	}
	// 主型号不全都支持，但部分子型号支持热成像
	type SubType struct {
		Type    int
		SubType int
	}
	supportSubTypes := mapset.NewSet[SubType]()
	supportSubTypes.Add(SubType{Type: 77, SubType: 1})
	supportSubTypes.Add(SubType{Type: 99, SubType: 1})
	if supportSubTypes.Contains(SubType{Type: d.Type, SubType: d.SubType}) {
		return true
	}
	return false
}
