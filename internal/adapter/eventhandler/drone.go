package eventhandler

import (
	"context"
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/bytedance/sonic"
	"github.com/dronesphere/internal/event"
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/service"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log/slog"
)

type DroneEventHandler struct {
	eb   EventBus.Bus
	l    *slog.Logger
	svc  service.DroneSvc
	mqtt mqtt.Client
}

func registerDroneHandlers(eb EventBus.Bus, l *slog.Logger, mqtt mqtt.Client, drone service.DroneSvc) {
	handler := &DroneEventHandler{
		eb:   eb,
		l:    l,
		svc:  drone,
		mqtt: mqtt,
	}
	var err error
	err = eb.Subscribe(event.RemoteControllerLoggedIn, handler.HandleTopoUpdate)
	if err != nil {
		panic(err)
	}
	_ = eb.Subscribe(event.DroneConnected, handler.HandleDroneConnected)
	err = eb.Subscribe(event.DroneConnected, handler.HandleDroneOSD)
	if err != nil {
		panic(err)
	}
	//err = eb.Subscribe(event.DroneConnected, handler.HandleDroneState)
	//if err != nil {
	//	panic(err)
	//}
}

// HandleTopoUpdate 处理拓扑更新事件
//
// 监听 sys/product/{sn}/status 主题，处理无人机拓扑更新事件
func (d *DroneEventHandler) HandleTopoUpdate(ctx context.Context) error {
	gatewaySN := ctx.Value(event.RemoteControllerLoginSNKey).(string)
	template := "sys/product/%s/status"
	subscribeTopic := fmt.Sprintf(template, gatewaySN)
	d.l.Info("识别网关设备上线", slog.Any("topic", subscribeTopic), slog.Any("desc", "设备上下线、更新拓扑"), slog.Any("gatewaySN", gatewaySN))

	token := d.mqtt.Subscribe(subscribeTopic, 1, func(c mqtt.Client, m mqtt.Message) {
		var p struct {
			dto.MessageCommon
			Data dto.UpdateTopoPayload `json:"data"`
		}
		_ = sonic.Unmarshal(m.Payload(), &p)
		d.l.Info("接收网关设备上下线消息", slog.Any("topic", m.Topic()), slog.Any("payload", p))

		// SubDevices 够长说明为无人机上线事件，否则为下线事件
		if len(p.Data.SubDevices) > 0 {
			droneSN := p.Data.SubDevices[0].SN
			d.l.Info("识别无人机上线", slog.Any("droneSN", droneSN))
			ctx = context.WithValue(ctx, event.DroneEventSNKey, droneSN)
			ctx = context.WithValue(ctx, event.DroneEventTopoKey, p.Data.SubDevices[0].ProductTopo)
			d.eb.Publish(event.DroneConnected, ctx)
		} else {
			d.l.Info("识别无人机下线", slog.Any("gatewaySN", gatewaySN))
		}

		// 发布成功消息响应
		r, _ := sonic.Marshal(dto.NewMessageResult(0))
		publishTopic := fmt.Sprintf("sys/product/%s/status_reply", gatewaySN)
		d.mqtt.Publish(publishTopic, 1, false, []byte(r))
	})
	if token.Wait() && token.Error() != nil {
		d.l.Error("设备上下线订阅失败", slog.Any("topic", subscribeTopic), slog.Any("err", token.Error()))
		return token.Error()
	} else {
		d.l.Info("设备上下线订阅成功", slog.Any("topic", subscribeTopic))
	}

	return nil
}

// HandleDroneConnected 处理无人机连接事件
//
// 处理无人机通过网关设备连接后，需要进行的持久化、数据更新等操作
func (d *DroneEventHandler) HandleDroneConnected(ctx context.Context) error {
	droneSN := ctx.Value(event.DroneEventSNKey).(string)
	topo := ctx.Value(event.DroneEventTopoKey).(dto.ProductTopo)
	d.l.Info("处理无人机连接事件", slog.Any("droneSN", droneSN), slog.Any("topo", topo))

	e := entity.NewDroneFromMsg(droneSN, topo)
	if err := d.svc.Repo().Save(ctx, *e); err != nil {
		d.l.Error("保存无人机信息失败", slog.Any("err", err))
		return err
	}
	d.l.Info("保存无人机信息成功", slog.Any("droneSN", droneSN))

	return nil
}

func (d *DroneEventHandler) parseHeartBeat(m mqtt.Message) (dto.DroneMessageProperty, bool) {
	var p struct {
		dto.MessageCommon
		Data dto.DroneMessageProperty `json:"data"`
	}
	_ = sonic.Unmarshal(m.Payload(), &p)

	return p.Data, true
}

// HandleDroneOSD 处理无人机 OSD 事件
//
// 监听 thing/product/{sn}/osd 主题，处理无人机 OSD 事件
func (d *DroneEventHandler) HandleDroneOSD(ctx context.Context) error {
	gatewaySN := ctx.Value(event.RemoteControllerLoginSNKey).(string)
	droneSN := ctx.Value(event.DroneEventSNKey).(string)
	template := "thing/product/%s/osd"
	topic := fmt.Sprintf(template, droneSN)
	d.l.Info("识别无人机上线", slog.Any("topic", topic), slog.Any("droneSN", droneSN), slog.Any("gatewaySN", gatewaySN))

	token := d.mqtt.Subscribe(topic, 1, func(c mqtt.Client, m mqtt.Message) {
		p, ok := d.parseHeartBeat(m)
		if !ok {
			return
		}
		d.l.Info("接收无人机 OSD 消息", slog.Any("topic", m.Topic()), slog.Any("payload", p))

		if err := d.svc.UpdateStateBySN(ctx, droneSN, p); err != nil {
			d.l.Error("更新无人机实时数据失败", slog.Any("err", err))
			return
		}
		d.l.Info("更新无人机实时数据成功", slog.Any("droneSN", droneSN))
	})
	if token.Wait() && token.Error() != nil {
		d.l.Error("无人机 OSD 订阅失败", slog.Any("topic", topic), slog.Any("err", token.Error()))
		return token.Error()
	} else {
		d.l.Info("无人机 OSD 订阅成功", slog.Any("topic", topic))
	}

	return nil
}

func (d *DroneEventHandler) HandleDroneState(ctx context.Context) error {
	droneSN := ctx.Value(event.DroneEventSNKey).(string)
	d.l.Info("Handle drone state event", slog.Any("droneSN", droneSN))
	template := "thing/product/%s/state"
	topic := fmt.Sprintf(template, droneSN)
	d.l.Info("Subscribe drone state topic", slog.Any("topic", topic))

	token := d.mqtt.Subscribe(topic, 1, func(c mqtt.Client, m mqtt.Message) {
		d.l.Info("Received message", slog.Any("topic", m.Topic()), slog.Any("message", string(m.Payload())))
		// 解析消息
		p, ok := d.parseHeartBeat(m)
		if !ok {
			d.l.Error("Parse heartbeat failed")
			return
		}

		if err := d.svc.UpdateStateBySN(ctx, droneSN, p); err != nil {
			d.l.Error("Update drone online failed", slog.Any("err", err))
			return
		}

		d.l.Info("Update drone online success", slog.Any("droneSN", droneSN))
	})
	if token.Wait() && token.Error() != nil {
		d.l.Error("Subscribe topic failed", slog.Any("err", token.Error()))
		return token.Error()
	}

	return nil
}
