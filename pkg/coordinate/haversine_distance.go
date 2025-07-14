package coordinate

import (
	"math"
)

const earthRadius = 6371000 // 地球半径，单位米

// haversin(θ) = sin²(θ/2)
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// HaversineDistance 计算两个经纬度点之间的距离，单位米
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// 将经纬度从度转换为弧度
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	h := hsin(lat2Rad-lat1Rad) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*hsin(lon2Rad-lon1Rad)
	distance := 2 * earthRadius * math.Asin(math.Sqrt(h))
	return distance
}
