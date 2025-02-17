package po

import (
	"github.com/dronesphere/internal/model/vo"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DetectAlgo struct {
	gorm.Model
	Name        string                              `json:"name"`
	Description string                              `json:"description"`
	Version     string                              `json:"algo_version"`
	Path        string                              `json:"algo_path"`
	Classes     datatypes.JSONSlice[vo.DetectClass] `json:"classes"`
}
