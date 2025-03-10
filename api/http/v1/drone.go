package v1

type DroneUpdateRequest struct {
	Callsign string `json:"callsign" binding:"required"` // 呼号
}

type DroneItemResult struct {
	ID                 uint   `json:"id"`       // ID
	Callsign           string `json:"callsign"` // 呼号
	SN                 string `json:"sn"`
	Status             string `json:"status"`
	ProductModel       string `json:"product_model"` // 产品型号
	IsRTKAvailable     bool   `json:"is_rtk_available"`
	IsThermalAvailable bool   `json:"is_thermal_available"`
	CreatedAt          string `json:"created_at"`     // 创建时间
	LastOnlineAt       string `json:"last_online_at"` // 最后在线时间
}

type DroneDetailResult struct {
	ID                 uint   `json:"id"`
	SN                 string `json:"sn" binding:"required"`                // 序列号
	Callsign           string `json:"callsign"`                             // 呼号
	Domain             int    `json:"domain" binding:"required"`            // 领域
	Type               int    `json:"type" binding:"required"`              // 类型
	SubType            int    `json:"sub_type" binding:"required"`          // 子类型
	ProductModel       string `json:"product_model" binding:"required"`     // 产品型号
	ProductModelKey    string `json:"product_model_key" binding:"required"` // 产品型号标识符
	Status             string `json:"status"`                               // 在线状态
	IsRTKAvailable     bool   `json:"is_rtk_available"`                     // 是否支持RTK◊
	IsThermalAvailable bool   `json:"is_thermal_available"`                 // 是否支持热成像
}
