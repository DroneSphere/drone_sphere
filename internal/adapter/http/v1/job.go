package v1

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/model/vo"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
)

type JobRouter struct {
	svc      service.JobSvc
	areaSvc  service.AreaSvc
	modelSvc service.ModelSvc
	l        *slog.Logger
}

func NewJobRouter(handler fiber.Router, svc service.JobSvc, areaSvc service.AreaSvc, modelSvc service.ModelSvc, l *slog.Logger) {
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
		h.Get("/creation/drones", r.getCreationDrones)
		h.Post("/", r.create)
		h.Put("/", r.update)
		h.Delete("/:id", r.delete)
	}
}

func (r *JobRouter) getJobs(c *fiber.Ctx) error {
	ctx := context.Background()
	var params struct {
		JobName           string `query:"job_name"`
		AreaName          string `query:"area_name"`
		ScheduleTimeStart string `query:"schedule_time_start"`
		ScheduleTimeEnd   string `query:"schedule_time_end"`
	}
	if err := c.QueryParser(&params); err != nil {
		return c.JSON(Fail(InvalidParams))
	}
	r.l.Debug("getJobs", "params", params)
	// 将解析到的时间参数传递给仓储层
	jobs, err := r.svc.FetchAll(ctx, params.JobName, params.AreaName, params.ScheduleTimeStart, params.ScheduleTimeEnd)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	var result []struct {
		ID           uint     `json:"id"`
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		AreaName     string   `json:"area_name"`
		ScheduleTime string   `json:"schedule_time"` // 任务计划执行时间
		Drones       []string `json:"drones"`
	}

	for _, job := range jobs {
		var item struct {
			ID           uint     `json:"id"`
			Name         string   `json:"name"`
			Description  string   `json:"description"`
			AreaName     string   `json:"area_name"`
			ScheduleTime string   `json:"schedule_time"` // 任务计划执行时间
			Drones       []string `json:"drones"`
		}
		item.ID = job.ID
		item.Name = job.Name
		item.Description = job.Description
		item.AreaName = job.Area.Name
		item.ScheduleTime = job.ScheduleTime.Format("2006-01-02 15:04:05")
		for _, drone := range job.Drones {
			item.Drones = append(item.Drones, drone.Key)
		}
		result = append(result, item)
	}
	return c.JSON(Success(result))
}

func (r *JobRouter) getJob(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.JSON(Fail(InvalidParams))
	}
	job, err := r.svc.FetchByID(context.Background(), uint(id))
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	var result struct {
		entity.Job
		ScheduleTime string `json:"schedule_time"` // 任务计划执行时间
	}
	if err := copier.Copy(&result, job); err != nil {
		r.l.Error("复制任务数据失败", slog.Any("error", err))
		return c.JSON(Fail(InternalError))
	}
	result.ScheduleTime = job.ScheduleTime.Format("2006-01-02 15:04:05")
	return c.JSON(Success(result))
}

// getCreationOptions  创建任务时依赖的选项数据
func (r *JobRouter) getCreationOptions(c *fiber.Ctx) error {
	ctx := context.Background()
	areas, _, err := r.areaSvc.FetchAll(ctx, "", "", "", 0, 0)
	if err != nil {
		return c.JSON(FailWithMsg(err.Error()))
	}
	droneVariations, err := r.modelSvc.Repo().SelectAllDroneVariation(ctx, nil)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}

	type jobCreationArea struct {
		ID          uint          `json:"id"`
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Points      []vo.GeoPoint `json:"points"`
	}
	var as []jobCreationArea
	for _, area := range areas {
		var e jobCreationArea
		_ = copier.Copy(&e, area)
		as = append(as, e)
	}

	type variantion struct {
		ID               uint             `json:"id"`
		Gimbal           *po.GimbalModel  `json:"gimbal,omitempty"`
		Name             string           `json:"name"`
		Payload          *po.PayloadModel `json:"payload,omitempty"`
		RTKAvailable     bool             `json:"rtk_available"`
		ThermalAvailable bool             `json:"thermal_available"`
	}
	type jobCreationDrone struct {
		ID          uint         `json:"id"`
		Name        string       `json:"name"`
		Description string       `json:"description"`
		Variantions []variantion `json:"variantions"`
	}

	var ds []jobCreationDrone
	var dmap = make(map[uint][]po.DroneVariation)
	for _, dv := range droneVariations {
		if _, ok := dmap[dv.DroneModel.ID]; !ok {
			dmap[dv.DroneModel.ID] = []po.DroneVariation{}
		}
		dmap[dv.DroneModel.ID] = append(dmap[dv.DroneModel.ID], dv)
	}
	for _, variations := range dmap {
		var e jobCreationDrone
		_ = copier.Copy(&e, variations[0].DroneModel)
		for _, variation := range variations {
			var v variantion
			_ = copier.Copy(&v, variation)
			v.Gimbal = nil
			v.Payload = nil
			if len(variation.Gimbals) > 0 {
				v.Gimbal = &variation.Gimbals[0]
			}
			if len(variation.Payloads) > 0 {
				v.Payload = &variation.Payloads[0]
			}
			v.RTKAvailable = variation.SupportsRTK()
			v.ThermalAvailable = variation.SupportsThermal()
			e.Variantions = append(e.Variantions, v)
		}
		ds = append(ds, e)
	}

	result := struct {
		Areas  []jobCreationArea  `json:"areas"`
		Drones []jobCreationDrone `json:"drones"`
	}{
		Areas:  as,
		Drones: ds,
	}
	return c.JSON(Success(result))
}

