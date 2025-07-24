package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/model/ro"
	"github.com/dronesphere/internal/repo"
	"github.com/dronesphere/pkg/txlive"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
)

func init() {
	// 初始化连接存储
	connections = make(map[string][]*websocket.Conn)
	mutex = sync.RWMutex{}
}

type (
	DroneSvc interface {
		Repo() DroneRepo
		SaveDroneTopo(ctx context.Context, update dto.UpdateTopoPayload) error
		FetchDeviceTopo(ctx context.Context, workspace string) ([]entity.Drone, []entity.RC, error)
		UpdateStateBySN(ctx context.Context, sn string, msg dto.DroneMessageProperty) error
		// 新增：从消息创建无人机实体
		CreateDroneFromMsg(ctx context.Context, sn string, msg dto.ProductTopo, modelRepo *repo.ModelDefaultRepo) (*entity.Drone, error)
		// 新增：根据无人机SN启动直播
		StartLiveBySN(ctx context.Context, sn string) (string, error)
		// 新增：根据无人机SN停止直播
		StopLiveBySN(ctx context.Context, sn string) error
		CheckControlConnection(ctx context.Context, conn *websocket.Conn, sn string) error
		HandleControlSession(ctx context.Context, conn *websocket.Conn, sn string, mt int, msg string) error
	}

	DroneRepo interface {
		SelectAll(ctx context.Context, sn string, callsign string, modelID uint, page, pageSize int) ([]entity.Drone, int64, error)
		Save(ctx context.Context, d entity.Drone) error
		SelectBySN(ctx context.Context, sn string) (entity.Drone, error)
		SelectByID(ctx context.Context, id uint) (entity.Drone, error)
		SelectByIDV2(ctx context.Context, id uint) (*po.Drone, error)
		FetchStateBySN(ctx context.Context, sn string) (ro.Drone, error)
		SaveState(ctx context.Context, state ro.Drone) error
		SelectAllByID(ctx context.Context, ids []uint) ([]entity.Drone, error)
		UpdateDroneInfo(ctx context.Context, sn string, updates map[string]interface{}) error
		FetchDroneModelOptions(ctx context.Context) ([]dto.DroneModelOption, error)                 // 获取无人机型号选项列表
		UpdateLiveInfoBySN(ctx context.Context, sn, pushRTMPUrl, pullRTMPUrl, videoID string) error // 修改签名以包含 videoID
		FetchGatewaySNByDroneSN(ctx context.Context, droneSN string) (string, error)                // 新增获取网关SN的方法
	}
)

type DroneImpl struct {
	r         DroneRepo
	modelRepo ModelRepo
	l         *slog.Logger
	mqtt      mqtt.Client
}

func NewDroneImpl(r DroneRepo, modelRepo ModelRepo, l *slog.Logger, mqtt mqtt.Client) DroneSvc {
	return &DroneImpl{
		r:         r,
		modelRepo: modelRepo,
		l:         l,
		mqtt:      mqtt,
	}
}

func (s *DroneImpl) Repo() DroneRepo {
	return s.r
}

const pushDomain = "140433.livepush.myqcloud.com"
const pullDomain = "lisoft.com.cn"
const pushKey = "61c3e60eb8cf33bd4b4fb6a504fb51df"
const pullKey = "drone"

