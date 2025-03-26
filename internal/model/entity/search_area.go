package entity

import (
	"time"

	"github.com/dronesphere/internal/model/vo"
)

// Area 搜索区域
type Area struct {
	ID          uint          `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	CenterLat   float64       `json:"center_lat"`
	CenterLng   float64       `json:"center_lng"`
	Points      []vo.GeoPoint `json:"points"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func (a *Area) CalcCenter() error {
	var lat, lng float64
	for _, p := range a.Points {
		lat += p.Lat
		lng += p.Lng
	}
	a.CenterLat = lat / float64(len(a.Points))
	a.CenterLng = lng / float64(len(a.Points))
	return nil
}
