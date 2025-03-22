package dto

type PhysicalDrone struct {
	ID       int64            `json:"id"`
	SN       string           `json:"sn"`
	Callsign string           `json:"callsign"`
	Model    PhysicalModel    `json:"model"`
	Gimbals  []PhysicalGimbal `json:"gimbal,omitempty"`
}

type PhysicalModel struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type PhysicalGimbal struct {
	ID         int64  `json:"-"`
	Name       string `json:"name"`
	AircraftID int64  `json:"-"`
}
