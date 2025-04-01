package po

import (
	"gorm.io/gorm"
)

type Drone struct {
	gorm.Model
	SN          string `json:"sn"`          // 序列号
	Callsign    string `json:"callsign"`    // 呼号
	Description string `json:"description"` // 描述

	// 与 DroneModel 的关联（多对一）
	DroneModelID uint       `json:"drone_model_id" gorm:"index"` // 无人机型号ID
	DroneModel   DroneModel `json:"drone_model" gorm:"-"`        // 无人机型号信息

	// 与 DroneVariation 的关联（多对一）
	VariationID uint           `json:"variation_id" gorm:"index"` // 变体ID
	Variation   DroneVariation `json:"variation" gorm:"-"`        // 无人机配置变体
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
	if d.VariationID > 0 {
		if err := db.Preload("Gimbals").Preload("Payloads").First(&d.Variation, d.VariationID).Error; err != nil {
			return err
		}
	}

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
	if d.VariationID > 0 && d.Variation.ID > 0 {
		return d.Variation.GetSummary()
	}
	return d.GetModelSummary()
}
