package mock_tool

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"sync"
	"time"
)

type MQTTClient struct {
	mu             sync.Mutex
	connected      bool
	connectionOpen bool
	PublishCh      chan *PublishPacket
	SubscribeCh    chan *SubscribePacket
	UnsubCh        chan *UnsubscribePacket
}

type PublishPacket struct {
	TopicName string
	Qos       byte
	Retain    bool
	Payload   []byte
}

type SubscribePacket struct {
	Topics []string
	Qos    []byte
}

type UnsubscribePacket struct {
	Topics []string
}

func NewMockMQTTClient() *MQTTClient {
	return &MQTTClient{
		connected:      true,
		connectionOpen: true,
		PublishCh:      make(chan *PublishPacket, 10),
		SubscribeCh:    make(chan *SubscribePacket, 10),
		UnsubCh:        make(chan *UnsubscribePacket, 10),
	}
}

func (m *MQTTClient) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected
}

func (m *MQTTClient) IsConnectionOpen() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connectionOpen
}

func (m *MQTTClient) Connect() mqtt.Token {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = true
	m.connectionOpen = true
	return &mqtt.DummyToken{}
}

func (m *MQTTClient) Disconnect(quiesce uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	time.Sleep(time.Duration(quiesce) * time.Millisecond)
	m.connected = false
	m.connectionOpen = false
}

func (m *MQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	m.PublishCh <- &PublishPacket{
		TopicName: topic,
		Qos:       qos,
		Retain:    retained,
		Payload:   payload.([]byte),
	}
	return &mqtt.DummyToken{}
}

func (m *MQTTClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	m.SubscribeCh <- &SubscribePacket{
		Topics: []string{topic},
		Qos:    []byte{qos},
	}
	return &mqtt.DummyToken{}
}

func (m *MQTTClient) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	topics := make([]string, 0, len(filters))
	qos := make([]byte, 0, len(filters))
	for topic, q := range filters {
		topics = append(topics, topic)
		qos = append(qos, q)
	}
	m.SubscribeCh <- &SubscribePacket{
		Topics: topics,
		Qos:    qos,
	}
	return &mqtt.DummyToken{}
}

func (m *MQTTClient) Unsubscribe(topics ...string) mqtt.Token {
	m.UnsubCh <- &UnsubscribePacket{
		Topics: topics,
	}
	return &mqtt.DummyToken{}
}

func (m *MQTTClient) AddRoute(topic string, callback mqtt.MessageHandler) {
	// No-op for mock
}

func (m *MQTTClient) OptionsReader() mqtt.ClientOptionsReader {
	return mqtt.ClientOptionsReader{}
}
