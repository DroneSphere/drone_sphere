package entity

import "github.com/dronesphere/internal/model/vo"

// SearchArea 搜索区域
type SearchArea struct {
	ID          uint          `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	CenterLat   float64       `json:"center_lat"`
	CenterLng   float64       `json:"center_lng"`
	Points      []vo.GeoPoint `json:"points"`
}

func (a *SearchArea) CalcCenter() {
	var lat, lng float64
	for _, p := range a.Points {
		lat += p.Lat
		lng += p.Lng
	}
	a.CenterLat = lat / float64(len(a.Points))
	a.CenterLng = lng / float64(len(a.Points))
}
