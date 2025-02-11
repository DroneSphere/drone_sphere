package dto

// 定义枚举常量
const (
	ModeCodeStandby            = 0
	ModeCodeTakeoffPreparation = 1
	ModeCodeTakeoffReady       = 2
	ModeCodeManualFlight       = 3
	ModeCodeAutoTakeoff        = 4
	ModeCodeRouteFlight        = 5
	ModeCodePanoramaPhoto      = 6
	ModeCodeIntelligentFollow  = 7
	ModeCodeADSBEvasion        = 8
	ModeCodeAutoReturn         = 9
	ModeCodeAutoLanding        = 10
	ModeCodeForcedLanding      = 11
	ModeCodeThreeBladeLanding  = 12
	ModeCodeUpgrading          = 13
	ModeCodeDisconnected       = 14
	ModeCodeAPAS               = 15
	ModeCodeVirtualJoystick    = 16
	ModeCodeCommandFlight      = 17
	ModeCodeAirRTKConvergence  = 18
)

var ModeCodeMap = map[int]string{
	ModeCodeStandby:            "待机",
	ModeCodeTakeoffPreparation: "起飞准备",
	ModeCodeTakeoffReady:       "起飞准备完毕",
	ModeCodeManualFlight:       "手动飞行",
	ModeCodeAutoTakeoff:        "自动起飞",
	ModeCodeRouteFlight:        "航线飞行",
	ModeCodePanoramaPhoto:      "全景拍照",
	ModeCodeIntelligentFollow:  "智能跟随",
	ModeCodeADSBEvasion:        "ADS-B 躲避",
	ModeCodeAutoReturn:         "自动返航",
	ModeCodeAutoLanding:        "自动降落",
	ModeCodeForcedLanding:      "强制降落",
	ModeCodeThreeBladeLanding:  "三桨叶降落",
	ModeCodeUpgrading:          "升级中",
	ModeCodeDisconnected:       "未连接",
	ModeCodeAPAS:               "APAS",
	ModeCodeVirtualJoystick:    "虚拟摇杆状态",
	ModeCodeCommandFlight:      "指令飞行",
	ModeCodeAirRTKConvergence:  "空中 RTK 收敛模式",
}

// 定义其他枚举常量
const (
	CameraModePhoto          = 0
	CameraModeVideo          = 1
	CameraModeIntelligentLow = 2
	CameraModePanoramaPhoto  = 3
)

var CameraModeMap = map[int]string{
	CameraModePhoto:          "拍照",
	CameraModeVideo:          "录像",
	CameraModeIntelligentLow: "智能低光",
	CameraModePanoramaPhoto:  "全景拍照",
}

type DroneHeartBeatPayload struct {
	Country                           string                  `json:"country"`
	ModeCode                          int                     `json:"mode_code"`
	ModeCodeReason                    int                     `json:"mode_code_reason"`
	Cameras                           []CameraInfo            `json:"cameras"`
	DongleInfos                       []DongleInfo            `json:"dongle_infos"`
	ObstacleAvoidance                 ObstacleAvoidance       `json:"obstacle_avoidance"`
	IsNearAreaLimit                   int                     `json:"is_near_area_limit"`
	IsNearHeightLimit                 int                     `json:"is_near_height_limit"`
	HeightLimit                       int                     `json:"height_limit"`
	NightLightsState                  int                     `json:"night_lights_state"`
	ActivationTime                    int                     `json:"activation_time"`
	MaintainStatus                    MaintainStatus          `json:"maintain_status"`
	TotalFlightSorties                int                     `json:"total_flight_sorties"`
	TrackID                           string                  `json:"track_id"`
	PositionState                     PositionState           `json:"position_state"`
	Storage                           Storage                 `json:"storage"`
	Battery                           Battery                 `json:"battery"`
	TotalFlightDistance               float64                 `json:"total_flight_distance"`
	TotalFlightTime                   int                     `json:"total_flight_time"`
	SeriousLowBatteryWarningThreshold int                     `json:"serious_low_battery_warning_threshold"`
	LowBatteryWarningThreshold        int                     `json:"low_battery_warning_threshold"`
	ControlSource                     string                  `json:"control_source"`
	WindDirection                     int                     `json:"wind_direction"`
	WindSpeed                         float64                 `json:"wind_speed"`
	HomeDistance                      float64                 `json:"home_distance"`
	HomeLatitude                      float64                 `json:"home_latitude"`
	HomeLongitude                     float64                 `json:"home_longitude"`
	AttitudeHead                      float64                 `json:"attitude_head"`
	AttitudeRoll                      float64                 `json:"attitude_roll"`
	AttitudePitch                     float64                 `json:"attitude_pitch"`
	Elevation                         float64                 `json:"elevation"`
	Height                            float64                 `json:"height"`
	Latitude                          float64                 `json:"latitude"`
	Longitude                         float64                 `json:"longitude"`
	VerticalSpeed                     float64                 `json:"vertical_speed"`
	HorizontalSpeed                   float64                 `json:"horizontal_speed"`
	FirmwareUpgradeStatus             int                     `json:"firmware_upgrade_status"`
	CompatibleStatus                  int                     `json:"compatible_status"`
	FirmwareVersion                   string                  `json:"firmware_version"`
	Gear                              int                     `json:"gear"`
	CameraWatermarkSettings           CameraWatermarkSettings `json:"camera_watermark_settings"`
}

