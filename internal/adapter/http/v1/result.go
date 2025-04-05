package v1

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
)

type ResultRouter struct {
	svc service.ResultSvc
	l   *slog.Logger
}

func newResultRouter(handler fiber.Router, svc service.ResultSvc, l *slog.Logger) {
	r := &ResultRouter{
		svc: svc,
		l:   l,
	}

	h := handler.Group("/results")
	{
		h.Get("/", r.list)
		h.Get("/job_options", r.getJobOptions)
		h.Get("/object_options", r.getObjectTypeOptions)
		h.Get("/:id", r.getByID)
		h.Post("/", r.create)
	}
}

// list 获取检测结果列表
//
//	@Router			/results	[get]
//	@Summary		获取检测结果列表
//	@Description	获取检测结果列表
//	@Tags			results
//	@Accept			json
//	@Produce		json
//	@Param			job_id		query	int		false	"任务ID"
//	@Param			object_type	query	int		false	"物体类型"
//	@Param			page		query	int		true	"页码"
//	@Param			page_size	query	int		true	"每页数量"
//	@Success		200			{object}	Response{data=[]dto.ResultItemDTO}	"成功"
func (r *ResultRouter) list(c *fiber.Ctx) error {
	var query dto.ResultQuery
	if err := c.QueryParser(&query); err != nil {
		return c.JSON(Fail(InvalidParams))
	}

	// 参数校验
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	items, total, err := r.svc.List(context.Background(), query)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(map[string]interface{}{
		"total": total,
		"items": items,
	}))
}

// getJobOptions 获取任务选项
//
//	@Router			/results/job_options	[get]
//	@Summary		获取任务选项
//	@Description	获取任务选项
//	@Tags			results
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Response{data=[]dto.JobOption}	"成功"
func (r *ResultRouter) getJobOptions(c *fiber.Ctx) error {
	options, err := r.svc.GetJobOptions(context.Background())
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(options))
}

// getObjectTypeOptions 获取物体类型选项
//
//	@Router			/results/object_options	[get]
//	@Summary		获取物体类型选项
//	@Description	获取物体类型选项
//	@Tags			results
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Response{data=[]dto.ObjectTypeOption}	"成功"
func (r *ResultRouter) getObjectTypeOptions(c *fiber.Ctx) error {
	options := r.svc.GetObjectTypeOptions(context.Background())
	return c.JSON(Success(options))
}

// getByID 获取检测结果详情
//
//	@Router			/results/{id}	[get]
//	@Summary		获取检测结果详情
//	@Description	获取检测结果详情
//	@Tags			results
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"结果ID"
//	@Success		200	{object}	Response{data=dto.ResultDetailDTO}	"成功"
func (r *ResultRouter) getByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.JSON(Fail(InvalidParams))
	}

	result, err := r.svc.GetByID(context.Background(), uint(id))
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(result))
}

// create 创建检测结果
//
//	@Router			/results	[post]
//	@Summary		创建检测结果
//	@Description	创建检测结果
//	@Tags			results
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.CreateResultDTO	true	"请求参数"
//	@Success		200	{object}	Response{data=map[string]uint}	"成功"
func (r *ResultRouter) create(c *fiber.Ctx) error {
	var req dto.CreateResultDTO
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(Fail(InvalidParams))
	}

	id, err := r.svc.Create(context.Background(), req)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(map[string]uint{"id": id}))
}