// StartLiveBySN 根据无人机SN启动直播并返回拉流地址
// 此方法会为指定无人机生成推流和拉流地址，更新到数据库，并发送MQTT指令启动设备推流
func (s *DroneImpl) StartLiveBySN(ctx context.Context, sn string) (string, error) {
	// 获取无人机信息
	drone, err := s.r.SelectBySN(ctx, sn)
	if err != nil {
		s.l.Error("获取无人机信息失败", slog.String("sn", sn), slog.Any("error", err))
		return "", fmt.Errorf("获取无人机信息失败: %w", err)
	}
	droneModel, err := s.modelRepo.SelectDroneModelByID(ctx, drone.DroneModelID)
	if err != nil {
		s.l.Error("获取无人机信息失败", slog.String("sn", sn), slog.Any("error", err))
		return "", fmt.Errorf("获取无人机信息失败: %w", err)
	}

	// 检查无人机是否有云台信息
	if len(droneModel.Gimbals) == 0 {
		s.l.Error("无人机没有云台信息", slog.String("sn", sn))
		return "", fmt.Errorf("无人机 %s 没有云台信息", sn)
	}

	// 获取第一个云台信息
	gimbal := droneModel.Gimbals[0]

	// 构建直播所需参数
	cameraIndex := strconv.Itoa(gimbal.Type) + "-" + strconv.Itoa(gimbal.SubType) + "-" + strconv.Itoa(gimbal.Gimbalindex)
	videoIndex := "normal-0" // 普通视频流索引
	streamName := fmt.Sprintf("%s-%s", sn, cameraIndex)
	videoID := fmt.Sprintf("%s/%s/%s", sn, cameraIndex, videoIndex)

	// 设置过期时间为24小时后
	expireTime := time.Now().Add(24 * time.Hour)
	expireTimeHex := strconv.FormatInt(expireTime.Unix(), 16)

	// 构建推流和拉流地址
	pushRTMPUrl := txlive.BuildRTMPUrl(pushDomain, streamName, pushKey, expireTimeHex)
	pullRTMPUrl := txlive.BuildRTMPUrl(pullDomain, streamName, pullKey, expireTimeHex)

	s.l.Info("生成直播地址成功",
		slog.String("sn", sn),
		slog.String("push_url", pushRTMPUrl),
		slog.String("pull_url", pullRTMPUrl),
		slog.String("video_id", videoID),
		slog.String("expire_time", expireTime.Format(time.RFC3339)),
	)

	// 更新无人机的直播信息到数据库
	if err := s.r.UpdateLiveInfoBySN(ctx, sn, pushRTMPUrl, pullRTMPUrl, videoID); err != nil {
		s.l.Error("更新无人机直播信息到数据库失败", slog.String("sn", sn), slog.Any("error", err))
		return "", fmt.Errorf("更新无人机直播信息到数据库失败: %w", err)
	}

	// 获取 GatewaySN
	gatewaySN, err := s.r.FetchGatewaySNByDroneSN(ctx, sn) // 使用仓库层方法获取
	// 从无人机实体中获取 GwSN (在 repo.FetchStateBySN 中填充)
	if gatewaySN == "" || err != nil {
		s.l.Error("未能获取无人机关联的GatewaySN", slog.String("sn", sn))
		// 注意：根据业务需求，这里可能需要返回错误，或者允许在没有GatewaySN的情况下继续（例如，仅更新数据库而不发送MQTT）
		// 当前实现为继续执行，但实际直播可能无法启动
		return "", fmt.Errorf("未能获取无人机关联的GatewaySN，无法发送MQTT指令")
	}

	if gatewaySN != "" {
		// 构造 MQTT 消息
		liveStartData := dto.LiveStartPushData{
			URL:          pushRTMPUrl, // 设备端使用推流地址
			URLType:      1,           // 1 代表 RTMP
			VideoID:      videoID,
			VideoQuality: 0, // 0 代表自适应
		}
		mqttRequest := dto.LiveStartPushRequest{
			BID:       uuid.New().String(),
			Data:      liveStartData,
			TID:       uuid.New().String(),
			Timestamp: time.Now().UnixMilli(),
			Method:    "live_start_push",
		}
		payload, marshalErr := json.Marshal(mqttRequest)
		if marshalErr != nil {
			s.l.Error("序列化MQTT live_start_push 请求失败", slog.String("sn", sn), slog.Any("error", marshalErr))
			return "", fmt.Errorf("序列化MQTT live_start_push请求失败: %w", marshalErr)
		}

		topic := fmt.Sprintf("thing/product/%s/services", gatewaySN)
		s.l.Info("准备发送MQTT live_start_push 消息", slog.String("topic", topic), slog.String("payload", string(payload)))

		token := s.mqtt.Publish(topic, 1, false, payload)
		if token.WaitTimeout(5*time.Second) && token.Error() != nil {
			s.l.Error("发送MQTT live_start_push 消息失败", slog.String("sn", sn), slog.String("topic", topic), slog.Any("error", token.Error()))
			return "", fmt.Errorf("发送MQTT live_start_push消息失败: %w", token.Error())
		} else if !token.WaitTimeout(5 * time.Second) {
			s.l.Error("发送MQTT live_start_push 消息超时", slog.String("sn", sn), slog.String("topic", topic))
			return "", fmt.Errorf("发送MQTT live_start_push消息超时")
		}
		s.l.Info("MQTT live_start_push 消息发送成功", slog.String("sn", sn), slog.String("topic", topic))

		s.l.Info("启动无人机直播流程成功", slog.String("sn", sn))
		return pullRTMPUrl, nil
	} else {
		s.l.Warn("GatewaySN为空，跳过发送MQTT live_start_push消息", slog.String("sn", sn))
		return "", fmt.Errorf("GatewaySN为空，无法发送MQTT指令")
	}

}