// getCreationDrones  创建任务时依赖的选项数据
func (r *JobRouter) getCreationDrones(c *fiber.Ctx) error {
	ctx := context.Background()
	drones, err := r.svc.Repo().SelectPhysicalDrones(ctx)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(drones))
}

func (r *JobRouter) create(c *fiber.Ctx) error {
	type Request struct {
		Name                    string                        `json:"name"`
		Description             string                        `json:"description"`
		AreaID                  int64                         `json:"area_id"`
		ScheduleTime            string                        `json:"schedule_time"` // 新增：任务计划执行时间
		Drones                  []po.JobDronePO               `json:"drones"`
		Waylines                []po.JobWaylinePO             `json:"waylines"`
		CommandDrones           []po.JobCommandDronePO        `json:"command_drones"`
		WaylineGenerationParams po.JobWaylineGenerationParams `json:"wayline_generation_params"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(Fail(InvalidParams))
	}

	// 解析时间字符串为当天的时间
	scheduleTime, err := time.Parse("2006-01-02 15:04:05", req.ScheduleTime)
	if err != nil {
		r.l.Error("无效的时间格式", slog.Any("err", err))
		return c.JSON(Fail(InvalidParams))
	}

	r.l.Info("create job", "req", req)
	ctx := context.Background()

	id, err := r.svc.CreateJob(ctx, req.Name, req.Description, uint(req.AreaID), scheduleTime, req.Drones, req.Waylines, req.CommandDrones, req.WaylineGenerationParams)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(id))
}

func (r *JobRouter) update(c *fiber.Ctx) error {
	// 定义接收参数的结构体，与create方法保持一致
	type Request struct {
		ID                      uint                          `json:"id"` // 任务ID
		Name                    string                        `json:"name"`
		Description             string                        `json:"description"`
		AreaID                  int64                         `json:"area_id"`
		ScheduleTime            string                        `json:"schedule_time"` // 新增：任务计划执行时间
		Drones                  []po.JobDronePO               `json:"drones"`
		Waylines                []po.JobWaylinePO             `json:"waylines"`
		CommandDrones           []po.JobCommandDronePO        `json:"command_drones"`
		WaylineGenerationParams po.JobWaylineGenerationParams `json:"wayline_generation_params"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(Fail(InvalidParams))
	}

	// 解析时间字符串为当天的时间
	scheduleTime, err := time.Parse("2006-01-02 15:04:05", req.ScheduleTime)
	if err != nil {
		r.l.Error("无效的时间格式", slog.Any("err", err))
		return c.JSON(Fail(InvalidParams))
	}

	ctx := context.Background()
	job, err := r.svc.ModifyJob(ctx, req.ID, req.Name, req.Description, uint(req.AreaID), scheduleTime, req.Drones, req.Waylines, req.CommandDrones, req.WaylineGenerationParams)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(job))
}

func (r *JobRouter) delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.JSON(Fail(InvalidParams))
	}
	ctx := context.Background()
	if err := r.svc.Repo().DeleteByID(ctx, uint(id)); err != nil {
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(nil))
}
