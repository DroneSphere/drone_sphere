package v1

type JobItemResult struct {
	ID            uint     `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	AreaName      string   `json:"area_name"`
	Drones        []string `json:"drones"`
	TargetClasses []string `json:"target_classes"`
}

type JobResult struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SubJobResult struct {
	Index int           `json:"index"`
	Area  JobAreaResult `json:"area"`
	Drone JobDrone      `json:"drone"`
}

type JobAreaResult struct {
	Name   string `json:"name"`
	Points []struct {
		Lat    float64 `json:"lat"`
		Lng    float64 `json:"lng"`
		Marker string  `json:"marker"`
	} `json:"points"`
}

type JobDrone struct {
	SN    string `json:"sn"`
	Name  string `json:"name"`
	Model string `json:"model"`
}

type DroneState struct {
	SN      string  `json:"sn"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Height  float64 `json:"height"`
	Heading float64 `json:"heading"`
	Speed   float64 `json:"speed"`
	Battery int     `json:"battery"`
}
