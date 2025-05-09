package v1

import (
	"log/slog"

	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
)

type WaylineRouter struct {
	svc service.WaylineSvc
	l   *slog.Logger
}

func NewWaylineRouter(handler fiber.Router, svc service.WaylineSvc, l *slog.Logger) {
	r := &WaylineRouter{
		svc: svc,
		l:   l,
	}
	h := handler.Group("/wayline")
	{
		h.Get("", r.GetWaylineByJobIDAndDroneSN)
	}
}

func (r *WaylineRouter) GetWaylineByJobIDAndDroneSN(c *fiber.Ctx) error {
	var req struct {
		JobID   uint   `query:"job_id"`
		DroneSN string `query:"drone_sn"`
	}

	// 解析请求参数
	if err := c.QueryParser(&req); err != nil {
		r.l.Error("解析请求参数失败", "err", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": "请求参数无效",
		})
	}

	// 参数验证
	if req.JobID == 0 || req.DroneSN == "" {
		r.l.Error("请求参数不完整", "job_id", req.JobID, "drone_sn", req.DroneSN)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": "作业ID和无人机序列号不能为空",
		})
	}

	// 调用服务层方法获取航线URL
	wayline, err := r.svc.FetchWaylineByJobIDAndDroneSN(c.Context(), req.JobID, req.DroneSN)
	if err != nil {
		r.l.Error("获取航线U失败", "err", err, "job_id", req.JobID, "drone_sn", req.DroneSN)
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	// 返回成功响应
	return c.JSON(Success(wayline))
}
