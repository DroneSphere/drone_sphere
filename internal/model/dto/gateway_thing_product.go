package dto

type GatewayOSDData struct {
	LiveCapacity *struct {
		AvailableVideoNumber  int `json:"available_video_number,omitempty"`
		CoexistVideoNumberMax int `json:"coexist_video_number_max,omitempty"`
		DeviceList            []struct {
			SN                    string `json:"sn,omitempty"`
			AvailableVideoNumber  int    `json:"available_video_number,omitempty"`
			CoexistVideoNumberMax int    `json:"coexist_video_number_max,omitempty"`
			CameraList            []struct {
				CameraIndex           string `json:"camera_index,omitempty"`
				AvailableVideoNumber  int    `json:"available_video_number,omitempty"`
				CoexistVideoNumberMax int    `json:"coexist_video_number_max,omitempty"`
				VideoList             []struct {
					VideoIndex           string   `json:"video_index,omitempty"`
					VideoType            string   `json:"video_type,omitempty"`
					SwitchableVideoTypes []string `json:"switchable_video_types,omitempty"`
				} `json:"video_list,omitempty"`
			} `json:"camera_list,omitempty"`
		} `json:"device_list,omitempty"`
	} `json:"live_capacity,omitempty"`
	Country         string  `json:"country,omitempty"`
	CapacityPercent int     `json:"capacity_percent,omitempty"`
	Height          float64 `json:"height,omitempty"`
	DongleInfos     []struct {
		IMEI              string `json:"imei,omitempty"`
		DongleType        int    `json:"dongle_type,omitempty"`
		EID               string `json:"eid,omitempty"`
		EsimActivateState int    `json:"esim_activate_state,omitempty"`
		SimCardState      int    `json:"sim_card_state,omitempty"`
		SimSlot           int    `json:"sim_slot,omitempty"`
		EsimInfos         []struct {
			TelecomOperator int    `json:"telecom_operator,omitempty"`
			Enabled         bool   `json:"enabled,omitempty"`
			Iccid           string `json:"iccid,omitempty"`
		} `json:"esim_infos,omitempty"`
		SimInfo *struct {
			TelecomOperator int    `json:"telecom_operator,omitempty"`
			SimType         int    `json:"sim_type,omitempty"`
			Iccid           string `json:"iccid,omitempty"`
		} `json:"sim_info,omitempty"`
	} `json:"dongle_infos,omitempty"`
	LiveStatus []struct {
		VideoID      string `json:"video_id,omitempty"`
		VideoType    string `json:"video_type,omitempty"`
		VideoQuality int    `json:"video_quality,omitempty"`
		Status       int    `json:"status,omitempty"`
		ErrorStatus  int    `json:"error_status,omitempty"`
	} `json:"live_status,omitempty"`
	WirelessLink *struct {
		DongleNumber    int     `json:"dongle_number,omitempty"`
		FourGLinkState  int     `json:"4g_link_state,omitempty"`
		SDRLinkState    int     `json:"sdr_link_state,omitempty"`
		LinkWorkmode    int     `json:"link_workmode,omitempty"`
		SDRQuality      int     `json:"sdr_quality,omitempty"`
		FourGQuality    int     `json:"4g_quality,omitempty"`
		FourGUAVQuality int     `json:"4g_uav_quality,omitempty"`
		FourGGNDQuality int     `json:"4g_gnd_quality,omitempty"`
		SDRFreqBand     float64 `json:"sdr_freq_band,omitempty"`
		FourGFreqBand   float64 `json:"4g_freq_band,omitempty"`
	} `json:"wireless_link,omitempty"`
	FirmwareVersion string  `json:"firmware_version,omitempty"`
	Latitude        float64 `json:"latitude,omitempty"`
	Longitude       float64 `json:"longitude,omitempty"`
}

// LiveStartPushRequest 对应MQTT live_start_push方法的请求体
// Topic: thing/product/*{gateway_sn}*/services
// Method: live_start_push
type LiveStartPushData struct {
	URL          string `json:"url"`           // RTMP推流地址
	URLType      int    `json:"url_type"`      // 直播协议类型, 1 for RTMP
	VideoID      string `json:"video_id"`      // 直播视频流的 ID, 格式: {sn}/{camera_index}/{video_index}
	VideoQuality int    `json:"video_quality"` // 直播质量, 0:自适应, 1:流畅, 2:标清, 3:高清, 4:超清
}

type LiveStartPushRequest struct {
	BID       string            `json:"bid"`       // 请求唯一ID
	Data      LiveStartPushData `json:"data"`      // 具体业务数据
	TID       string            `json:"tid"`       // 终端唯一ID
	Timestamp int64             `json:"timestamp"` // 请求时间戳 (ms)
	Method    string            `json:"method"`    // 固定为 "live_start_push"
}

// LiveStopPushRequest 对应MQTT live_stop_push方法的请求体
// Topic: thing/product/*{gateway_sn}*/services
// Method: live_stop_push
type LiveStopPushData struct {
	VideoID string `json:"video_id"` // 直播视频流的 ID
}

type LiveStopPushRequest struct {
	BID       string           `json:"bid"`       // 请求唯一ID
	Data      LiveStopPushData `json:"data"`      // 具体业务数据
	TID       string           `json:"tid"`       // 终端唯一ID
	Timestamp int64            `json:"timestamp"` // 请求时间戳 (ms)
	Method    string           `json:"method"`    // 固定为 "live_stop_push"
}
