package entity

type Drone struct {
	SN      string `json:"sn"`
	Domain  string `json:"domain"`
	Type    int    `json:"type"`
	SubType int    `json:"sub_type"`
	Status  int    `json:"status" gorm:"-"`
}

const (
	DroneStatusOffline = 0
	DroneStatusOnline  = 1
)
