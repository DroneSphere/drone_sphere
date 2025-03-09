package ro

import "github.com/dronesphere/internal/model/dto"

const (
	DroneStatusOnline  = "online"
	DroneStatusOffline = "offline"
	DroneStatusUnknown = "unknown"
)

type Drone struct {
	ID     uint   `json:"id"`
	SN     string `json:"sn"`
	Status string `json:"online_status"`
	dto.DroneMessageProperty
}
