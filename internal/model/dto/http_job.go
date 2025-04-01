package dto

type JobCreationDrone struct {
	ID           uint            `json:"id"`                      // ID，前端生成的临时ID
	Name         string          `json:"name"`                    // 名称
	Description  string          `json:"description"`             // 描述
	Index        int             `json:"index"`                   // 无人机序号
	Key          string          `json:"key"`                     // 无人机唯一标识
	Color        string          `json:"color"`                   // 颜色
	ModelID      uint            `json:"model_id,omitempty"`      // 无人机型号ID，新增字段，可选
	VariantionID uint            `json:"variantion_id,omitempty"` // 无人机变体ID，新增字段，可选
	Variantion   DroneVariantion `json:"variantion"`              // 无人机变体信息
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
	PhysicalDroneID       uint   `json:"physical_drone_id"`       // 实际物理无人机ID
	PhysicalDroneSN       string `json:"physical_drone_sn"`       // 实际物理无人机SN
	SelectedDroneKey      string `json:"selected_drone_key"`      // 选择的无人机Key
	PhysicalDroneCallsign string `json:"physical_drone_callsign"` // 实际物理无人机呼号
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
