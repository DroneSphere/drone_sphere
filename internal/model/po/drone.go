package po

import (
	"github.com/bytedance/sonic"
	"gorm.io/gorm"
)

type ORMDrone struct {
	gorm.Model
	SN       string `json:"sn"`
	Callsign string `json:"callsign"`
	Domain   string `json:"domain"`
	Type     int    `json:"type"`
	SubType  int    `json:"sub_type"`
}

func (ORMDrone) TableName() string {
	return "drones"
}

type RTDrone struct {
	SN           string `json:"sn" redis:"sn"`
	OnlineStatus bool   `json:"online_status" redis:"online_status"`
	RCSN         string `json:"rcsn" redis:"rcsn"`
	// 以下为 dto.DroneHeartBeat 信息
	Country                           string                  `json:"country" redis:"country"`
	ModeCode                          int                     `json:"mode_code" redis:"mode_code"`
	ModeCodeReason                    int                     `json:"mode_code_reason" redis:"mode_code_reason"`
	Cameras                           []CameraInfo            `json:"cameras" redis:"-"`
	DongleInfos                       []DongleInfo            `json:"dongle_infos" redis:"-"`
	ObstacleAvoidance                 ObstacleAvoidance       `json:"obstacle_avoidance" redis:"-"`
	IsNearAreaLimit                   int                     `json:"is_near_area_limit" redis:"is_near_area_limit"`
	IsNearHeightLimit                 int                     `json:"is_near_height_limit" redis:"is_near_height_limit"`
	HeightLimit                       int                     `json:"height_limit" redis:"height_limit"`
	NightLightsState                  int                     `json:"night_lights_state" redis:"night_lights_state"`
	ActivationTime                    int                     `json:"activation_time" redis:"activation_time"`
	MaintainStatus                    MaintainStatus          `json:"maintain_status" redis:"-"`
	TotalFlightSorties                int                     `json:"total_flight_sorties" redis:"total_flight_sorties"`
	TrackID                           string                  `json:"track_id" redis:"track_id"`
	PositionState                     PositionState           `json:"position_state" redis:"-"`
	Storage                           Storage                 `json:"storage" redis:"-"`
	Battery                           Battery                 `json:"battery" redis:"-"`
	TotalFlightDistance               float64                 `json:"total_flight_distance" redis:"total_flight_distance"`
	TotalFlightTime                   int                     `json:"total_flight_time" redis:"total_flight_time"`
	SeriousLowBatteryWarningThreshold int                     `json:"serious_low_battery_warning_threshold" redis:"serious_low_battery_warning_threshold"`
	LowBatteryWarningThreshold        int                     `json:"low_battery_warning_threshold" redis:"low_battery_warning_threshold"`
	ControlSource                     string                  `json:"control_source" redis:"control_source"`
	WindDirection                     int                     `json:"wind_direction" redis:"wind_direction"`
	WindSpeed                         float64                 `json:"wind_speed" redis:"wind_speed"`
	HomeDistance                      float64                 `json:"home_distance" redis:"home_distance"`
	HomeLatitude                      float64                 `json:"home_latitude" redis:"home_latitude"`
	HomeLongitude                     float64                 `json:"home_longitude" redis:"home_longitude"`
	AttitudeHead                      float64                 `json:"attitude_head" redis:"attitude_head"`
	AttitudeRoll                      float64                 `json:"attitude_roll" redis:"attitude_roll"`
	AttitudePitch                     float64                 `json:"attitude_pitch" redis:"attitude_pitch"`
	Elevation                         float64                 `json:"elevation" redis:"elevation"`
	Height                            float64                 `json:"height" redis:"height"`
	Latitude                          float64                 `json:"latitude" redis:"latitude"`
	Longitude                         float64                 `json:"longitude" redis:"longitude"`
	VerticalSpeed                     float64                 `json:"vertical_speed" redis:"vertical_speed"`
	HorizontalSpeed                   float64                 `json:"horizontal_speed" redis:"horizontal_speed"`
	FirmwareUpgradeStatus             int                     `json:"firmware_upgrade_status" redis:"firmware_upgrade_status"`
	CompatibleStatus                  int                     `json:"compatible_status" redis:"compatible_status"`
	FirmwareVersion                   string                  `json:"firmware_version" redis:"firmware_version"`
	Gear                              int                     `json:"gear" redis:"gear"`
	CameraWatermarkSettings           CameraWatermarkSettings `json:"camera_watermark_settings" redis:"-"`
	// 以上为 dto.DroneHeartBeat 信息
}

