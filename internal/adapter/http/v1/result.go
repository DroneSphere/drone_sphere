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
		h.Post("/batch", r.createBatch)
		h.Delete("/:id", r.delete)
	}
}

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

func (r *ResultRouter) getJobOptions(c *fiber.Ctx) error {
	options, err := r.svc.GetJobOptions(context.Background())
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(options))
}

func (r *ResultRouter) getObjectTypeOptions(c *fiber.Ctx) error {
	options, err := r.svc.Repo().GetObjectTypeOptions(context.Background())
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(options))
}

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

func (r *ResultRouter) create(c *fiber.Ctx) error {
	var req dto.CreateResultDTO
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(Fail(InvalidParams))
	}

	id, err := r.svc.Create(context.Background(), req)
	if err != nil {
		return c.JSON(FailWithMsg(err.Error()))
	}

	return c.JSON(Success(map[string]uint{"id": id}))
}

func (r *ResultRouter) createBatch(c *fiber.Ctx) error {
	var req []dto.CreateResultDTO
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(Fail(InvalidParams))
	}

	ids, err := r.svc.CreateBatch(context.Background(), req)
	if err != nil {
		return c.JSON(FailWithMsg(err.Error()))
	}

	return c.JSON(Success(map[string][]uint{"ids": ids}))
}

func (r *ResultRouter) delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.JSON(Fail(InvalidParams))
	}

	err = r.svc.Repo().DeleteByID(context.Background(), uint(id))
	if err != nil {
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(nil))
}
