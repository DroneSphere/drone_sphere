package v1

import (
	"github.com/dronesphere/internal/model/entity"
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
	var points []entity.AreaPoint
	for _, p := range r.Points {
		points = append(points, entity.AreaPoint{
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

type AreaListResult struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CenterLat   float64 `json:"center_lat"`
	CenterLng   float64 `json:"center_lng"`
}

type AreaResult struct {
	ID          uint          `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	CenterLat   float64       `json:"center_lat"`
	CenterLng   float64       `json:"center_lng"`
	Points      []PointResult `json:"points"`
}

type PointResult struct {
	Index int     `json:"index"`
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
}
