package mock_tool_test

import (
	"testing"
	"time"

	"github.com/dronesphere/tools/mock_tool"
	"github.com/stretchr/testify/assert"
)

func TestMQTTClient_IsConnected(t *testing.T) {
	client := mock_tool.NewMockMQTTClient()
	assert.True(t, client.IsConnected())
}

func TestMQTTClient_IsConnectionOpen(t *testing.T) {
	client := mock_tool.NewMockMQTTClient()
	assert.True(t, client.IsConnectionOpen())
}

func TestMQTTClient_Connect(t *testing.T) {
	client := mock_tool.NewMockMQTTClient()
	client.Disconnect(0)
	token := client.Connect()
	assert.True(t, token.WaitTimeout(time.Second))
	assert.True(t, client.IsConnected())
	assert.True(t, client.IsConnectionOpen())
}

func TestMQTTClient_Disconnect(t *testing.T) {
	client := mock_tool.NewMockMQTTClient()
	client.Disconnect(100)
	assert.False(t, client.IsConnected())
	assert.False(t, client.IsConnectionOpen())
}

func TestMQTTClient_Publish(t *testing.T) {
	client := mock_tool.NewMockMQTTClient()
	payload := []byte("test payload")
	token := client.Publish("test/topic", 1, false, payload)
	assert.True(t, token.WaitTimeout(time.Second))
	packet := <-client.PublishCh
	assert.Equal(t, "test/topic", packet.TopicName)
	assert.Equal(t, byte(1), packet.Qos)
	assert.False(t, packet.Retain)
	assert.Equal(t, payload, packet.Payload)
}

func TestMQTTClient_Subscribe(t *testing.T) {
	client := mock_tool.NewMockMQTTClient()
	token := client.Subscribe("test/topic", 1, nil)
	assert.True(t, token.WaitTimeout(time.Second))
	packet := <-client.SubscribeCh
	assert.Equal(t, []string{"test/topic"}, packet.Topics)
	assert.Equal(t, []byte{1}, packet.Qos)
}

func TestMQTTClient_SubscribeMultiple(t *testing.T) {
	client := mock_tool.NewMockMQTTClient()
	filters := map[string]byte{"test/topic1": 1, "test/topic2": 2}
	token := client.SubscribeMultiple(filters, nil)
	assert.True(t, token.WaitTimeout(time.Second))
	packet := <-client.SubscribeCh
	assert.ElementsMatch(t, []string{"test/topic1", "test/topic2"}, packet.Topics)
	assert.ElementsMatch(t, []byte{1, 2}, packet.Qos)
}

func TestMQTTClient_Unsubscribe(t *testing.T) {
	client := mock_tool.NewMockMQTTClient()
	token := client.Unsubscribe("test/topic1", "test/topic2")
	assert.True(t, token.WaitTimeout(time.Second))
	packet := <-client.UnsubCh
	assert.ElementsMatch(t, []string{"test/topic1", "test/topic2"}, packet.Topics)
}
