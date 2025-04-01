package v1

type DroneUpdateRequest struct {
	Callsign string `json:"callsign" binding:"required"` // 呼号
}
