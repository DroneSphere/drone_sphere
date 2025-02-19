package v1

import (
	api "github.com/dronesphere/api/http/v1"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

type JobRouter struct {
	l *slog.Logger
}

func newJobRouter(handler fiber.Router, l *slog.Logger) {
	r := &JobRouter{
		l: l,
	}
	h := handler.Group("/job")
	{
		h.Get("/", r.getJobs)
		h.Get("/:id", r.getJob)
	}
}

// getJobs 获取所有任务
//
//	@Router			/job		[get]
//	@Summary		获取所有任务
//	@Description	获取所有任务
//	@Tags			job
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=[]v1.JobItemResult}	"成功"
func (r *JobRouter) getJobs(c *fiber.Ctx) error {
	var jobs []api.JobItemResult
	jobs = append(jobs, api.JobItemResult{
		ID:            1,
		Name:          "任务1",
		Description:   "任务1描述",
		AreaName:      "区域1",
		Drones:        []string{"无人机1", "无人机2"},
		TargetClasses: []string{"目标1", "目标2"},
	})
	jobs = append(jobs, api.JobItemResult{
		ID:            2,
		Name:          "任务2",
		Description:   "任务2描述",
		AreaName:      "区域2",
		Drones:        []string{"无人机3"},
		TargetClasses: []string{"目标3"},
	})
	return c.JSON(Success(jobs))
}

// getJob  获取任务详细信息
//
//	@Router			/job/{id}	[get]
//	@Summary		获取任务详细信息
//	@Description	获取任务详细信息
//	@Tags			job
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=[]v1.SubJobResult}	"成功"
func (r *JobRouter) getJob(c *fiber.Ctx) error {
	var items []api.SubJobResult
	items = append(items, api.SubJobResult{
		Index: 1,
		Area: api.JobAreaResult{
			Name: "区域1",
			Points: []struct {
				Lat    float64 `json:"lat"`
				Lng    float64 `json:"lng"`
				Marker string  `json:"marker"`
			}{
				{Lng: 117.138244, Lat: 36.666409, Marker: "marker1"},
				{Lng: 117.139102, Lat: 36.666837, Marker: "marker2"},
				{Lng: 117.13949, Lat: 36.66688, Marker: "marker3"},
				{Lng: 117.140627, Lat: 36.665554, Marker: "marker4"},
				{Lng: 117.13944, Lat: 36.66498, Marker: "marker5"},
			},
		},
		Drone: api.JobDrone{
			SN:    "1581F5FHC246H00DRM66",
			Name:  "无人机1",
			Model: "型号1",
		},
		//}, api.SubJobResult{
		//	Index: 2,
		//	Area: api.JobAreaResult{
		//		Name: "区域2",
		//		Points: []struct {
		//			Lat    float64 `json:"lat"`
		//			Lng    float64 `json:"lng"`
		//			Marker string  `json:"marker"`
		//		}{
		//			{Lat: 2.1, Lng: 2.1, Marker: "marker1"},
		//			{Lat: 2.2, Lng: 2.2, Marker: "marker2"},
		//		},
		//	},
		//	Drone: api.JobDrone{
		//		SN:    "SN2",
		//		Name:  "无人机2",
		//		Model: "型号2",
		//	},
	})
	return c.JSON(Success(items))
}