func (t RTDrone) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t RTDrone) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}

// GetHeading 获取无人机的航向角
func (t RTDrone) GetHeading() float64 {
	return t.AttitudeHead
}

type CameraInfo struct {
	RemainPhotoNum                  int             `json:"remain_photo_num" redis:"remain_photo_num"`
	RemainRecordDuration            int             `json:"remain_record_duration" redis:"remain_record_duration"`
	RecordTime                      int             `json:"record_time" redis:"record_time"`
	PayloadIndex                    string          `json:"payload_index" redis:"payload_index"`
	CameraMode                      int             `json:"camera_mode" redis:"camera_mode"`
	PhotoState                      int             `json:"photo_state" redis:"photo_state"`
	ScreenSplitEnable               bool            `json:"screen_split_enable" redis:"screen_split_enable"`
	RecordingState                  int             `json:"recording_state" redis:"recording_state"`
	ZoomFactor                      float64         `json:"zoom_factor" redis:"zoom_factor"`
	IRZoomFactor                    float64         `json:"ir_zoom_factor" redis:"ir_zoom_factor"`
	LiveviewWorldRegion             Region          `json:"liveview_world_region" redis:"liveview_world_region"`
	PhotoStorageSettings            []string        `json:"photo_storage_settings" redis:"photo_storage_settings"`
	VideoStorageSettings            []string        `json:"video_storage_settings" redis:"video_storage_settings"`
	WideExposureMode                int             `json:"wide_exposure_mode" redis:"wide_exposure_mode"`
	WideISO                         int             `json:"wide_iso" redis:"wide_iso"`
	WideShutterSpeed                int             `json:"wide_shutter_speed" redis:"wide_shutter_speed"`
	WideExposureValue               int             `json:"wide_exposure_value" redis:"wide_exposure_value"`
	ZoomExposureMode                int             `json:"zoom_exposure_mode" redis:"zoom_exposure_mode"`
	ZoomISO                         int             `json:"zoom_iso" redis:"zoom_iso"`
	ZoomShutterSpeed                int             `json:"zoom_shutter_speed" redis:"zoom_shutter_speed"`
	ZoomExposureValue               int             `json:"zoom_exposure_value" redis:"zoom_exposure_value"`
	ZoomFocusMode                   int             `json:"zoom_focus_mode" redis:"zoom_focus_mode"`
	ZoomFocusValue                  int             `json:"zoom_focus_value" redis:"zoom_focus_value"`
	ZoomMaxFocusValue               int             `json:"zoom_max_focus_value" redis:"zoom_max_focus_value"`
	ZoomMinFocusValue               int             `json:"zoom_min_focus_value" redis:"zoom_min_focus_value"`
	ZoomCalibrateFarthestFocusValue int             `json:"zoom_calibrate_farthest_focus_value" redis:"zoom_calibrate_farthest_focus_value"`
	ZoomCalibrateNearestFocusValue  int             `json:"zoom_calibrate_nearest_focus_value" redis:"zoom_calibrate_nearest_focus_value"`
	ZoomFocusState                  int             `json:"zoom_focus_state" redis:"zoom_focus_state"`
	IRMeteringMode                  int             `json:"ir_metering_mode" redis:"ir_metering_mode"`
	IRMeteringPoint                 IRMeteringPoint `json:"ir_metering_point" redis:"ir_metering_point"`
	IRMeteringArea                  IRMeteringArea  `json:"ir_metering_area" redis:"ir_metering_area"`
}

func (t CameraInfo) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t CameraInfo) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}

type Region struct {
	Left   float64 `json:"left" redis:"left"`
	Top    float64 `json:"top" redis:"top"`
	Right  float64 `json:"right" redis:"right"`
	Bottom float64 `json:"bottom" redis:"bottom"`
}

func (t Region) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t Region) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}

