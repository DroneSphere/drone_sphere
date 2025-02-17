package entity

type Job struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Area        SearchArea `json:"area"`
	Algo        DetectAlgo `json:"algo"`
	Drones      []Drone    `json:"drones"`
}
