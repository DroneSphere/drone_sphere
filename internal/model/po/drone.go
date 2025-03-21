package po

import (
	"gorm.io/gorm"
)

type Drone struct {
	gorm.Model
	SN              string `json:"sn"`                      // 序列号
	Callsign        string `json:"callsign"`                // 呼号
	Description     string `json:"description"`             // 描述
	Domain          int    `json:"domain" gorm:"default:0"` // 领域
	Type            int    `json:"type"`                    // 类型
	SubType         int    `json:"sub_type"`                // 子类型
	ProductModel    string `json:"product_model"`           // 产品型号
	ProductModelKey string `json:"product_model_key"`       // 产品型号标识符
}
