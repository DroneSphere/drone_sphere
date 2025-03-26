package v1

type AreaResult struct {
	ID          uint          `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	CenterLat   float64       `json:"center_lat"`
	CenterLng   float64       `json:"center_lng"`
	Points      []PointResult `json:"points"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
}

type PointResult struct {
	Index int     `json:"index"`
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
}
