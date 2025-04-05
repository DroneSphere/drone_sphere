package po

import (
	"time"
)

// Gateway 网关设备持久化对象
type Gateway struct {
	ID             uint         `json:"id" gorm:"primaryKey;column:id"`
	CreatedTime    time.Time    `json:"created_time" gorm:"autoCreateTime;column:created_time"`
	UpdatedTime    time.Time    `json:"updated_time" gorm:"autoUpdateTime;column:updated_time"`
	State          int          `json:"state" gorm:"default:0;column:state"` // -1: deleted, 0: active
	SN             string       `json:"sn" gorm:"unique;column:sn"`          // 序列号
	Callsign       string       `json:"callsign" gorm:"column:callsign"`     // 呼号
	Description    string       `json:"description" gorm:"column:description"`
	GatewayModelID uint         `json:"gateway_model_id" gorm:"column:gateway_model_id"` // 关联网关型号
	GatewayModel   GatewayModel `json:"gateway_model" gorm:"foreignKey:GatewayModelID"`  // 网关型号信息
	UserID         uint         `json:"user_id" gorm:"column:user_id"`                   // 关联用户
	Status         int          `json:"status" gorm:"default:0;column:status"`           // 0: offline, 1: online
	LastOnlineAt   time.Time    `json:"last_online_at" gorm:"column:last_online_at"`     // 最后在线时间

	// 关联的无人机列表
	Drones []Drone `json:"drones" gorm:"many2many:tb_gateway_drone_relations;"`
}

// GatewayDroneRelation 网关与无人机关联关系表
type GatewayDroneRelation struct {
	ID          uint      `json:"id" gorm:"primaryKey;column:id"`
	CreatedTime time.Time `json:"created_time" gorm:"autoCreateTime;column:created_time"`
	UpdatedTime time.Time `json:"updated_time" gorm:"autoUpdateTime;column:updated_time"`
	State       int       `json:"state" gorm:"default:0;column:state"` // -1: deleted, 0: active

	GatewayID   uint      `json:"gateway_id" gorm:"column:gateway_id"`     // 网关ID
	GatewaySN   string    `json:"gateway_sn" gorm:"column:gateway_sn"`     // 网关序列号
	DroneID     uint      `json:"drone_id" gorm:"column:drone_id"`         // 无人机ID
	DroneSN     string    `json:"drone_sn" gorm:"column:drone_sn"`         // 无人机序列号
	ConnectedAt time.Time `json:"connected_at" gorm:"column:connected_at"` // 连接时间
}

// TableName 指定关联表名为 tb_gateway_drone_relations
func (r GatewayDroneRelation) TableName() string {
	return "tb_gateway_drone_relations"
}

// TableName 指定 Gateway 表名为 tb_gateways
func (g Gateway) TableName() string {
	return "tb_gateways"
}

// StatusText 获取状态文本描述
func (g *Gateway) StatusText() string {
	if g.Status == 1 {
		return "在线"
	}
	return "离线"
}

// GetModelName 获取型号名称
func (g *Gateway) GetModelName() string {
	return g.GatewayModel.Name
}
