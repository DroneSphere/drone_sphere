package v1

type DroneUpdateRequest struct {
	Callsign    string `json:"callsign"`    // 呼号
	Description string `json:"description"` // 描述
}

// DroneCreateRequest 创建无人机请求
type DroneCreateRequest struct {
	SN           string `json:"sn" binding:"required"`             // 序列号，必填
	Callsign     string `json:"callsign"`                          // 呼号
	Description  string `json:"description"`                       // 描述
	DroneModelID uint   `json:"drone_model_id" binding:"required"` // 无人机型号ID，必填
	VariationID  uint   `json:"variation_id"`                      // 变体ID，可选
}