// StopLiveBySN 根据无人机SN停止直播
// 此方法会发送MQTT指令停止设备推流，并清空数据库中的直播信息
func (s *DroneImpl) StopLiveBySN(ctx context.Context, sn string) error {
	// 获取无人机信息，特别是 CurrentVideoID 和 GatewaySN
	drone, err := s.r.SelectBySN(ctx, sn)
	if err != nil {
		s.l.Error("获取无人机信息失败", slog.String("sn", sn), slog.Any("error", err))
		return fmt.Errorf("获取无人机信息失败: %w", err)
	}

	if drone.CurrentVideoID == "" {
		s.l.Info("无人机当前没有正在进行的直播", slog.String("sn", sn))
		return nil // 或者根据业务返回特定错误，例如 ErrNoActiveStream
	}

	gatewaySN, err := s.r.FetchGatewaySNByDroneSN(ctx, sn) // 使用仓库层方法获取
	if err != nil {
		// handle
		s.l.Error("获取无人机关联的GatewaySN失败", slog.String("sn", sn), slog.Any("error", err))
		return fmt.Errorf("获取无人机关联的GatewaySN失败: %w", err)
	}
	// 从无人机实体中获取 GwSN (在 repo.FetchStateBySN 中填充)
	if gatewaySN == "" {
		s.l.Error("未能获取无人机关联的GatewaySN", slog.String("sn", sn))
		// 注意：根据业务需求，这里可能需要返回错误
		// return fmt.Errorf("未能获取无人机关联的GatewaySN，无法发送MQTT指令")
		s.l.Info("无人机没有关联的网关，跳过发送MQTT指令", slog.String("sn", sn))
		return nil
	}

	if gatewaySN != "" {
		// 构造 MQTT 消息
		liveStopData := dto.LiveStopPushData{
			VideoID: drone.CurrentVideoID,
		}
		mqttRequest := dto.LiveStopPushRequest{
			BID:       uuid.New().String(),
			Data:      liveStopData,
			TID:       uuid.New().String(),
			Timestamp: time.Now().UnixMilli(),
			Method:    "live_stop_push",
		}
		payload, marshalErr := json.Marshal(mqttRequest)
		if marshalErr != nil {
			s.l.Error("序列化MQTT live_stop_push 请求失败", slog.String("sn", sn), slog.Any("error", marshalErr))
			return fmt.Errorf("序列化MQTT live_stop_push请求失败: %w", marshalErr)
		}

		topic := fmt.Sprintf("thing/product/%s/services", gatewaySN)
		s.l.Info("准备发送MQTT live_stop_push 消息", slog.String("topic", topic), slog.String("payload", string(payload)))

		token := s.mqtt.Publish(topic, 1, false, payload)
		if token.WaitTimeout(5*time.Second) && token.Error() != nil {
			s.l.Error("发送MQTT live_stop_push 消息失败", slog.String("sn", sn), slog.String("topic", topic), slog.Any("error", token.Error()))
			return fmt.Errorf("发送MQTT live_stop_push消息失败: %w", token.Error())
		} else if !token.WaitTimeout(5 * time.Second) {
			s.l.Error("发送MQTT live_stop_push 消息超时", slog.String("sn", sn), slog.String("topic", topic))
			return fmt.Errorf("发送MQTT live_stop_push消息超时")
		}
		s.l.Info("MQTT live_stop_push 消息发送成功", slog.String("sn", sn), slog.String("topic", topic))
	} else {
		s.l.Warn("GatewaySN为空，跳过发送MQTT live_stop_push消息", slog.String("sn", sn))
	}

	// 清空数据库中的直播信息
	if err := s.r.UpdateLiveInfoBySN(ctx, sn, "", "", ""); err != nil {
		s.l.Error("清空无人机直播信息失败", slog.String("sn", sn), slog.Any("error", err))
		return fmt.Errorf("清空无人机直播信息失败: %w", err)
	}

	s.l.Info("停止无人机直播流程成功", slog.String("sn", sn))
	return nil
}

