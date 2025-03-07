package dto

type MessageCommon struct {
	TID       string `json:"tid"`
	BID       string `json:"bid"`
	Method    string `json:"method"`
	Timestamp int64  `json:"timestamp"`
}

type MessageResultCommon struct {
	MessageCommon
	Data MessageResultData `json:"data"`
}

type MessageResultData struct {
	Result int `json:"result"`
}

func NewMessageResult(res int) MessageResultCommon {
	return MessageResultCommon{
		Data: MessageResultData{
			Result: res,
		},
	}
}

type WSCommon struct {
	BizCode   string `json:"biz_code"`  // 消息业务码，见BizCode常量定义
	Version   string `json:"version"`   // 消息版本号
	Timestamp int64  `json:"timestamp"` // 消息发送时间（毫秒时间戳）
}

const (
	WSBizCodeDeviceOsd        = "device_osd"
	WSBizCodeDeviceOnline     = "device_online"
	WSBizCodeDeviceOffline    = "device_offline"
	WSBizCodeDeviceUpdateTopo = "device_update_topo"
)

var BizCodeMap = map[string]string{
	WSBizCodeDeviceOsd:        "设备遥感信息",
	WSBizCodeDeviceOnline:     "设备上线",
	WSBizCodeDeviceOffline:    "设备下线",
	WSBizCodeDeviceUpdateTopo: "设备拓扑更新",
}