type IRMeteringPoint struct {
	X           float64 `json:"x" redis:"x"`
	Y           float64 `json:"y" redis:"y"`
	Temperature float64 `json:"temperature" redis:"temperature"`
}

func (t IRMeteringPoint) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t IRMeteringPoint) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}

type IRMeteringArea struct {
	X                   float64 `json:"x" redis:"x"`
	Y                   float64 `json:"y" redis:"y"`
	Width               float64 `json:"width" redis:"width"`
	Height              float64 `json:"height" redis:"height"`
	AverTemperature     float64 `json:"aver_temperature" redis:"aver_temperature"`
	MinTemperaturePoint Point   `json:"min_temperature_point" redis:"min_temperature_point"`
	MaxTemperaturePoint Point   `json:"max_temperature_point" redis:"max_temperature_point"`
}

func (t IRMeteringArea) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t IRMeteringArea) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}

type Point struct {
	X           float64 `json:"x" redis:"x"`
	Y           float64 `json:"y" redis:"y"`
	Temperature float64 `json:"temperature" redis:"temperature"`
}

func (t Point) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t Point) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}

type DongleInfo struct {
	IMEI              string     `json:"imei" redis:"imei"`
	DongleType        int        `json:"dongle_type" redis:"dongle_type"`
	EID               string     `json:"eid" redis:"eid"`
	ESIMActivateState int        `json:"esim_activate_state" redis:"esim_activate_state"`
	SIMCardState      int        `json:"sim_card_state" redis:"sim_card_state"`
	SIMSlot           int        `json:"sim_slot" redis:"sim_slot"`
	ESIMInfos         []ESIMInfo `json:"esim_infos" redis:"esim_infos"`
	SIMInfo           SIMInfo    `json:"sim_info" redis:"sim_info"`
}

func (t DongleInfo) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t DongleInfo) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}

type ESIMInfo struct {
	TelecomOperator int    `json:"telecom_operator" redis:"telecom_operator"`
	Enabled         bool   `json:"enabled" redis:"enabled"`
	ICCID           string `json:"iccid" redis:"iccid"`
}

func (t ESIMInfo) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t ESIMInfo) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}

type SIMInfo struct {
	TelecomOperator int    `json:"telecom_operator" redis:"telecom_operator"`
	SIMType         int    `json:"sim_type" redis:"sim_type"`
	ICCID           string `json:"iccid" redis:"iccid"`
}

func (t SIMInfo) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t SIMInfo) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}

type ObstacleAvoidance struct {
	Horizon  int `json:"horizon" redis:"horizon"`
	Upside   int `json:"upside" redis:"upside"`
	Downside int `json:"downside" redis:"downside"`
}

func (t ObstacleAvoidance) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t ObstacleAvoidance) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}

type MaintainStatus struct {
	MaintainStatusArray []MaintainStatusItem `json:"maintain_status_array"`
}

func (t MaintainStatus) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t MaintainStatus) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}

type MaintainStatusItem struct {
	State                     int   `json:"state"`
	LastMaintainType          int   `json:"last_maintain_type"`
	LastMaintainTime          int64 `json:"last_maintain_time"`
	LastMaintainFlightTime    int   `json:"last_maintain_flight_time"`
	LastMaintainFlightSorties int   `json:"last_maintain_flight_sorties"`
}

func (t MaintainStatusItem) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t MaintainStatusItem) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}

type PositionState struct {
	IsFixed   int `json:"is_fixed"`
	Quality   int `json:"quality"`
	GPSNumber int `json:"gps_number"`
	RTKNumber int `json:"rtk_number"`
}

func (t PositionState) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t PositionState) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}

type Storage struct {
	Total int `json:"total"`
	Used  int `json:"used"`
}

func (t Storage) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t Storage) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}

type Battery struct {
	CapacityPercent  int             `json:"capacity_percent"`
	RemainFlightTime int             `json:"remain_flight_time"`
	ReturnHomePower  int             `json:"return_home_power"`
	LandingPower     int             `json:"landing_power"`
	Batteries        []BatteryDetail `json:"batteries"`
}

func (t Battery) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t Battery) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
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

func (t BatteryDetail) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t BatteryDetail) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
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

func (t CameraWatermarkSettings) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(t)
}

func (t CameraWatermarkSettings) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, t)
}