func (s *DroneImpl) SaveDroneTopo(ctx context.Context, data dto.UpdateTopoPayload) error {
	rc := ctx.Value(dto.SNKey).(string)
	s.l.Info("SaveDroneTopo", slog.Any("data", data), slog.Any("rc", rc))
	// 如果没有子设备，按照遥控器SN删除无人机
	if len(data.SubDevices) == 0 {
		s.l.Info("SubDevices is empty, remove drone", slog.Any("rc", rc))
		return nil
	}

	// 保存无人机信息
	subDevice := data.SubDevices[0]
	drone := entity.Drone{
		SN:      subDevice.SN,
		Type:    subDevice.Type,
		SubType: subDevice.SubType,
	}
	s.l.Info("SaveDroneTopo", slog.Any("data", data))
	if err := s.r.Save(ctx, drone); err != nil {
		s.l.Error("SaveDroneTopo failed", slog.Any("err", err))
		return err
	}
	s.l.Info("SaveDroneTopo success", slog.Any("drone", drone))

	return nil
}

func (s *DroneImpl) FetchDeviceTopo(ctx context.Context, workspace string) ([]entity.Drone, []entity.RC, error) {
	var ds []entity.Drone
	var rcs []entity.RC
	//dds, rccs, err := s.r.FetchDeviceTopoByWorkspace(ctx, workspace)
	//if err != nil {
	//	return nil, nil, err
	//}
	//for _, d := range dds {
	//	var e entity.Drone
	//	if err := copier.Copy(&e, &d); err != nil {
	//		s.l.Error("SelectAll copier failed", slog.Any("err", err))
	//		return nil, nil, err
	//	}
	//	ds = append(ds, e)
	//}
	//for _, rc := range rccs {
	//	var e entity.RC
	//	if err := copier.Copy(&e, &rc); err != nil {
	//		s.l.Error("SelectAll copier failed", slog.Any("err", err))
	//		return nil, nil, err
	//	}
	//	rcs = append(rcs, e)
	//}
	return ds, rcs, nil
}

// UpdateStateBySN 更新无人机实时数据状态
func (s *DroneImpl) UpdateStateBySN(ctx context.Context, sn string, msg dto.DroneMessageProperty) error {
	var state = ro.Drone{
		SN:                   sn,
		Status:               ro.DroneStatusOnline,
		DroneMessageProperty: msg,
	}
	if err := s.r.SaveState(ctx, state); err != nil {
		s.l.Error("Save realtime drone failed", slog.Any("err", err))
		return err
	}
	return nil
}

