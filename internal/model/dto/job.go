package dto

type PhysicalDrone struct {
	ID       int64  `json:"id"`
	SN       string `json:"sn"`
	Callsign string `json:"callsign"`
	Model    struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"model"`
	Gimbals []struct {
		ID   int64  `json:"-"`
		Name string `json:"name"`
	} `json:"gimbal,omitempty"`
}
