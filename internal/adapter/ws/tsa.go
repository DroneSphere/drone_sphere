package ws

import (
	"log/slog"
	"time"

	"github.com/dronesphere/internal/service"
	"github.com/gofiber/contrib/websocket"
)

func DeviceOSDBroadcast(conn *websocket.Conn, l *slog.Logger, drone service.DroneSvc) {
	tm := time.Now().UnixMicro()
	l.Info("Device OSD", slog.Any("tm", tm))
	// payload := dto.WSDeviceOsdPayload{
	// 	WSCommon: dto.WSCommon{
	// 		BizCode:   dto.WSBizCodeDeviceOsd,
	// 		Version:   "1.0",
	// 		Timestamp: tm,
	// 	},
	// 	Data: struct {
	// 		Host dto.HostInfo `json:"host"`
	// 		SN   string       `json:"sn"`
	// 	}{Host: dto.HostInfo{}, SN: ""},
	// }
	// json, err := sonic.Marshal(payload)
	// if err != nil {
	// 	l.Info("Marshal Failed", slog.Any("err", err))
	// }
	// l.Info("Send", slog.Any("json", json))

	// if err := conn.WriteMessage(1, json); err != nil {
	// 	l.Info("Send Failed:", slog.Any("err", err))
	// }
	l.Info("Send Success")
}
