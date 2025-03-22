package po

import (
	"gorm.io/gorm"
)

type Drone struct {
	gorm.Model
	SN              string `json:"sn"`                      // 序列号
	Callsign        string `json:"callsign"`                // 呼号
	Description     string `json:"description"`             // 描述
	Domain          int    `json:"domain" gorm:"default:0"` // Deprecated, 领域
	Type            int    `json:"type"`                    // Deprecated, 类型
	SubType         int    `json:"sub_type"`                // Deprecated, 子类型
	ModelID         uint   `json:"model_id"`                // 型号ID
	ProductModel    string `json:"product_model"`           // 产品型号
	ProductModelKey string `json:"product_model_key"`       // 产品型号标识符
}
