package po

import (
	"github.com/dronesphere/internal/model/vo"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Area 搜索区域
type Area struct {
	gorm.Model
	Name        string                           `json:"name" gorm:"unique"`
	Description string                           `json:"description"`
	CenterLat   float64                          `json:"center_lat"`
	CenterLng   float64                          `json:"center_lng"`
	Points      datatypes.JSONSlice[vo.GeoPoint] `json:"points"`
}