// CreateDroneFromMsg 从消息创建无人机实体
// 该方法将原本位于 entity.Drone 中的 NewDroneFromMsg 功能提升到服务层
// 避免了 entity 包对 repo 包的循环引用问题
func (s *DroneImpl) CreateDroneFromMsg(ctx context.Context, sn string, msg dto.ProductTopo, modelRepo *repo.ModelDefaultRepo) (*entity.Drone, error) {
	// 创建基本的无人机实体
	d := &entity.Drone{
		SN:       sn,
		Status:   ro.DroneStatusOnline, // 新接入的无人机默认为在线状态
		Callsign: sn[:8],               // 使用SN前8位作为默认呼号
	}

	// 使用 modelRepo 查询匹配的无人机型号
	if modelRepo != nil {
		// 找到匹配的无人机型号
		droneModel, err := findDroneModelByDomainTypeSubType(ctx, modelRepo, "0", msg.Type, msg.SubType)
		if err == nil && droneModel != nil {
			// 找到匹配的无人机型号
			d.DroneModelID = droneModel.ID
			d.DroneModel = *droneModel
			s.l.Info("无人机型号匹配成功",
				slog.String("sn", sn),
				slog.String("model_name", droneModel.Name),
				slog.Uint64("model_id", uint64(droneModel.ID)))

			// 查询该型号的默认变体
			variation, err := findDefaultDroneVariation(ctx, modelRepo, droneModel.ID)
			if err == nil && variation != nil {
				d.VariationID = variation.ID
				d.Variation = *variation
				s.l.Info("无人机默认变体匹配成功",
					slog.String("sn", sn),
					slog.String("model_name", droneModel.Name),
					slog.String("variation_name", variation.Name))
			} else {
				s.l.Warn("无人机默认变体匹配失败",
					slog.String("sn", sn),
					slog.Uint64("model_id", uint64(droneModel.ID)),
					slog.Any("error", err))
			}
		} else {
			// 未找到匹配的无人机型号
			s.l.Warn("无人机型号匹配失败",
				slog.String("sn", sn),
				slog.String("domain", msg.Domain),
				slog.Int64("type", int64(msg.Type)),
				slog.Int64("sub_type", int64(msg.SubType)),
				slog.Any("error", err))
		}
	}

	return d, nil
}

// findDroneModelByDomainTypeSubType 是内部辅助方法，根据 Domain、Type、SubType 查找无人机型号
// 将此功能从 modelRepo 中提取到服务层，作为内部实现
func findDroneModelByDomainTypeSubType(ctx context.Context, modelRepo *repo.ModelDefaultRepo, domain string, deviceType int, subType int) (*po.DroneModel, error) {
	// 将 domain 字符串转换为整数
	// var domainInt int
	// var err error

	// // 兼容不同类型的 domain 格式
	// if domain != "" {
	// 	domainInt, err = strconv.Atoi(domain)
	// 	if err != nil {
	// 		// 默认为 0，表示未知领域
	// 		domainInt = 0
	// 	}
	// }

	// 在数据库中查询匹配的无人机型号
	return modelRepo.FindDroneModelByDomainTypeSubType(ctx, domain, deviceType, subType)
}

// findDefaultDroneVariation 是内部辅助方法，查询指定型号的默认变体
func findDefaultDroneVariation(ctx context.Context, modelRepo *repo.ModelDefaultRepo, droneModelID uint) (*po.DroneVariation, error) {
	return modelRepo.FindDefaultDroneVariation(ctx, droneModelID)
}

// connections 存储 SN 对应的多个连接
var connections map[string][]*websocket.Conn

// mutex 保护并发访问
var mutex sync.RWMutex

// AddConnection 为指定SN添加连接
func (dcm *DroneImpl) AddConnection(sn string, conn *websocket.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	connections[sn] = append(connections[sn], conn)
}

