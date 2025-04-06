package v1

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	api "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/internal/model/dto"
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
		// h.Get("/edition/:id/options", r.getEditionOptions)
		h.Post("/", r.create)
		h.Put("/", r.update)
		h.Delete("/:id", r.delete)
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
	var params struct {
		JobName  string `query:"job_name"`
		AreaName string `query:"area_name"`
	}
	if err := c.QueryParser(&params); err != nil {
		return c.JSON(Fail(InvalidParams))
	}
	r.l.Debug("getJobs", "params", params)
	jobs, err := r.svc.FetchAll(ctx, params.JobName, params.AreaName)
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
	_ = result.FromJobEntity(job)
	return c.JSON(Success(result))
}

// getCreationOptions  创建任务时依赖的选项数据
func (r *JobRouter) getCreationOptions(c *fiber.Ctx) error {
	ctx := context.Background()
	areas, err := r.areaSvc.FetchAll(ctx, "")
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

// // getEditionOptions  编辑任务时依赖的选项数据
// //
// //	@Router			/job/edition/{id}/options		[get]
// //	@Summary		编辑任务时依赖的选项数据
// //	@Description	编辑任务时依赖的选项数据，包括可选的无人机列表
// //	@Tags			job
// //	@Accept			json
// //	@Produce		json
// //	@Param			id	path		int												true	"任务ID"
// //	@Success		200	{object}	v1.Response{data=v1.JobEditionOptionsResult}	"成功"
// func (r *JobRouter) getEditionOptions(c *fiber.Ctx) error {
// 	// 解析Path 中的 ID
// 	id, err := strconv.Atoi(c.Params("id"))
// 	ctx := context.Background()
// 	var drones []entity.Drone
// 	var j *entity.Job
// 	drones, err = r.svc.FetchAvailableDrones(ctx)
// 	if err != nil {
// 		return c.JSON(Fail(InternalError))
// 	}

// 	j, err = r.svc.FetchByID(ctx, uint(id))
// 	if err != nil {
// 		return c.JSON(Fail(InternalError))
// 	}

// 	var result api.JobEditionOptionsResult
// 	for _, drone := range drones {
// 		result.Drones = append(result.Drones, struct {
// 			ID               uint   `json:"id"`
// 			Callsign         string `json:"callsign"`
// 			Description      string `json:"description"`
// 			SN               string `json:"sn"`
// 			Model            string `json:"model"`
// 			RTKAvailable     bool   `json:"rtk_available"`
// 			ThermalAvailable bool   `json:"thermal_available"`
// 		}{
// 			ID:               drone.ID,
// 			Callsign:         drone.Callsign,
// 			Description:      "",
// 			SN:               drone.SN,
// 			Model:            drone.GetModel(),
// 			RTKAvailable:     drone.IsRTKAvailable(),
// 			ThermalAvailable: drone.IsThermalAvailable(),
// 		})
// 	}
// 	result.ID = j.ID
// 	result.Name = j.Name
// 	result.Description = j.Description
// 	points := make([]struct {
// 		Lat    float64 `json:"lat"`
// 		Lng    float64 `json:"lng"`
// 		Marker string  `json:"marker"`
// 	}, 0)
// 	for _, p := range j.Area.Points {
// 		points = append(points, struct {
// 			Lat    float64 `json:"lat"`
// 			Lng    float64 `json:"lng"`
// 			Marker string  `json:"marker"`
// 		}{
// 			Lat:    p.Lat,
// 			Lng:    p.Lng,
// 			Marker: "",
// 		})
// 	}
// 	result.Area = api.JobAreaResult{
// 		Name:   j.Area.Name,
// 		Points: points,
// 	}
// 	return c.JSON(Success(result))
// }

func (r *JobRouter) create(c *fiber.Ctx) error {
	type Request struct {
		Name         string                   `json:"name"`
		Description  string                   `json:"description"`
		AreaID       int64                    `json:"area_id"`
		ScheduleTime string                   `json:"schedule_time"` // 新增：任务计划执行时间
		Drones       []dto.JobCreationDrone   `json:"drones"`
		Waylines     []dto.JobCreationWayline `json:"waylines"`
		Mappings     []dto.JobCreationMapping `json:"mappings"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(Fail(InvalidParams))
	}

	// 解析时间字符串为当天的时间
	scheduleTime, err := time.Parse("15:04:05", req.ScheduleTime)
	if err != nil {
		r.l.Error("无效的时间格式", slog.Any("err", err))
		return c.JSON(Fail(InvalidParams))
	}

	// 获取当前的日期部分
	now := time.Now()
	// 组合当前日期和用户指定的时间
	scheduleTime = time.Date(
		now.Year(), now.Month(), now.Day(),
		scheduleTime.Hour(), scheduleTime.Minute(), scheduleTime.Second(),
		0, now.Location(),
	)

	r.l.Info("create job", "req", req)
	ctx := context.Background()

	id, err := r.svc.CreateJob(ctx, req.Name, req.Description, uint(req.AreaID), scheduleTime, req.Drones, req.Waylines, req.Mappings)
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

	var scheduleTime *time.Time
	if req.ScheduleTime != "" {
		t, err := time.Parse("15:04:05", req.ScheduleTime)
		if err != nil {
			r.l.Error("无效的时间格式", slog.Any("err", err))
			return c.JSON(Fail(InvalidParams))
		}
		// 获取当前的日期部分
		now := time.Now()
		// 组合当前日期和用户指定的时间
		combinedTime := time.Date(
			now.Year(), now.Month(), now.Day(),
			t.Hour(), t.Minute(), t.Second(),
			0, now.Location(),
		)
		scheduleTime = &combinedTime
	}

	ctx := context.Background()
	j, err := r.svc.ModifyJob(ctx, req.ID, req.Name, req.Description, scheduleTime, req.DroneIDs)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	var result api.JobDetailResult
	if err := result.FromJobEntity(j); err != nil {
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(result))
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
