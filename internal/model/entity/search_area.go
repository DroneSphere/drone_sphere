package entity

// SearchArea 搜索区域
type SearchArea struct {
	ID          uint        `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	CenterLat   float64     `json:"center_lat"`
	CenterLng   float64     `json:"center_lng"`
	Points      []AreaPoint `json:"points"`
}

// AreaPoint 搜索区域点
type AreaPoint struct {
	Index int     `json:"index"`
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
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
