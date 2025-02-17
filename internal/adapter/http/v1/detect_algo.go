package v1

import (
	api "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/vo"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
	"log/slog"
	"strconv"
)

type DetectAlgoRouter struct {
	svc service.DetectAlgoSvc
	l   *slog.Logger
}

func newDetectAlgoRouter(handler fiber.Router, svc service.DetectAlgoSvc, l *slog.Logger) {
	r := &DetectAlgoRouter{
		svc: svc,
		l:   l,
	}
	h := handler.Group("/algo")
	{
		h.Get("/", r.getAll)
		h.Get("/:id", r.getByID)
		h.Post("/", r.create)
		h.Put("/:id/classes", r.updateClasses)
		h.Delete("/:id", r.deleteByID)
	}
}

func (r *DetectAlgoRouter) toResult(algo entity.DetectAlgo) api.DetectAlgoResult {
	var res api.DetectAlgoResult
	if err := copier.Copy(&res, &algo); err != nil {
		r.l.Warn("CopyError", slog.Any("err", err))
		return api.DetectAlgoResult{}
	}
	return res
}

// getAll 列出所有检测算法
//
//	@Router			/algo [get]
//	@Summary		列出所有检测算法
//	@Description	列出所有检测算法
//	@Tags			algo
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=[]v1.DetectAlgoResult}	"成功"
func (r *DetectAlgoRouter) getAll(c *fiber.Ctx) error {
	algos, err := r.svc.GetAll()
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	var res []api.DetectAlgoResult
	for _, algo := range algos {
		res = append(res, r.toResult(algo))
	}
	return c.JSON(Success(res))
}

// getByID 获取检测算法
//
//	@Router			/algo/{id} [get]
//	@Summary		获取检测算法
//	@Description	获取检测算法
//	@Tags			algo
//	@Param			id	path	int	true	"算法ID"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=v1.DetectAlgoResult}	"成功"
func (r *DetectAlgoRouter) getByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 400, Msg: "Invalid ID"}))
	}
	algo, err := r.svc.GetByID(uint(id))
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	return c.JSON(Success(r.toResult(algo)))
}

// create 创建检测算法
//
//	@Router			/algo [post]
//	@Summary		创建检测算法
//	@Description	创建检测算法
//	@Tags			algo
//	@Accept			json
//	@Produce		json
//	@Param			req	body		v1.CreateDetectAlgoRequest				true	"请求体"
//	@Success		200	{object}	v1.Response{data=v1.DetectAlgoResult}	"成功"
func (r *DetectAlgoRouter) create(c *fiber.Ctx) error {
	var req api.CreateDetectAlgoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(Fail(ErrorBody{Code: 400, Msg: "Invalid request"}))
	}
	algo := req.ToEntity()
	algo, err := r.svc.Create(algo)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	return c.JSON(Success(r.toResult(algo)))
}

// updateClasses 更新检测算法类别
//
//	@Router			/algo/{id}/classes [put]
//	@Summary		更新检测算法类别
//	@Description	更新检测算法类别
//	@Tags			algo
//	@Param			id	path	int	true	"算法ID"
//	@Accept			json
//	@Produce		json
//	@Param			req	body		[]v1.DetectClassResult						true	"请求体"
//	@Success		200	{object}	v1.Response{data=v1.DetectAlgoResult}	"成功"
func (r *DetectAlgoRouter) updateClasses(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 400, Msg: "Invalid ID"}))
	}
	var classes []vo.DetectClass
	if err := c.BodyParser(&classes); err != nil {
		return c.JSON(Fail(ErrorBody{Code: 400, Msg: "Invalid request"}))
	}
	algo, err := r.svc.UpdateClasses(uint(id), classes)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	return c.JSON(Success(r.toResult(algo)))
}

// deleteByID 删除检测算法
//
//	@Router			/algo/{id} [delete]
//	@Summary		删除检测算法
//	@Description	删除检测算法
//	@Tags			algo
//	@Param			id	path	int	true	"算法ID"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{}	"成功"
func (r *DetectAlgoRouter) deleteByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 400, Msg: "Invalid ID"}))
	}
	err = r.svc.DeleteByID(uint(id))
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	return c.JSON(Success(nil))
}
