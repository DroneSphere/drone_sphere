package v1

import (
	"context"
	"log/slog"
	"strconv"

	api "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
)

type JobRouter struct {
	svc      service.JobSvc
	areaSvc  service.SearchAreaSvc
	modelSvc service.ModelSvc
	l        *slog.Logger
}

func newJobRouter(handler fiber.Router, svc service.JobSvc, areaSvc service.SearchAreaSvc, modelSvc service.ModelSvc, l *slog.Logger) {
	r := &JobRouter{
		svc:      svc,
		areaSvc:  areaSvc,
		modelSvc: modelSvc,
		l:        l,
	}
	h := handler.Group("/job")
	{
		h.Get("/", r.getJobs)
		h.Get("/:id", r.getJob)
		h.Get("/creation/options", r.getCreationOptions)
		h.Get("/edition/:id/options", r.getEditionOptions)
		h.Post("/", r.create)
		h.Put("/", r.update)
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
	ctx := context.Background()
	jobs, err := r.svc.FetchAll(ctx)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	var result []api.JobItemResult
	for _, job := range jobs {
		var item api.JobItemResult
		if err := item.FromJobEntity(job); err != nil {
			return c.JSON(Fail(InternalError))
		}
		result = append(result, item)
	}
	return c.JSON(Success(result))
}

// getJob  获取任务详细信息
//
//	@Router			/job/{id}	[get]
//	@Summary		获取任务详细信息
//	@Description	获取任务详细信息
//	@Tags			job
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=v1.JobDetailResult}	"成功"
func (r *JobRouter) getJob(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.JSON(Fail(InvalidParams))
	}
	job, err := r.svc.FetchByID(context.Background(), uint(id))
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	var result api.JobDetailResult
	if err := result.FromJobEntity(job); err != nil {
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(result))
}

type JobCreationArea struct {
	ID     uint               `json:"id"`
	Name   string             `json:"name"`
	Points []JobCreationPoint `json:"points"`
}

type JobCreationPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type JobCreationOptionDrone struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Model       string       `json:"model"`
	Variantions []Variantion `json:"variantions"`
}

// JobCreationDroneVariationResult
type Variantion struct {
	Gimbal           *po.GimbalModel  `json:"gimbal,omitempty"`
	Index            int              `json:"index"`
	Name             string           `json:"name"`
	Payload          *po.PayloadModel `json:"payload,omitempty"`
	RtkAvailable     bool             `json:"rtk_available"`
	ThermalAvailable bool             `json:"thermal_available"`
}

func (r *JobRouter) getCreationOptions(c *fiber.Ctx) error {
	ctx := context.Background()
	areas, err := r.areaSvc.FetchList(ctx)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	droneModels, err := r.modelSvc.Repo().SelectAllDroneModel(ctx)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	type Result struct {
		Areas  []JobCreationArea        `json:"areas"`
		Drones []JobCreationOptionDrone `json:"drones"`
	}
	var as []JobCreationArea
	for _, area := range areas {
		var a JobCreationArea
		a.ID = area.ID
		a.Name = area.Name
		var points []JobCreationPoint
		for _, p := range area.Points {
			points = append(points, JobCreationPoint{
				Lat: p.Lat,
				Lng: p.Lng,
			})
		}
		a.Points = points
		as = append(as, a)
	}
	var ds []JobCreationOptionDrone
	for _, drone := range droneModels {
		var d JobCreationOptionDrone
		d.Name = drone.Name
		d.Description = drone.Description
		d.ID = drone.ID
		d.Model = drone.Name
		d.Variantions = make([]Variantion, 0)
		for _, g := range drone.Gimbals {
			var v Variantion
			v.Gimbal = &g
			v.Index = g.Gimbalindex
			v.Name = g.Name
			// v.RtkAvailable = drone.is()
			// v.ThermalAvailable = drone.IsThermalAvailable()
			d.Variantions = append(d.Variantions, v)
		}
		// 如果无人机没有云台，则添加一个默认的云台
		if len(drone.Gimbals) == 0 {
			var v Variantion
			v.Gimbal = nil
			v.Index = 0
			v.Name = "默认云台"
			d.Variantions = append(d.Variantions, v)
		}
		ds = append(ds, d)
	}
	// Mock data for creation options
	result := Result{
		Areas:  as,
		Drones: ds,
	}
	return c.JSON(Success(result))
}

