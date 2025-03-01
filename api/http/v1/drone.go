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
