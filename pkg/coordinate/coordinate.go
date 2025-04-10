package coordinate

import "math"

const (
	pi = 3.1415926535897932384626 // π
	a  = 6378245.0                // 长半轴
	ee = 0.00669342162296594323   // 扁率
)

// 判断坐标是否在中国以外
func outOfChina(lat, lng float64) bool {
	if lng < 72.004 || lng > 137.8347 {
		return true
	}
	if lat < 0.8293 || lat > 55.8271 {
		return true
	}
	return false
}

// 转换纬度
func transformLat(x, y float64) float64 {
	ret := -100.0 + 2.0*x + 3.0*y + 0.2*y*y + 0.1*x*y + 0.2*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*pi) + 20.0*math.Sin(2.0*x*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(y*pi) + 40.0*math.Sin(y/3.0*pi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(y/12.0*pi) + 320*math.Sin(y*pi/30.0)) * 2.0 / 3.0
	return ret
}

// 转换经度
func transformLng(x, y float64) float64 {
	ret := 300.0 + x + 2.0*y + 0.1*x*x + 0.1*x*y + 0.1*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*pi) + 20.0*math.Sin(2.0*x*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(x*pi) + 40.0*math.Sin(x/3.0*pi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(x/12.0*pi) + 300.0*math.Sin(x/30.0*pi)) * 2.0 / 3.0
	return ret
}

// GCJ-02 转 WGS-84
func gcj02ToWGS84(lng, lat float64) (float64, float64) {
	if outOfChina(lat, lng) {
		return lng, lat
	}

	dlat := transformLat(lng-105.0, lat-35.0)
	dlng := transformLng(lng-105.0, lat-35.0)
	radlat := lat / 180.0 * pi
	magic := math.Sin(radlat)
	magic = 1 - ee*magic*magic
	sqrtmagic := math.Sqrt(magic)
	dlat = (dlat * 180.0) / ((a * (1 - ee)) / (magic * sqrtmagic) * pi)
	dlng = (dlng * 180.0) / (a / sqrtmagic * math.Cos(radlat) * pi)
	mglat := lat + dlat
	mglng := lng + dlng
	return lng*2 - mglng, lat*2 - mglat
}

// GCJ02ToWGS84 转换坐标系
func GCJ02ToWGS84(lng, lat float64) (float64, float64) {
	if outOfChina(lat, lng) {
		return lng, lat
	}
	return gcj02ToWGS84(lng, lat)
}

func WGS84ToGCJ02(lng, lat float64) (float64, float64) {
	if outOfChina(lat, lng) {
		return lng, lat
	}
	dlat := transformLat(lng-105.0, lat-35.0)
	dlng := transformLng(lng-105.0, lat-35.0)
	radlat := lat / 180.0 * pi
	magic := math.Sin(radlat)
	magic = 1 - ee*magic*magic
	sqrtmagic := math.Sqrt(magic)
	dlat = (dlat * 180.0) / ((a * (1 - ee)) / (magic * sqrtmagic) * pi)
	dlng = (dlng * 180.0) / (a / sqrtmagic * math.Cos(radlat) * pi)
	mglat := lat + dlat
	mglng := lng + dlng
	return mglng, mglat
}
