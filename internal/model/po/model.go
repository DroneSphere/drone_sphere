package po

import (
	"github.com/dronesphere/internal/pkg/misc"
)

type GatewayModel struct {
	misc.BaseModel
	// 型号名称，DJI 文档中收录的标准名称
	Name string `json:"name"`
	// 描述，自定义的型号描述
	Description string `json:"description,omitempty"`
	// 领域，DJI 文档指定
	Domain int `json:"domain"`
	// 主型号，DJI 文档指定
	Type int `json:"type"`
	// 子型号，DJI 文档指定
	SubType int `json:"sub_type"`
}

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
	// 对应的网关ID
	GatewayID uint `json:"gateway_id"`
	// 可搭载云台
	Gimbals []GimbalModel `json:"gimbals,omitempty" gorm:"many2many:drone_gimbal;"`
}

type GimbalModel struct {
	misc.BaseModel
	// 云台名称
	Name string `json:"name"`
	// 描述
	Description string `json:"description"`
	// 产品线名称，FPV相机、云台相机和机场相机
	Product string `json:"product"`
	// 领域
	Domain int `json:"domain"`
	// 型号
	Type int `json:"type"`
	// 子型号
	SubType int `json:"sub_type"`
	// 相机位置索引
	Gimbalindex int `json:"gimbalindex"`
	// 对应的无人机
	Drons []DroneModel `json:"drones,omitempty" gorm:"many2many:drone_gimbal;"`
}

type PayloadModel struct {
	misc.BaseModel
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category"`
}
