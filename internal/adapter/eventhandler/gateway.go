package eventhandler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/asaskevich/EventBus"
	"github.com/bytedance/sonic"
	"github.com/dronesphere/internal/event"
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/repo"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// GatewayHandler 网关设备事件处理器
type GatewayHandler struct {
	repo repo.GatewayRepo
	mqtt mqtt.Client
	l    *slog.Logger
	eb   EventBus.Bus
}

// NewGatewayHandler 创建网关事件处理器
func NewGatewayHandler(eb EventBus.Bus, mqtt mqtt.Client, repo repo.GatewayRepo, l *slog.Logger) *GatewayHandler {
	handler := &GatewayHandler{
		repo: repo,
		mqtt: mqtt,
		l:    l,
		eb:   eb,
	}

	// 订阅网关相关的 MQTT 主题
	// handler.subscribeMQTTTopics()

	return handler
}

// subscribeMQTTTopics 订阅网关相关的 MQTT 主题
func (h *GatewayHandler) subscribeMQTTTopics() {
	h.l.Info("初始化网关 MQTT 主题订阅")

	// 网关在线状态由 MQTT 客户端的连接状态自动维护，我们只需要监听拓扑更新消息
	template := "sys/product/+/status" // + 是通配符，表示匹配任意网关 SN
	token := h.mqtt.Subscribe(template, 1, h.handleUpdateTopo)
	if token.Wait() && token.Error() != nil {
		h.l.Error("网关状态主题订阅失败", "error", token.Error())
		return
	}
	h.l.Info("网关状态主题订阅成功", "topic", template)
}

// handleUpdateTopo 处理更新拓扑消息
func (h *GatewayHandler) handleUpdateTopo(c mqtt.Client, msg mqtt.Message) {
	// 从主题中提取网关 SN，格式：sys/product/{gateway_sn}/status
	parts := strings.Split(msg.Topic(), "/")
	if len(parts) != 4 {
		h.l.Error("无效的主题格式", "topic", msg.Topic())
		return
	}
	gatewaySN := parts[2]

	// 解析消息负载
	var p struct {
		dto.MessageCommon
		Data dto.UpdateTopoPayload `json:"data"`
	}
	if err := sonic.Unmarshal(msg.Payload(), &p); err != nil {
		h.l.Error("解析拓扑更新消息失败", "error", err)
		return
	}

	h.l.Info("收到网关拓扑更新消息",
		"gateway_sn", gatewaySN,
		"method", p.Method,
		"sub_devices_count", len(p.Data.SubDevices))

	// 提取已连接的无人机列表
	var droneSNs []string
	for _, device := range p.Data.SubDevices {
		droneSNs = append(droneSNs, device.SN)
	}

	// 发布拓扑更新事件
	payload := &event.GatewayUpdateTopoPayload{
		GatewaySN:       gatewaySN,
		ConnectedDrones: droneSNs,
	}
	h.eb.Publish(event.GatewayUpdateTopoEvent, payload)

	// 返回响应消息
	response := dto.NewMessageResult(dto.MessageCommon{}, 0)
	response.TID = p.TID
	response.BID = p.BID
	response.Method = p.Method
	response.Timestamp = p.Timestamp

	// 发布响应消息
	replyTopic := fmt.Sprintf("sys/product/%s/status_reply", gatewaySN)
	responseBytes, _ := sonic.Marshal(response)
	h.mqtt.Publish(replyTopic, 1, false, responseBytes)
}

// Subscribe 订阅事件总线上的事件
func (h *GatewayHandler) Subscribe(eb EventBus.Bus) {
	// 订阅网关相关事件
	eb.Subscribe(event.GatewayUpdateTopoEvent, h.handleGatewayUpdateTopo)
}

// handleGatewayUpdateTopo 处理网关拓扑更新事件
func (h *GatewayHandler) handleGatewayUpdateTopo(payload *event.GatewayUpdateTopoPayload) {
	// 获取网关序列号
	gatewaySN := payload.GatewaySN
	if gatewaySN == "" {
		h.l.Error("网关拓扑更新消息格式错误：缺少网关SN字段")
		return
	}

	// 获取连接的无人机列表
	droneSNs := payload.ConnectedDrones

	ctx := context.Background()

	// 获取当前已连接的无人机
	currentDrones, err := h.repo.GetConnectedDrones(ctx, gatewaySN)
	if err != nil {
		h.l.Error("获取当前连接的无人机失败", "error", err)
		return
	}

	// 构建当前连接的无人机 SN 集合
	currentDroneSNs := make(map[string]struct{})
	for _, drone := range currentDrones {
		currentDroneSNs[drone.SN] = struct{}{}
	}

	// 处理新增的连接关系
	for _, sn := range droneSNs {
		if _, exists := currentDroneSNs[sn]; !exists {
			err = h.repo.AddDroneRelation(ctx, gatewaySN, sn)
			if err != nil {
				h.l.Error("添加网关与无人机关联关系失败", "error", err)
				continue
			}
			h.l.Info("网关与无人机关联关系建立成功", "gateway_sn", gatewaySN, "drone_sn", sn)
		}
		delete(currentDroneSNs, sn)
	}

	// 处理需要删除的连接关系
	for sn := range currentDroneSNs {
		err = h.repo.RemoveDroneRelation(ctx, gatewaySN, sn)
		if err != nil {
			h.l.Error("移除网关与无人机关联关系失败", "error", err)
			continue
		}
		h.l.Info("网关与无人机关联关系移除成功", "gateway_sn", gatewaySN, "drone_sn", sn)
	}
}
