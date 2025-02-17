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
