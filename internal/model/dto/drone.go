package dto

// DroneModelOption 无人机型号选择项
// 用于前端下拉菜单展示
type DroneModelOption struct {
	ID   uint   `json:"id"`   // 型号 ID
	Name string `json:"name"` // 型号名称
}
