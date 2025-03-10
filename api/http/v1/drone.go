package v1

type DroneItemResult struct {
	ID       uint   `json:"id"`       // ID
	Callsign string `json:"callsign"` // 呼号
	// 以下字段来自实体信息
	SN      string `json:"sn"`
	Domain  string `json:"domain"`
	Type    int    `json:"type"`
	SubType int    `json:"sub_type"`
	// 以上字段来自实体信息
	Status string `json:"status"`
	// ProductType 无人机的型号名称
	ProductType string `json:"product_type"`
	// IsRTKAvailable 是否支持RTK
	IsRTKAvailable bool `json:"is_rtk_available"`
	// IsThermalAvailable 是否支持热成像
	IsThermalAvailable bool `json:"is_thermal_available"`
	// LastLoginAt 最后登录时间
	LastLoginAt string `json:"last_login_at"`
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
