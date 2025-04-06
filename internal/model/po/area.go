package po

import (
	"time"

	"github.com/dronesphere/internal/model/vo"
	"gorm.io/datatypes"
)

// Area 搜索区域
type Area struct {
	ID          uint                             `json:"area_id" gorm:"primaryKey;column:area_id"`
	CreatedTime time.Time                        `json:"created_time" gorm:"autoCreateTime;column:created_time"`
	UpdatedTime time.Time                        `json:"updated_time" gorm:"autoUpdateTime;column:updated_time"`
	State       int                              `json:"state" gorm:"default:0;column:state"` // -1: deleted, 0: active
	Name        string                           `json:"area_name" gorm:"unique;column:area_name"`
	Description string                           `json:"area_description" gorm:"column:area_description"`
	CenterLat   float64                          `json:"center_lat" gorm:"column:center_lat"`
	CenterLng   float64                          `json:"center_lng" gorm:"column:center_lng"`
	Points      datatypes.JSONSlice[vo.GeoPoint] `json:"area_points" gorm:"column:area_points"`
}

// TableName 指定 Area 表名为 tb_areas
func (a Area) TableName() string {
	return "tb_areas"
}