// RemoveConnection 移除指定SN的连接
func (dcm *DroneImpl) RemoveConnection(sn string, conn *websocket.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	conns := connections[sn]
	for i, c := range conns {
		if c == conn {
			// 从切片中移除该连接
			conns[i] = conns[len(conns)-1] // 将最后一个元素移到当前位置
			conns = conns[:len(conns)-1]   // 缩短切片长度
			break
		}
	}

	// 如果该SN没有连接了，删除该键
	if len(conns) == 0 {
		delete(connections, sn)
	}
}

func (dcm *DroneImpl) IsConnectionExists(sn string, conn *websocket.Conn) bool {
	mutex.RLock()
	defer mutex.RUnlock()

	conns, exists := connections[sn]
	if !exists {
		return false
	}

	return slices.Contains(conns, conn)
}

// GetConnections 获取指定SN的所有连接
func (dcm *DroneImpl) GetConnections(sn string) []*websocket.Conn {
	mutex.RLock()
	defer mutex.RUnlock()

	conns := connections[sn]
	// 返回副本避免并发问题
	result := make([]*websocket.Conn, len(conns))
	copy(result, conns)
	return result
}

// BroadcastToSN 向指定SN的所有连接广播消息
func (dcm *DroneImpl) BroadcastToSN(sn string, messageType int, data string) {
	conns := dcm.GetConnections(sn)

	for _, conn := range conns {
		if err := conn.WriteMessage(messageType, []byte(data)); err != nil {
			// 发送失败，移除该连接
			dcm.RemoveConnection(sn, conn)
		}
	}
}

func (dcm *DroneImpl) ReplyToConn(sn string, conn *websocket.Conn, msg string) {
	if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		dcm.l.Error("回应消息发送失败", slog.String("sn", sn), slog.Any("error", err))
		// 发送失败，移除该连接
		dcm.RemoveConnection(sn, conn)
	} else {
		dcm.l.Info("回应消息发送成功", slog.String("sn", sn))
	}
}

// GetAllSNs 获取所有有连接的SN列表
func (dcm *DroneImpl) GetAllSNs() []string {
	mutex.RLock()
	defer mutex.RUnlock()

	sns := make([]string, 0, len(connections))
	for sn := range connections {
		sns = append(sns, sn)
	}
	return sns
}

// GetConnectionCount 获取指定SN的连接数量
func (dcm *DroneImpl) GetConnectionCount(sn string) int {
	mutex.RLock()
	defer mutex.RUnlock()

	return len(connections[sn])
}

func (s *DroneImpl) CheckControlConnection(ctx context.Context, conn *websocket.Conn, sn string) error {
	if s.IsConnectionExists(sn, conn) {
		s.l.Info("连接已存在，继续处理", slog.String("sn", sn))
	} else {
		s.l.Info("连接不存在，添加连接", slog.String("sn", sn))
		// 添加连接到 SN 的连接列表
		s.AddConnection(sn, conn)
	}
	return nil
}