// getEditionOptions  编辑任务时依赖的选项数据
//
//	@Router			/job/edition/{id}/options		[get]
//	@Summary		编辑任务时依赖的选项数据
//	@Description	编辑任务时依赖的选项数据，包括可选的无人机列表
//	@Tags			job
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int												true	"任务ID"
//	@Success		200	{object}	v1.Response{data=v1.JobEditionOptionsResult}	"成功"
func (r *JobRouter) getEditionOptions(c *fiber.Ctx) error {
	// 解析Path 中的 ID
	id, err := strconv.Atoi(c.Params("id"))
	ctx := context.Background()
	var drones []entity.Drone
	var j *entity.Job
	drones, err = r.svc.FetchAvailableDrones(ctx)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}

	j, err = r.svc.FetchByID(ctx, uint(id))
	if err != nil {
		return c.JSON(Fail(InternalError))
	}

	var result api.JobEditionOptionsResult
	for _, drone := range drones {
		result.Drones = append(result.Drones, struct {
			ID               uint   `json:"id"`
			Callsign         string `json:"callsign"`
			Description      string `json:"description"`
			SN               string `json:"sn"`
			Model            string `json:"model"`
			RTKAvailable     bool   `json:"rtk_available"`
			ThermalAvailable bool   `json:"thermal_available"`
		}{
			ID:               drone.ID,
			Callsign:         drone.Callsign,
			Description:      "",
			SN:               drone.SN,
			Model:            drone.GetModel(),
			RTKAvailable:     drone.IsRTKAvailable(),
			ThermalAvailable: drone.IsThermalAvailable(),
		})
	}
	result.ID = j.ID
	result.Name = j.Name
	result.Description = j.Description
	points := make([]struct {
		Lat    float64 `json:"lat"`
		Lng    float64 `json:"lng"`
		Marker string  `json:"marker"`
	}, 0)
	for _, p := range j.Area.Points {
		points = append(points, struct {
			Lat    float64 `json:"lat"`
			Lng    float64 `json:"lng"`
			Marker string  `json:"marker"`
		}{
			Lat:    p.Lat,
			Lng:    p.Lng,
			Marker: "",
		})
	}
	result.Area = api.JobAreaResult{
		Name:   j.Area.Name,
		Points: points,
	}
	return c.JSON(Success(result))
}

func (r *JobRouter) create(c *fiber.Ctx) error {
	type Request struct {
		Name        string                   `json:"name"`
		Description string                   `json:"description"`
		AreaID      int64                    `json:"area_id"`
		Drones      []dto.JobCreationDrone   `json:"drones"`
		Waylines    []dto.JobCreationWayline `json:"waylines"`
		Mappings    []dto.JobCreationMapping `json:"mappings"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(Fail(InvalidParams))
	}
	r.l.Info("create job", "req", req)
	ctx := context.Background()
	id, err := r.svc.CreateJob(ctx, req.Name, req.Description, uint(req.AreaID), req.Drones, req.Waylines, req.Mappings)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(id))
}

// update 更新任务
//
//	@Router			/job		[put]
//	@Summary		更新任务
//	@Description	更新任务
//	@Tags			job
//	@Accept			json
//	@Produce		json
//
//	@Param			req	body		v1.JobEditionRequest					true	"更新任务请求"
//
//	@Success		200	{object}	v1.Response{data=v1.JobDetailResult}	"成功"
func (r *JobRouter) update(c *fiber.Ctx) error {
	var req api.JobEditionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(Fail(InvalidParams))
	}
	ctx := context.Background()
	j, err := r.svc.ModifyJob(ctx, req.ID, req.Name, req.Description, req.DroneIDs)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	var result api.JobDetailResult
	if err := result.FromJobEntity(j); err != nil {
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(result))
}
