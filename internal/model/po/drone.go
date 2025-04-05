package po

import (
	"time"

	"gorm.io/gorm"
)

type Drone struct {
	ID          uint      `json:"drone_id" gorm:"primaryKey;column:drone_id"`
	CreatedTime time.Time `json:"created_time" gorm:"autoCreateTime;column:created_time"`
	UpdatedTime time.Time `json:"updated_time" gorm:"autoUpdateTime;column:updated_time"`
	State       int       `json:"state" gorm:"default:0;column:state"`               // -1: deleted, 0: active
	SN          string    `json:"sn" gorm:"column:sn"`                               // 序列号
	Callsign    string    `json:"callsign" gorm:"column:callsign"`                   // 呼号
	Description string    `json:"drone_description" gorm:"column:drone_description"` // 描述
	Status      int       `json:"status" gorm:"default:0;column:status"`             // 0: offline, 1: online

	// 与 DroneModel 的关联（多对一）
	DroneModelID uint       `json:"drone_model_id" gorm:"index;column:drone_model_id"` // 无人机型号ID
	DroneModel   DroneModel `json:"drone_model" gorm:"foreignKey:DroneModelID"`        // 无人机型号
}

// TableName 指定 Drone 表名为 tb_drones
func (d Drone) TableName() string {
	return "tb_drones"
}

// 根据无人机查询型号和变体信息
func (d *Drone) LoadModelInfo(db *gorm.DB) error {
	// 加载无人机型号信息
	if d.DroneModelID > 0 {
		if err := db.First(&d.DroneModel, d.DroneModelID).Error; err != nil {
			return err
		}
	}

	// 加载无人机变体信息
	// if d.VariationID > 0 {
	// 	if err := db.Preload("Gimbals").Preload("Payloads").First(&d.Variation, d.VariationID).Error; err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

// 获取无人机型号摘要信息
func (d *Drone) GetModelSummary() string {
	if d.DroneModelID > 0 && d.DroneModel.Name != "" {
		return d.DroneModel.Name
	}
	return "未知型号"
}

// 获取无人机完整配置信息
func (d *Drone) GetConfigSummary() string {
	// if d.VariationID > 0 && d.Variation.ID > 0 {
	// 	return d.Variation.GetSummary()
	// }
	return d.GetModelSummary()
}

// GetStatusText 获取无人机状态文本描述
func (d *Drone) GetStatusText() string {
	if d.Status == 1 {
		return "在线"
	}
	return "离线"
}
