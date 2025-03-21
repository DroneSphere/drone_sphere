package dto

type JobCreationDrone struct {
	Color       string     `json:"color"`
	Description string     `json:"description"`
	ID          int64      `json:"id"`
	Index       int64      `json:"index"`
	Key         string     `json:"key"`
	Model       string     `json:"model"`
	Name        string     `json:"name"`
	Variantion  Variantion `json:"variantion"`
}

type JobCreationWayline struct {
	DroneKey string  `json:"drone_key"`
	Color    string  `json:"color"`
	Height   float64 `json:"height"`
	Points   []struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"points"`
}

type JobCreationMapping struct {
	PhysicalDroneID  int64  `json:"physicalDroneId"`
	PhysicalDroneSN  string `json:"physicalDroneSN"`
	SelectedDroneKey string `json:"selectedDroneKey"`
}

type Variantion struct {
	Gimbal           Gimbal `json:"gimbal"`
	Index            int64  `json:"index"`
	Name             string `json:"name"`
	RtkAvailable     bool   `json:"rtk_available"`
	ThermalAvailable bool   `json:"thermal_available"`
}

type Gimbal struct {
	Description string `json:"description"`
	Domain      int64  `json:"domain"`
	Gimbalindex int64  `json:"gimbalindex"`
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Product     string `json:"product"`
	SubType     int64  `json:"sub_type"`
	Type        int64  `json:"type"`
	UpdatedAt   string `json:"updated_at"`
}
