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
