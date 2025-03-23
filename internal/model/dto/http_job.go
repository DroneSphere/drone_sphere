package dto

type JobCreationDrone struct {
	Color       string          `json:"color"`
	Description string          `json:"description"`
	ID          uint            `json:"id"`
	Index       int             `json:"index"`
	Key         string          `json:"key"`
	Model       string          `json:"model"`
	Name        string          `json:"name"`
	Variantion  DroneVariantion `json:"variantion"`
}

type JobCreationWayline struct {
	DroneKey string  `json:"drone_key"`
	Color    string  `json:"color"`
	Height   float64 `json:"height"`
	Path     []struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"path"`
	Points []struct {
		Index int     `json:"index"`
		Lat   float64 `json:"lat"`
		Lng   float64 `json:"lng"`
	} `json:"points"`
}

type JobCreationMapping struct {
	PhysicalDroneID       uint   `json:"physical_drone_id"`
	PhysicalDroneSN       string `json:"physical_drone_sn"`
	SelectedDroneKey      string `json:"selected_drone_key"`
	PhysicalDroneCallsign string `json:"physical_drone_callsign"`
}

type DroneVariantion struct {
	Gimbal           Gimbal `json:"gimbal"`
	Index            int    `json:"index"`
	Name             string `json:"name"`
	RtkAvailable     bool   `json:"rtk_available"`
	ThermalAvailable bool   `json:"thermal_available"`
}

type Gimbal struct {
	Description string `json:"description"`
	Domain      int    `json:"domain"`
	Gimbalindex int    `json:"gimbalindex"`
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Product     string `json:"product"`
	SubType     int    `json:"sub_type"`
	Type        int    `json:"type"`
	UpdatedAt   string `json:"updated_at"`
}
