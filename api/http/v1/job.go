package v1

import "github.com/dronesphere/internal/model/entity"

type JobItemResult struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	AreaName    string   `json:"area_name"`
	Drones      []string `json:"drones"`
}

func (r *JobItemResult) FromJobEntity(j *entity.Job) error {
	r.ID = j.ID
	r.Name = j.Name
	r.Description = j.Description
	r.AreaName = j.Area.Name
	for _, d := range j.Drones {
		r.Drones = append(r.Drones, d.Callsign)
	}
	return nil
}

type JobAreaResult struct {
	Name   string `json:"name"`
	Points []struct {
		Lat    float64 `json:"lat"`
		Lng    float64 `json:"lng"`
		Marker string  `json:"marker"`
	} `json:"points"`
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

// JobCreationOptionsResult 创建任务时依赖的选项数据
type JobCreationOptionsResult struct {
	// Area 可选的区域列表
	Areas []struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"areas"`
}

// JobCreationRequest 创建任务请求
type JobCreationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AreaID      uint   `json:"area_id"`
}

// JobCreationResult 创建任务结果
type JobCreationResult struct {
	ID uint `json:"id"`
}

// JobEditionOptionsResult 编辑任务时依赖的选项数据
type JobEditionOptionsResult struct {
	ID          uint          `json:"id"` // 任务ID
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Area        JobAreaResult `json:"area"` // 区域信息
	// Drones 可用的无人机列表
	Drones []struct {
		ID               uint   `json:"id"`
		Callsign         string `json:"callsign"`
		Description      string `json:"description"`
		SN               string `json:"sn"`    // 无人机序列号
		Model            string `json:"model"` // 无人机型号
		RTKAvailable     bool   `json:"rtk_available"`
		ThermalAvailable bool   `json:"thermal_available"` // 是否支持热成像
	} `json:"drones"`
}

// JobEditionRequest 编辑任务请求
type JobEditionRequest struct {
	ID          uint   `json:"id"`          // 任务ID
	Name        string `json:"name"`        // 任务名称
	Description string `json:"description"` // 任务描述
	DroneIDs    []uint `json:"drone_ids"`   // 无人机ID列表
}

// JobDetailResult 任务详情
type JobDetailResult struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Area        struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Points      []struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"points"`
	} `json:"area"`
	Drones []struct {
		ID          uint   `json:"id"`
		Callsign    string `json:"callsign"`
		Description string `json:"description"`
		SN          string `json:"sn"`
		Model       string `json:"model"`
	} `json:"drones"`
}

func (r *JobDetailResult) FromJobEntity(j *entity.Job) error {
	points := make([]struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}, 0)
	for _, p := range j.Area.Points {
		points = append(points, struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		}{
			Lat: p.Lat,
			Lng: p.Lng,
		})
	}
	r.ID = j.ID
	r.Name = j.Name
	r.Description = j.Description
	r.Area.ID = j.Area.ID
	r.Area.Name = j.Area.Name
	r.Area.Description = j.Area.Description
	r.Area.Points = points
	for _, d := range j.Drones {
		r.Drones = append(r.Drones, struct {
			ID          uint   `json:"id"`
			Callsign    string `json:"callsign"`
			Description string `json:"description"`
			SN          string `json:"sn"`
			Model       string `json:"model"`
		}{
			ID:          d.ID,
			Callsign:    d.Callsign,
			Description: "",
			SN:          d.SN,
			Model:       d.GetModel(),
		})
	}
	return nil
}
