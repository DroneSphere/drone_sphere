package entity

import (
	"github.com/dronesphere/internal/model/vo"
)

type DetectAlgo struct {
	ID          uint             `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Version     string           `json:"algo_version"`
	Path        string           `json:"algo_path"`
	Classes     []vo.DetectClass `json:"classes"`
}
