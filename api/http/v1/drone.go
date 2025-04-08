package v1

type DroneUpdateRequest struct {
	Callsign    string `json:"callsign"`    // 呼号
	Description string `json:"description"` // 描述
}
