package dji

import (
	"log/slog"

	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
)

type WaylineRouter struct {
	svc service.WaylineSvc
	eb  EventBus.Bus
	l   *slog.Logger
}

func NewWaylineRouter(handler fiber.Router, svc service.WaylineSvc, eb EventBus.Bus, l *slog.Logger) {
	r := &WaylineRouter{
		svc: svc,
		eb:  eb,
		l:   l,
	}
	h := handler.Group("/wayline/api/v1")
	{
		h.Get("/workspaces/:workspace_id/waylines", r.getWaylines)
		// /wayline/api/v1/workspaces/{workspace_id}/waylines/{id}/url
		h.Get("/workspaces/:workspace_id/waylines/:id/url", r.getWaylineURL)

	}
}

func (r *WaylineRouter) getWaylines(c *fiber.Ctx) error {
	workspaceID := c.Params("workspace_id")
	r.l.Info("getWaylines", slog.Any("workspaceID", workspaceID))
	type StartPoint struct {
		StartLatitude  float64 `json:"start_latitude" gorm:"column:start_latitude"`
		StartLontitude float64 `json:"start_lontitude" gorm:"column:start_lontitude"`
	}
	type WaylineItemResult struct {
		ID                string     `json:"id" gorm:"column:id"`
		Name              string     `json:"name" gorm:"column:name"`
		Username          string     `json:"username" gorm:"column:username"`
		UpdateTime        int64      `json:"update_time" gorm:"column:update_time"`
		CreateTime        int64      `json:"create_time" gorm:"column:create_time"`
		DroneModelKey     string     `json:"drone_model_key" gorm:"column:drone_model_key"`
		PayloadModelKeys  []string   `json:"payload_model_keys" gorm:"column:payload_model_keys;type:json"`
		Favorited         bool       `json:"favorited" gorm:"column:favorited"`
		TemplateTypes     []int      `json:"template_types" gorm:"column:template_types;type:json"`
		ActionType        int        `json:"action_type" gorm:"column:action_type"`
		StartWaylinePoint StartPoint `json:"start_wayline_point" gorm:"column:start_wayline_point;type:json"`
	}

	waylines, err := r.svc.Repo().SelectAll(c.Context())
	if err != nil {
		r.l.Error("Failed to fetch waylines", slog.Any("err", err))
		return c.JSON(Fail(InternalError))
	}
	var list []WaylineItemResult
	for _, w := range waylines {
		var e WaylineItemResult
		if err := copier.Copy(&e, &w); err != nil {
			r.l.Error("Failed to copy wayline", slog.Any("err", err))
			return c.JSON(Fail(InternalError))
		}
		// 根据原有 ID 生成 UUID 作为返回的 ID
		// e.ID = w.UUID
		e.Username = "admin"
		e.UpdateTime = w.UpdatedTime.UnixMilli()
		e.CreateTime = w.CreatedTime.UnixMilli()
		// e.StartWaylinePoint = StartPoint(w.StartWaylinePoint.Data())

		list = append(list, e)
	}

	type Pagination struct {
		Page     int `json:"page"`
		PageSize int `json:"page_size"`
		Total    int `json:"total"`
	}
	type Result struct {
		List       []WaylineItemResult `json:"list"`
		Pagination Pagination          `json:"pagination"`
	}
	res := &Result{
		List: list,
		Pagination: Pagination{
			Page:     1,
			PageSize: 10,
			Total:    len(list),
		},
	}
	return c.JSON(Success(res))
}

func (r *WaylineRouter) getWaylineURL(c *fiber.Ctx) error {
	workspaceID := c.Params("workspace_id")
	waylineID := c.Params("id")
	r.l.Info("getWaylineURL", slog.Any("workspaceID", workspaceID), slog.Any("waylineID", waylineID))

	url, err := r.svc.GetWaylineURL(c.Context(), workspaceID, waylineID)
	if err != nil {
		r.l.Error("Failed to get wayline URL", slog.Any("err", err))
		return c.JSON(Fail(InternalError))
	}
	r.l.Info("getWaylineURL", slog.Any("url", url))

	return c.JSON(Success(url))
}
