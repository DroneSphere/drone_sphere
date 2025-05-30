package vo

type GeoPoint struct {
	Index    int     `json:"index"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Altitude float64 `json:"altitude"` // 高度
}
