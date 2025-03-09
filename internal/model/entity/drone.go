package entity

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/model/ro"
	"github.com/jinzhu/copier"
)

type Drone struct {
	ID              uint   `json:"id"`
	SN              string `json:"sn"`                // 序列号
	Callsign        string `json:"callsign"`          // 呼号
	Domain          string `json:"domain"`            // 领域
	Type            int    `json:"type"`              // 类型
	SubType         int    `json:"sub_type"`          // 子类型
	ProductModel    string `json:"product_model"`     // 产品型号
	ProductModelKey string `json:"product_model_key"` // 产品型号标识符
	Status          string `json:"online_status"`     // 在线状态
	dto.DroneMessageProperty
}

func NewDrone(po *po.Drone, rt *ro.Drone) *Drone {
	var d = &Drone{}
	if err := copier.Copy(d, po); err != nil {
		panic(err)
	}
	if rt != nil {
		d.Status = rt.Status
		d.DroneMessageProperty = rt.DroneMessageProperty
	}
	return d
}

func (d *Drone) StatusText() string {
	var statusMap = map[string]string{
		ro.DroneStatusOffline: "离线",
		ro.DroneStatusOnline:  "在线",
		ro.DroneStatusUnknown: "未知",
	}
	return statusMap[d.Status]
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
		"0-99-0": "Mavic 4E",
		"0-99-1": "Mavic 4T",
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
