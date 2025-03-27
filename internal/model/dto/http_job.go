package dto

type JobCreationDrone struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Index       int             `json:"index"`
	Key         string          `json:"key"`
	Color       string          `json:"color"`
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
		Index  int     `json:"index"`
		Height float64 `json:"height"`
		Lat    float64 `json:"lat"`
		Lng    float64 `json:"lng"`
	} `json:"points"`
}

type JobCreationMapping struct {
	PhysicalDroneID       uint   `json:"physical_drone_id"`
	PhysicalDroneSN       string `json:"physical_drone_sn"`
	SelectedDroneKey      string `json:"selected_drone_key"`
	PhysicalDroneCallsign string `json:"physical_drone_callsign"`
}

type DroneVariantion struct {
	Name             string `json:"name"`
	Type             int    `json:"type"`
	SubType          int    `json:"sub_type"`
	Index            int    `json:"index"`
	Gimbal           Gimbal `json:"gimbal"`
	RTKAvailable     bool   `json:"rtk_available"`
	ThermalAvailable bool   `json:"thermal_available"`
}

type Gimbal struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      int    `json:"domain"`
	Type        int    `json:"type"`
	SubType     int    `json:"sub_type"`
	Gimbalindex int    `json:"gimbalindex"`
	Product     string `json:"product"`
}
