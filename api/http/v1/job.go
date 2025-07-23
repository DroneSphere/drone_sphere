package v1

import (
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
)

type JobItemResult struct {
	ID           uint     `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	AreaName     string   `json:"area_name"`
	ScheduleTime string   `json:"schedule_time"` // 任务计划执行时间
	Drones       []string `json:"drones"`
}

func (r *JobItemResult) FromJobEntity(j *entity.Job) error {
	r.ID = j.ID
	r.Name = j.Name
	r.Description = j.Description
	r.AreaName = j.Area.Name
	r.ScheduleTime = j.ScheduleTime.Format("15:04:05")
	// for _, m := range j.Mappings {
	// 	r.Drones = append(r.Drones, m.PhysicalDroneCallsign)
	// }
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
	Pitch   float64 `json:"pitch"` // 飞行器俯仰角
	Yaw     float64 `json:"yaw"`   // 飞行器偏航角
	Roll    float64 `json:"roll"`  // 飞行器横滚角
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
	Name         string `json:"name"`
	Description  string `json:"description"`
	AreaID       uint   `json:"area_id"`
	ScheduleTime string `json:"schedule_time"` // 任务计划执行时间，格式：HH:mm:ss
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
	ID           uint   `json:"id"`            // 任务ID
	Name         string `json:"name"`          // 任务名称
	Description  string `json:"description"`   // 任务描述
	ScheduleTime string `json:"schedule_time"` // 任务计划执行时间
	DroneIDs     []uint `json:"drone_ids"`     // 无人机ID列表
}

// JobDetailResult 任务详情
type JobDetailResult struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	ScheduleTime string `json:"schedule_time"` // 任务计划执行时间
	Area         struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Points      []struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"points"`
	} `json:"area"`
	Drones []struct {
		ID          uint                `json:"id"`
		Key         string              `json:"key"`
		Index       int                 `json:"index"`
		Callsign    string              `json:"name"`
		Description string              `json:"description"`
		SN          string              `json:"sn"`
		Model       string              `json:"model"`
		Color       string              `json:"color"`
		Variantion  dto.DroneVariantion `json:"variantion"`
	} `json:"drones"`
	Waylines []struct {
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
	} `json:"waylines"`
	Mappings []struct {
		PhysicalDroneID       uint   `json:"physical_drone_id"`
		PhysicalDroneSN       string `json:"physical_drone_sn"`
		SelectedDroneKey      string `json:"selected_drone_key"`
		PhysicalDroneCallsign string `json:"physical_drone_callsign"`
	} `json:"mappings"`
}

func (r *JobDetailResult) FromJobEntity(j *entity.Job) error {
	// points := make([]struct {
	// 	Lat float64 `json:"lat"`
	// 	Lng float64 `json:"lng"`
	// }, 0)
	// for _, p := range j.Area.Points {
	// 	points = append(points, struct {
	// 		Lat float64 `json:"lat"`
	// 		Lng float64 `json:"lng"`
	// 	}{
	// 		Lat: p.Lat,
	// 		Lng: p.Lng,
	// 	})
	// }
	// r.ID = j.ID
	// r.Name = j.Name
	// r.Description = j.Description
	// r.ScheduleTime = j.ScheduleTime.Format("2006-01-02 15:04:05")
	// r.Area.ID = j.Area.ID
	// r.Area.Name = j.Area.Name
	// r.Area.Description = j.Area.Description
	// r.Area.Points = points
	// for _, d := range j.Drones {
	// 	r.Drones = append(r.Drones, struct {
	// 		ID          uint                `json:"id"`
	// 		Key         string              `json:"key"`
	// 		Index       int                 `json:"index"`
	// 		Callsign    string              `json:"name"`
	// 		Description string              `json:"description"`
	// 		SN          string              `json:"sn"`
	// 		Model       string              `json:"model"`
	// 		Color       string              `json:"color"`
	// 		Variantion  dto.DroneVariantion `json:"variantion"`
	// 	}{

	// 		ID:          d.ID,
	// 		Key:         d.Key,
	// 		Index:       d.Index,
	// 		Callsign:    d.Name,
	// 		Description: d.Description,
	// 		Model:       "",
	// 		Color:       d.Color,
	// 		Variantion:  d.Variantion,
	// 	})
	// }
	// for _, w := range j.Waylines {
	// 	// Convert points to the required format
	// 	points := make([]struct {
	// 		Index int     `json:"index"`
	// 		Lat   float64 `json:"lat"`
	// 		Lng   float64 `json:"lng"`
	// 	}, len(w.Points))

	// 	for i, p := range w.Points {
	// 		points[i] = struct {
	// 			Index int     `json:"index"`
	// 			Lat   float64 `json:"lat"`
	// 			Lng   float64 `json:"lng"`
	// 		}{
	// 			Index: i,
	// 			Lat:   p.Lat,
	// 			Lng:   p.Lng,
	// 		}
	// 	}

	// 	path := make([]struct {
	// 		Lat float64 `json:"lat"`
	// 		Lng float64 `json:"lng"`
	// 	}, len(w.Path))
	// 	for i, p := range w.Path {
	// 		path[i] = struct {
	// 			Lat float64 `json:"lat"`
	// 			Lng float64 `json:"lng"`
	// 		}{
	// 			Lat: p.Lat,
	// 			Lng: p.Lng,
	// 		}
	// 	}

	// 	r.Waylines = append(r.Waylines, struct {
	// 		DroneKey string  `json:"drone_key"`
	// 		Color    string  `json:"color"`
	// 		Height   float64 `json:"height"`
	// 		Path     []struct {
	// 			Lat float64 `json:"lat"`
	// 			Lng float64 `json:"lng"`
	// 		} `json:"path"`
	// 		Points []struct {
	// 			Index int     `json:"index"`
	// 			Lat   float64 `json:"lat"`
	// 			Lng   float64 `json:"lng"`
	// 		} `json:"points"`
	// 	}{
	// 		DroneKey: w.DroneKey,
	// 		Color:    w.Color,
	// 		Height:   w.Height,
	// 		Path:     path,
	// 		Points:   points,
	// 	})
	// }
	// for _, m := range j.Mappings {
	// 	r.Mappings = append(r.Mappings, struct {
	// 		PhysicalDroneID       uint   `json:"physical_drone_id"`
	// 		PhysicalDroneSN       string `json:"physical_drone_sn"`
	// 		SelectedDroneKey      string `json:"selected_drone_key"`
	// 		PhysicalDroneCallsign string `json:"physical_drone_callsign"`
	// 	}{
	// 		PhysicalDroneID:       m.PhysicalDroneID,
	// 		PhysicalDroneSN:       m.PhysicalDroneSN,
	// 		SelectedDroneKey:      m.SelectedDroneKey,
	// 		PhysicalDroneCallsign: m.PhysicalDroneCallsign,
	// 	})
	// }
	return nil
}
