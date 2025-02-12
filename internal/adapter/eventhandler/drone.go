package eventhandler

import (
	"context"
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/bytedance/sonic"
	"github.com/dronesphere/internal/event"
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/service"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jinzhu/copier"
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
	err = eb.Subscribe(event.UserLoginSuccessEvent, handler.HandleTopoUpdate)
	if err != nil {
		l.Error(fmt.Sprintf("subscribe event %s failed: %v", event.UserLoginSuccessEvent, err))
	}
	err = eb.Subscribe(event.DroneOnlineEvent, handler.HandleDroneOSD)
	if err != nil {
		l.Error(fmt.Sprintf("subscribe event %s failed: %v", event.DroneOnlineEvent, err))
		panic(err)
	}
	err = eb.Subscribe(event.DroneOnlineEvent, handler.HandleDroneState)
	if err != nil {
		l.Error(fmt.Sprintf("subscribe event %s failed: %v", event.DroneOnlineEvent, err))
		panic(err)
	}
}

func (d *DroneEventHandler) HandleTopoUpdate(ctx context.Context) error {
	sn := ctx.Value("sn").(string)
	d.l.Info("OnTopoUpdate", slog.Any("sn", sn))

	template := "sys/product/%s/status"
	topic := fmt.Sprintf(template, sn)
	d.l.Info("Will subscribe topic", slog.Any("topic", topic))

	token := d.mqtt.Subscribe(topic, 1, func(c mqtt.Client, m mqtt.Message) {
		d.l.Info("Received message", slog.Any("topic", m.Topic()), slog.Any("message", string(m.Payload())))
		// 解析消息
		var p struct {
			dto.MessageCommon
			Data dto.UpdateTopoPayload `json:"data"`
		}
		if err := sonic.Unmarshal(m.Payload(), &p); err != nil {
			d.l.Error("Unmarshal message failed", slog.Any("err", err))
			return
		}
		d.l.Info("Unmarshal message", slog.Any("updatePayload", p))

		// 处理网络拓扑
		if err := d.svc.SaveDroneTopo(ctx, p.Data); err != nil {
			d.l.Error("SaveDroneTopo failed", slog.Any("err", err))
			return
		}

		// 当有子设备时，发布设备上线事件
		if len(p.Data.SubDevices) > 0 {
			droneSN := p.Data.SubDevices[0].SN
			d.l.Info("Publish drone online event", slog.Any("droneSN", droneSN))
			ctx := context.WithValue(context.Background(), event.DroneEventSNKey, droneSN)
			d.eb.Publish(event.DroneOnlineEvent, ctx)
		}

		// 发布成功消息响应
		res := struct {
			dto.MessageCommon
			Data struct {
				Result int `json:"result"`
			} `json:"data"`
		}{
			MessageCommon: p.MessageCommon,
			Data: struct {
				Result int `json:"result"`
			}{
				Result: 0,
			},
		}
		r, err := sonic.Marshal(res)
		if err != nil {
			d.l.Error("Marshal message failed", slog.Any("err", err))
			return
		}
		topic := fmt.Sprintf("sys/product/%s/status_reply", sn)
		d.mqtt.Publish(topic, 1, false, []byte(r))
	})
	if token.Wait() && token.Error() != nil {
		d.l.Error("Subscribe topic failed", slog.Any("err", token.Error()))
		return token.Error()
	}

	d.l.Info("Subscribe topic success", slog.Any("topic", topic))
	return nil
}

type HeartbeatPayload struct {
	dto.MessageCommon
	Data dto.DroneHeartBeat `json:"data"`
}

func (d *DroneEventHandler) parseHeartBeat(m mqtt.Message) (po.RTDrone, bool) {
	var p HeartbeatPayload
	res := po.RTDrone{}
	if err := sonic.Unmarshal(m.Payload(), &p); err != nil {
		d.l.Error("Unmarshal message failed", slog.Any("err", err))
		return res, false
	}
	if err := copier.Copy(&res, &p.Data); err != nil {
		d.l.Error("Copy message failed", slog.Any("err", err))
		return res, false
	}
	//if res.Battery != interface{}(nil) {
	//	res.OnlineStatus = true
	//} else {
	//	res.OnlineStatus = false
	//}
	return res, true
}

func (d *DroneEventHandler) HandleDroneOSD(ctx context.Context) error {
	droneSN := ctx.Value(event.DroneEventSNKey).(string)
	d.l.Info("Handle drone OSD event", slog.Any("droneSN", droneSN))
	template := "thing/product/%s/osd"
	topic := fmt.Sprintf(template, droneSN)
	d.l.Info("Subscribe drone OSD topic", slog.Any("topic", topic))

	token := d.mqtt.Subscribe(topic, 1, func(c mqtt.Client, m mqtt.Message) {
		p, ok := d.parseHeartBeat(m)
		if !ok {
			d.l.Error("Parse heartbeat failed")
			return
		}
		d.l.Info("Product OSD received message", slog.Any("topic", topic), slog.Any("payload", p))

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