type CameraInfo struct {
	RemainPhotoNum                  int             `json:"remain_photo_num"`
	RemainRecordDuration            int             `json:"remain_record_duration"`
	RecordTime                      int             `json:"record_time"`
	PayloadIndex                    string          `json:"payload_index"`
	CameraMode                      int             `json:"camera_mode"`
	PhotoState                      int             `json:"photo_state"`
	ScreenSplitEnable               bool            `json:"screen_split_enable"`
	RecordingState                  int             `json:"recording_state"`
	ZoomFactor                      float64         `json:"zoom_factor"`
	IRZoomFactor                    float64         `json:"ir_zoom_factor"`
	LiveviewWorldRegion             Region          `json:"liveview_world_region"`
	PhotoStorageSettings            []string        `json:"photo_storage_settings"`
	VideoStorageSettings            []string        `json:"video_storage_settings"`
	WideExposureMode                int             `json:"wide_exposure_mode"`
	WideISO                         int             `json:"wide_iso"`
	WideShutterSpeed                int             `json:"wide_shutter_speed"`
	WideExposureValue               int             `json:"wide_exposure_value"`
	ZoomExposureMode                int             `json:"zoom_exposure_mode"`
	ZoomISO                         int             `json:"zoom_iso"`
	ZoomShutterSpeed                int             `json:"zoom_shutter_speed"`
	ZoomExposureValue               int             `json:"zoom_exposure_value"`
	ZoomFocusMode                   int             `json:"zoom_focus_mode"`
	ZoomFocusValue                  int             `json:"zoom_focus_value"`
	ZoomMaxFocusValue               int             `json:"zoom_max_focus_value"`
	ZoomMinFocusValue               int             `json:"zoom_min_focus_value"`
	ZoomCalibrateFarthestFocusValue int             `json:"zoom_calibrate_farthest_focus_value"`
	ZoomCalibrateNearestFocusValue  int             `json:"zoom_calibrate_nearest_focus_value"`
	ZoomFocusState                  int             `json:"zoom_focus_state"`
	IRMeteringMode                  int             `json:"ir_metering_mode"`
	IRMeteringPoint                 IRMeteringPoint `json:"ir_metering_point"`
	IRMeteringArea                  IRMeteringArea  `json:"ir_metering_area"`
}

type Region struct {
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
}

type IRMeteringPoint struct {
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Temperature float64 `json:"temperature"`
}

type IRMeteringArea struct {
	X                   float64 `json:"x"`
	Y                   float64 `json:"y"`
	Width               float64 `json:"width"`
	Height              float64 `json:"height"`
	AverTemperature     float64 `json:"aver_temperature"`
	MinTemperaturePoint Point   `json:"min_temperature_point"`
	MaxTemperaturePoint Point   `json:"max_temperature_point"`
}

type Point struct {
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Temperature float64 `json:"temperature"`
}

type DongleInfo struct {
	IMEI              string     `json:"imei"`
	DongleType        int        `json:"dongle_type"`
	EID               string     `json:"eid"`
	ESIMActivateState int        `json:"esim_activate_state"`
	SIMCardState      int        `json:"sim_card_state"`
	SIMSlot           int        `json:"sim_slot"`
	ESIMInfos         []ESIMInfo `json:"esim_infos"`
	SIMInfo           SIMInfo    `json:"sim_info"`
}

type ESIMInfo struct {
	TelecomOperator int    `json:"telecom_operator"`
	Enabled         bool   `json:"enabled"`
	ICCID           string `json:"iccid"`
}

type SIMInfo struct {
	TelecomOperator int    `json:"telecom_operator"`
	SIMType         int    `json:"sim_type"`
	ICCID           string `json:"iccid"`
}

type ObstacleAvoidance struct {
	Horizon  int `json:"horizon"`
	Upside   int `json:"upside"`
	Downside int `json:"downside"`
}

type MaintainStatus struct {
	MaintainStatusArray []MaintainStatusItem `json:"maintain_status_array"`
}

type MaintainStatusItem struct {
	State                     int   `json:"state"`
	LastMaintainType          int   `json:"last_maintain_type"`
	LastMaintainTime          int64 `json:"last_maintain_time"`
	LastMaintainFlightTime    int   `json:"last_maintain_flight_time"`
	LastMaintainFlightSorties int   `json:"last_maintain_flight_sorties"`
}

type PositionState struct {
	IsFixed   int `json:"is_fixed"`
	Quality   int `json:"quality"`
	GPSNumber int `json:"gps_number"`
	RTKNumber int `json:"rtk_number"`
}

type Storage struct {
	Total int `json:"total"`
	Used  int `json:"used"`
}

type Battery struct {
	CapacityPercent  int             `json:"capacity_percent"`
	RemainFlightTime int             `json:"remain_flight_time"`
	ReturnHomePower  int             `json:"return_home_power"`
	LandingPower     int             `json:"landing_power"`
	Batteries        []BatteryDetail `json:"batteries"`
}

type BatteryDetail struct {
	CapacityPercent        int     `json:"capacity_percent"`
	Index                  int     `json:"index"`
	SN                     string  `json:"sn"`
	Type                   int     `json:"type"`
	SubType                int     `json:"sub_type"`
	FirmwareVersion        string  `json:"firmware_version"`
	LoopTimes              int     `json:"loop_times"`
	Voltage                int     `json:"voltage"`
	Temperature            float64 `json:"temperature"`
	HighVoltageStorageDays int     `json:"high_voltage_storage_days"`
}

type CameraWatermarkSettings struct {
	GlobalEnable           int    `json:"global_enable"`
	DroneTypeEnable        int    `json:"drone_type_enable"`
	DroneSNEnable          int    `json:"drone_sn_enable"`
	DateTimeEnable         int    `json:"datetime_enable"`
	GPSEnable              int    `json:"gps_enable"`
	UserCustomStringEnable int    `json:"user_custom_string_enable"`
	UserCustomString       string `json:"user_custom_string"`
	Layout                 int    `json:"layout"`
}
