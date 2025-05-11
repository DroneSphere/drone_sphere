package txlive

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestGetRTMPUrl(t *testing.T) {
	// 设置为当前时间加上1天
	curTime := time.Now().Add(24 * time.Hour) // 1天后的时间
	fmt.Println("time: ", curTime)
	fmt.Println("seconds: ", curTime.Unix())
	// 测试用例结构体
	tests := []struct {
		name       string // 测试名称
		pushDomain string // 域名参数
		pullDomain string
		streamName string // 流名称参数
		pushKey    string // 推流密钥参数
		pullKey    string // 拉流密钥参数
		time       string // 时间戳参数
	}{
		{
			name:       "测试",
			pushDomain: "140433.livepush.myqcloud.com",
			pullDomain: "lisoft.com.cn",
			streamName: "test",
			pushKey:    "61c3e60eb8cf33bd4b4fb6a504fb51df",
			pullKey:    "drone",
			time:       strconv.FormatInt(curTime.Unix(), 16),
		},
	}

	// 执行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(tt)
			pushUrl := BuildRTMPUrl(tt.pushDomain, tt.streamName, tt.pushKey, tt.time)
			pullUrl := BuildRTMPUrl(tt.pullDomain, tt.streamName, tt.pullKey, tt.time)
			fmt.Println("pushUrl: ", pushUrl)
			fmt.Println("pullUrl: ", pullUrl)
		})
	}
}
