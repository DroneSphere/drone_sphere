package coordinate

import "testing"

func Test_gcj02ToWGS84(t *testing.T) {
	type Point struct {
		Lng float64
		Lat float64
	}
	data := []Point{
		{117.138148, 36.667246},
		{117.138677, 36.66752},
		{117.139098, 36.666921},
		{117.139098, 36.666921},
	}
	for _, point := range data {
		lng, lat := gcj02ToWGS84(point.Lng, point.Lat)
		t.Logf("gcj02ToWGS84(%f, %f) = (%f, %f)", point.Lng, point.Lat, lng, lat)
	}
}