func (s *DroneImpl) HandleControlSession(ctx context.Context, conn *websocket.Conn, sn string, mt int, msgStr string) error {
	// 反序列化消息
	var msg dto.WSbaseModel
	if err := json.Unmarshal([]byte(msgStr), &msg); err != nil {
		s.l.Error("反序列化无人机控制消息失败", slog.String("sn", sn), slog.String("message", msgStr), slog.Any("error", err))
		return fmt.Errorf("反序列化无人机控制消息失败: %w", err)
	}
	s.l.Info("处理无人机控制消息", slog.String("sn", sn), slog.Any("message", msg))

	// 根据 Method 调用不同的处理逻辑
	switch msg.Method {
	case "init":
		dataBytes, err := json.Marshal(msg.Data)
		if err != nil {
			s.l.Error("序列化初始化控制数据失败", slog.String("sn", sn), slog.Any("data", msg.Data), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("序列化初始化控制数据失败: %w", err)
		}

		var InitData struct {
			Pitch         float64 `json:"pitch"`          // 云台俯仰角度
			Roll          float64 `json:"roll"`           // 云台横滚角度
			Yaw           float64 `json:"yaw"`            // 云台偏航角度
			CurrentCamera string  `json:"current_camera"` // 当前相机类型 (wide/zoom)
			Factor        float64 `json:"factor"`         // 缩放因子
		}
		if err := json.Unmarshal(dataBytes, &InitData); err != nil {
			s.l.Error("反序列化初始化控制数据失败", slog.String("sn", sn), slog.String("data", string(dataBytes)), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("反序列化初始化控制数据失败: %w", err)
		}
		s.l.Info("处理初始化控制数据", slog.String("sn", sn), slog.Any("data", InitData))
		msg.Timestamp = time.Now().Unix() // 更新时间戳
		resJson, err := json.Marshal(msg)
		if err != nil {
			s.l.Error("序列化初始化控制响应失败", slog.String("sn", sn), slog.Any("message", msg), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("序列化初始化控制响应失败: %w", err)
		}
		s.BroadcastToSN(sn, websocket.TextMessage, string(resJson))
		return nil
	case "auto_mode":
		// 处理相机切换控制
		dataBytes, err := json.Marshal(msg.Data)
		if err != nil {
			s.l.Error("序列化自动模式控制数据失败", slog.String("sn", sn), slog.Any("data", msg.Data), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("序列化自动模式控制数据失败: %w", err)
		}

		var switchCameraData struct {
			Action   string `json:"action"`   // 切换动作 (start/stop)
			Duration int    `json:"duration"` // 持续时间（秒）
		}
		if err := json.Unmarshal(dataBytes, &switchCameraData); err != nil {
			s.l.Error("反序列化自动模式控制数据失败", slog.String("sn", sn), slog.String("data", string(dataBytes)), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("反序列化自动模式控制数据失败: %w", err)
		}
		s.l.Info("处理自动模式控制", slog.String("sn", sn), slog.String("action", switchCameraData.Action))
		if switchCameraData.Action != "start" && switchCameraData.Action != "stop" {
			s.l.Error("无效的自动模式控制动作", slog.String("sn", sn), slog.String("action", switchCameraData.Action))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("无效的自动模式控制动作: %s", switchCameraData.Action)
		}
		msg.Timestamp = time.Now().Unix() // 更新时间戳
		resJson, err := json.Marshal(msg)
		if err != nil {
			s.l.Error("序列化自动模式控制响应失败", slog.String("sn", sn), slog.Any("message", msg), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("序列化自动模式控制响应失败: %w", err)
		}
		s.BroadcastToSN(sn, websocket.TextMessage, string(resJson))
		return nil
	case "switch_camera":
		// 处理相机切换控制
		dataBytes, err := json.Marshal(msg.Data)
		if err != nil {
			s.l.Error("序列化相机切换控制数据失败", slog.String("sn", sn), slog.Any("data", msg.Data), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("序列化相机切换控制数据失败: %w", err)
		}

		var switchCameraData struct {
			Camera string `json:"camera"` // 相机类型 (wide/zoom)
		}
		if err := json.Unmarshal(dataBytes, &switchCameraData); err != nil {
			s.l.Error("反序列化相机切换控制数据失败", slog.String("sn", sn), slog.String("data", string(dataBytes)), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("反序列化相机切换控制数据失败: %w", err)
		}
		s.l.Info("处理相机切换控制", slog.String("sn", sn), slog.String("camera", switchCameraData.Camera))
		msg.Timestamp = time.Now().Unix() // 更新时间戳
		resJson, err := json.Marshal(msg)
		if err != nil {
			s.l.Error("序列化相机切换控制响应失败", slog.String("sn", sn), slog.Any("message", msg), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("序列化相机切换控制响应失败: %w", err)
		}
		s.BroadcastToSN(sn, websocket.TextMessage, string(resJson))
		return nil
	case "set_gimbal_angle":
		// 处理云台角度控制
		dataBytes, err := json.Marshal(msg.Data)
		if err != nil {
			s.l.Error("序列化缩放控制数据失败", slog.String("sn", sn), slog.Any("data", msg.Data), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("序列化缩放控制数据失败: %w", err)
		}

		var GimbalAngleData struct {
			Pitch float64 `json:"pitch"` // 云台俯仰角度
			Roll  float64 `json:"roll"`  // 云台横滚角度
			Yaw   float64 `json:"yaw"`   // 云台偏航角度
		}
		if err := json.Unmarshal(dataBytes, &GimbalAngleData); err != nil {
			s.l.Error("反序列化云台角度控制数据失败", slog.String("sn", sn), slog.String("data", string(dataBytes)), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("反序列化云台角度控制数据失败: %w", err)
		}
		s.l.Info("处理云台角度控制", slog.String("sn", sn), slog.Float64("pitch", GimbalAngleData.Pitch), slog.Float64("roll", GimbalAngleData.Roll), slog.Float64("yaw", GimbalAngleData.Yaw))
		msg.Timestamp = time.Now().Unix() // 更新时间戳
		resJson, err := json.Marshal(msg)
		if err != nil {
			s.l.Error("序列化云台角度控制响应失败", slog.String("sn", sn), slog.Any("message", msg), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("序列化云台角度控制响应失败: %w", err)
		}
		s.BroadcastToSN(sn, websocket.TextMessage, string(resJson))
		return nil
	case "zoom":
		// 处理缩放控制
		// 首先将 msg.Data 转换为 JSON 字节数组进行二次反序列化
		dataBytes, err := json.Marshal(msg.Data)
		if err != nil {
			s.l.Error("序列化缩放控制数据失败", slog.String("sn", sn), slog.Any("data", msg.Data), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("序列化缩放控制数据失败: %w", err)
		}

		var zoomData struct {
			Factor float64 `json:"factor"` // 缩放因子
		}
		if err := json.Unmarshal(dataBytes, &zoomData); err != nil {
			s.l.Error("反序列化缩放控制数据失败", slog.String("sn", sn), slog.String("data", string(dataBytes)), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("反序列化缩放控制数据失败: %w", err)
		}

		s.l.Info("处理缩放控制", slog.String("sn", sn), slog.Float64("factor", zoomData.Factor))

		msg.Timestamp = time.Now().Unix() // 更新时间戳
		resJson, err := json.Marshal(msg)
		if err != nil {
			s.l.Error("序列化缩放控制响应失败", slog.String("sn", sn), slog.Any("message", msg), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("序列化缩放控制响应失败: %w", err)
		}
		s.BroadcastToSN(sn, websocket.TextMessage, string(resJson))
		return nil
	case "go_home":
		// 处理返航控制
		s.l.Info("处理返航控制", slog.String("sn", sn))
		msg.Timestamp = time.Now().Unix() // 更新时间戳
		resJson, err := json.Marshal(msg)
		if err != nil {
			s.l.Error("序列化返航控制响应失败", slog.String("sn", sn), slog.Any("message", msg), slog.Any("error", err))
			s.ReplyToConn(sn, conn, "error")
			return fmt.Errorf("序列化返航控制响应失败: %w", err)
		}
		s.BroadcastToSN(sn, websocket.TextMessage, string(resJson))
		return nil
	default:
		s.l.Warn("未知的无人机控制方法", slog.String("sn", sn), slog.String("method", msg.Method))
		msg.Timestamp = time.Now().Unix() // 更新时间戳
		s.ReplyToConn(sn, conn, "error")  // 这里可以根据实际情况选择是否回应连接
		return fmt.Errorf("未知的无人机控制方法: %s", msg.Method)
	}
}
