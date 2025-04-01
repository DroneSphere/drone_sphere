package po

import (
	"github.com/dronesphere/internal/model/vo"
	"gorm.io/datatypes"
)

// Area 搜索区域
type Area struct {
	AreaID      uint                             `json:"area_id" gorm:"primaryKey"`
	CreatedTune int64                            `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTune int64                            `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedTune int64                            `json:"deleted_time" gorm:"autoDeleteTime"`
	State       int                              `json:"area_state" gorm:"default:0"` // -1: deleted, 0: active
	Name        string                           `json:"area_name" gorm:"unique"`
	Description string                           `json:"area_description"`
	CenterLat   float64                          `json:"center_lat"`
	CenterLng   float64                          `json:"center_lng"`
	Points      datatypes.JSONSlice[vo.GeoPoint] `json:"area_points"`
}

// TableName 指定 Area 表名为 tb_areas
func (a Area) TableName() string {
	return "tb_areas"
}
