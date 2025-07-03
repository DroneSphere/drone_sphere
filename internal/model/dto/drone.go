package dto

// DroneModelOption 无人机型号选择项
// 用于前端下拉菜单展示
type DroneModelOption struct {
	ID   uint   `json:"id"`   // 型号 ID
	Name string `json:"name"` // 型号名称
}

type WSbaseModel struct {
	TID       string      `json:"tid"`       // 事务 ID, UUID
	Timestamp int64       `json:"timestamp"` // 时间戳, 秒
	Method    string      `json:"method"`    // 方法名
	Data      interface{} `json:"data"`      // 数据
}
