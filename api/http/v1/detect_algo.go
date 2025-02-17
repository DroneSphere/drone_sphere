package v1

import (
	"github.com/dronesphere/internal/model/entity"
	"github.com/jinzhu/copier"
)

type DetectAlgoResult struct {
	ID          uint                `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Version     string              `json:"algo_version"`
	Path        string              `json:"algo_path"`
	Classes     []DetectClassResult `json:"classes"`
}

type CreateDetectAlgoRequest struct {
	Name        string              `json:"name" example:"demo"`
	Description string              `json:"description" example:"demo"`
	Version     string              `json:"algo_version" example:"1.0.0"`
	Path        string              `json:"algo_path" example:"/demo/path/file.pth"`
	Classes     []DetectClassResult `json:"classes" example:"[{'name':'demo','key':'demo'}]"`
}

type DetectClassResult struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// ToEntity 将 DetectAlgoResult 转换为 entity.DetectAlgo
func (r *CreateDetectAlgoRequest) ToEntity() entity.DetectAlgo {
	var algo entity.DetectAlgo
	if err := copier.Copy(&algo, r); err != nil {
		return entity.DetectAlgo{}
	}
	return algo
}
