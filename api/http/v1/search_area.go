package v1

import (
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/vo"
	"github.com/jinzhu/copier"
)

type CreateAreaRequest struct {
	Name        string `json:"name" example:"test"`
	Description string `json:"description" example:"Request for test."`
	Points      []struct {
		Index int     `json:"index"`
		Lat   float64 `json:"lat"`
		Lng   float64 `json:"lng"`
	} `json:"points"`
}

func (r *CreateAreaRequest) ToEntity() *entity.SearchArea {
	var e entity.SearchArea
	if err := copier.Copy(&e, r); err != nil {
		return nil
	}
	var points []vo.GeoPoint
	for _, p := range r.Points {
		points = append(points, vo.GeoPoint{
			Index: p.Index,
			Lat:   p.Lat,
			Lng:   p.Lng,
		})
	}
	e.Points = points
	return &e
}

type AreaFetchParams struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type AreaItemResult struct {
	ID          uint          `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	CenterLat   float64       `json:"center_lat"`
	CenterLng   float64       `json:"center_lng"`
	Points      []PointResult `json:"points"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
}

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
