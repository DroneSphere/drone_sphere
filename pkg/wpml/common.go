package wpml

import (
	"encoding/xml"
	"errors"
)

// Document 文件创建信息
type Document struct {
	Author     *string       `xml:"wpml:author,omitempty"`     // 文件创建作者
	CreateTime *int64        `xml:"wpml:createTime,omitempty"` // 文件创建时间（Unix Timestamp，单位：ms）
	UpdateTime *int64        `xml:"wpml:updateTime,omitempty"` // 文件更新时间（Unix Timestamp，单位：ms）
	Mission    MissionConfig `xml:"wpml:missionConfig"`        // 任务信息
	Folders    []Folder      `xml:"Folder"`                    // 模板信息
}

func (d *Document) GenerateXML() (string, error) {
	// 没有 MissionConfig 信息时，返回错误
	if d.Mission == (MissionConfig{}) {
		return "", errors.New("MissionConfig is required")
	}
	type KML struct {
		XMLName  xml.Name `xml:"kml"`
		XMLNS    string   `xml:"xmlns,attr"`
		XMLNSWP  string   `xml:"xmlns:wpml,attr"`
		Document Document `xml:"Document"`
	}
	k := KML{
		XMLNS:    "http://www.opengis.net/kml/2.2",
		XMLNSWP:  "http://www.dji.com/wpmz/1.0.6",
		Document: *d,
	}
	data, err := xml.MarshalIndent(k, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(data), nil
}

// MissionConfig 任务信息
type MissionConfig struct {
	FlyToWaylineMode         FlyToWaylineMode `xml:"wpml:flyToWaylineMode"`                   // 飞向首航点模式
	FinishAction             FinishAction     `xml:"wpml:finishAction"`                       // 航线结束动作
	ExitOnRCLost             ExitOnRCLost     `xml:"wpml:exitOnRCLost"`                       // 失控是否继续执行航线
	ExecuteRCLostAction      RCLostAction     `xml:"wpml:executeRCLostAction"`                // 失控动作类型
	TakeOffSecurityHeight    float64          `xml:"wpml:takeOffSecurityHeight"`              // 安全起飞高度（单位：m）
	GlobalTransitionalSpeed  float64          `xml:"wpml:globalTransitionalSpeed"`            // 全局航线过渡速度（单位：m/s）
	GlobalRTHHeight          float64          `xml:"wpml:globalRTHHeight,omitempty"`          // 全局返航高度（单位：m）
	TakeOffRefPoint          *string          `xml:"wpml:takeOffRefPoint,omitempty"`          // 参考起飞点（格式：纬度,经度,高度）
	TakeOffRefPointAGLHeight *float64         `xml:"wpml:takeOffRefPointAGLHeight,omitempty"` // 参考起飞点海拔高度（单位：m）
	DroneInfo                DroneInfo        `xml:"wpml:droneInfo"`                          // 飞行器机型信息
	PayloadInfo              PayloadInfo      `xml:"wpml:payloadInfo"`                        // 负载机型信息
	AutoRerouteInfo          *AutoRerouteInfo `xml:"wpml:autoRerouteInfo,omitempty"`          // 航线绕行信息
}

// DefaultMissionConfig 生成默认的任务信息
func DefaultMissionConfig(drone DroneInfo, payload PayloadInfo) MissionConfig {
	return MissionConfig{
		FlyToWaylineMode:        FlyToWaylineSafely,
		FinishAction:            FinishActionGoHome,
		ExitOnRCLost:            ExitOnRCLostExecuteLostAction,
		ExecuteRCLostAction:     RCLostActionGoBack,
		TakeOffSecurityHeight:   20,
		GlobalTransitionalSpeed: 15,
		DroneInfo:               drone,
		PayloadInfo:             payload,
		AutoRerouteInfo:         nil,
	}
}

// FlyToWaylineMode 飞向首航点模式枚举
type FlyToWaylineMode string

const (
	FlyToWaylineSafely       FlyToWaylineMode = "safely"       // 安全模式
	FlyToWaylinePointToPoint FlyToWaylineMode = "pointToPoint" // 倾斜飞行模式
)

// FinishAction 航线结束动作枚举
type FinishAction string

const (
	FinishActionGoHome            FinishAction = "goHome"            // 返航
	FinishActionNoAction          FinishAction = "noAction"          // 无动作
	FinishActionAutoLand          FinishAction = "autoLand"          // 原地降落
	FinishActionGotoFirstWaypoint FinishAction = "gotoFirstWaypoint" // 飞向起始点
)

// ExitOnRCLost 失控是否继续执行航线枚举
type ExitOnRCLost string

const (
	ExitOnRCLostGoContinue        ExitOnRCLost = "goContinue"        // 继续执行航线
	ExitOnRCLostExecuteLostAction ExitOnRCLost = "executeLostAction" // 执行失控动作
)

// RCLostAction 失控动作类型枚举
type RCLostAction string

const (
	RCLostActionGoBack  RCLostAction = "goBack"  // 返航
	RCLostActionLanding RCLostAction = "landing" // 降落
	RCLostActionHover   RCLostAction = "hover"   // 悬停
)

// Folder 模板信息
type Folder struct {
	TemplateType              *TemplateType              `xml:"wpml:templateType,omitempty"`              // 预定义模板类型
	TemplateID                *int                       `xml:"wpml:templateId,omitempty"`                // 模板ID
	AutoFlightSpeed           float64                    `xml:"wpml:autoFlightSpeed"`                     // 全局航线飞行速度（单位：m/s）
	WaylineCoordinateSysParam *WaylineCoordinateSysParam `xml:"wpml:waylineCoordinateSysParam,omitempty"` // 坐标系参数
	PayloadParam              *PayloadParam              `xml:"wpml:payloadParam,omitempty"`              // 负载设置
	Distance                  *float64                   `xml:"wpml:distance,omitempty"`                  // 航线长度（单位：m）
	Duration                  *float64                   `xml:"wpml:duration,omitempty"`                  // 预计飞行时间（单位：s）
	CaliFlightEnable          BoolAsInt                  `xml:"wpml:caliFlightEnable,omitempty"`          // 是否开启校飞
	// wayline.wpml元素
	WaylineID         *int               `xml:"wpml:waylineId,omitempty"`         // 航线ID
	ExecuteHeightMode *ExecuteHeightMode `xml:"wpml:executeHeightMode,omitempty"` // 执行高度模式
	StartActionGroup  *ActionGroup       `xml:"wpml:startActionGroup,omitempty"`  // 起飞动作组
	//	航点飞行模板元素
	GlobalWaypointTurnMode     *WaypointTurnMode     `xml:"wpml:globalWaypointTurnMode,omitempty"`     // 全局航点类型
	GlobalUseStraightLine      *BoolAsInt            `xml:"wpml:globalUseStraightLine,omitempty"`      // 全局航段轨迹是否尽量贴合直线
	GimbalPitchMode            *GimbalPitchMode      `xml:"wpml:gimbalPitchMode,omitempty"`            // 云台俯仰角控制模式
	GlobalHeight               *float64              `xml:"wpml:globalHeight,omitempty"`               // 全局航线高度（相对起飞点高度）
	GlobalWaypointHeadingParam *WaypointHeadingParam `xml:"wpml:globalWaypointHeadingParam,omitempty"` // 全局偏航角模式参数
	// 航点信息元素
	Placemarks []Placemark `xml:"Placemark"` // 航点信息
}

func DefaultWaypointFolder(id int) Folder {
	t := TemplateTypeWaypoint
	waylineCoordinateSysParam := DefaultWaylineCoordinateSysParam()
	globalHeight := 50.0
	gimbalPitchMode := GimbalPitchModeManual
	globalWaypointHeadingParam := DefaultWaypointHeadingParam()
	globalWaypointTurnMode := ToPointAndStopWithDiscontinuityCurvature
	globalUseStraightLine := BoolAsInt(true)
	payloadParam := DefaultPayloadParam()
	return Folder{
		TemplateType:              &t,
		TemplateID:                &id,
		WaylineID:                 nil,
		AutoFlightSpeed:           5,
		WaylineCoordinateSysParam: &waylineCoordinateSysParam,
		PayloadParam:              &payloadParam,
		ExecuteHeightMode:         nil,
		CaliFlightEnable:          BoolAsInt(false),
		// 航点飞行模板元素
		GlobalWaypointTurnMode:     &globalWaypointTurnMode,
		GlobalUseStraightLine:      &globalUseStraightLine,
		GimbalPitchMode:            &gimbalPitchMode,
		GlobalHeight:               &globalHeight,
		GlobalWaypointHeadingParam: &globalWaypointHeadingParam,
		StartActionGroup:           nil,
	}
}

// TemplateType 预定义模板类型枚举
type TemplateType string

const (
	TemplateTypeWaypoint     TemplateType = "waypoint"     // 航点飞行
	TemplateTypeMapping2D    TemplateType = "mapping2d"    // 建图航拍
	TemplateTypeMapping3D    TemplateType = "mapping3d"    // 倾斜摄影
	TemplateTypeMappingStrip TemplateType = "mappingStrip" // 航带飞行
)

// ExecuteHeightMode 执行高度模式枚举
type ExecuteHeightMode string

const (
	ExecuteHeightModeWGS84                 ExecuteHeightMode = "WGS84"                 // 椭球高模式
	ExecuteHeightModeRelativeToStartPoint  ExecuteHeightMode = "relativeToStartPoint"  // 相对起飞点高度模式
	ExecuteHeightModeRealTimeFollowSurface ExecuteHeightMode = "realTimeFollowSurface" // 使用实时仿地模式
)

// GimbalPitchMode 云台俯仰角控制模式枚举
type GimbalPitchMode string

const (
	GimbalPitchModeManual          GimbalPitchMode = "manual"          // 手动控制
	GimbalPitchModeUsePointSetting GimbalPitchMode = "usePointSetting" // 依照每个航点设置
)

// WaylineCoordinateSysParam 坐标系参数
type WaylineCoordinateSysParam struct {
	CoordinateMode          CoordinateMode  `xml:"wpml:coordinateMode"`                    // 经纬度坐标系
	HeightMode              HeightMode      `xml:"wpml:heightMode"`                        // 航点高程参考平面
	PositioningType         PositioningType `xml:"wpml:positioningType,omitempty"`         // 经纬度与高度数据源
	GlobalShootHeight       float64         `xml:"wpml:globalShootHeight,omitempty"`       // 飞行器离被摄面高度（单位：m）
	SurfaceFollowModeEnable BoolAsInt       `xml:"wpml:surfaceFollowModeEnable,omitempty"` // 是否开启仿地飞行
	SurfaceRelativeHeight   float64         `xml:"wpml:surfaceRelativeHeight,omitempty"`   // 仿地飞行离地高度（单位：m）
}

func DefaultWaylineCoordinateSysParam() WaylineCoordinateSysParam {
	return WaylineCoordinateSysParam{
		CoordinateMode:  CoordinateModeWGS84,
		HeightMode:      HeightModeRelativeToStartPoint,
		PositioningType: PositioningTypeGPS,
	}
}

// CoordinateMode 经纬度坐标系枚举
type CoordinateMode string

const (
	CoordinateModeWGS84 CoordinateMode = "WGS84" // WGS84坐标系
)

// HeightMode 航点高程参考平面枚举
type HeightMode string

const (
	HeightModeEGM96                 HeightMode = "EGM96"                 // 使用海拔高编辑
	HeightModeRelativeToStartPoint  HeightMode = "relativeToStartPoint"  // 使用相对点的高度进行编辑
	HeightModeAboveGroundLevel      HeightMode = "aboveGroundLevel"      // 使用地形数据，AGL下编辑
	HeightModeRealTimeFollowSurface HeightMode = "realTimeFollowSurface" // 使用实时仿地模式
)

// PositioningType 经纬度与高度数据源枚举
type PositioningType string

const (
	PositioningTypeGPS            PositioningType = "GPS"            // GPS/BDS/GLONASS/GALILEO等
	PositioningTypeRTKBaseStation PositioningType = "RTKBaseStation" // RTK基站差分定位
	PositioningTypeQianXun        PositioningType = "QianXun"        // 千寻网络RTK
	PositioningTypeCustom         PositioningType = "Custom"         // 自定义网络RTK
)

type WaypointGimbalHeadingParam struct {
	WaypointGimbalPitchAngle float64 `xml:"wpml:waypointGimbalPitchAngle"` // 航点云台俯仰角（单位：°）
	WaypointGimbalYawAngle   float64 `xml:"wpml:waypointGimbalYawAngle"`   // 航点云台偏航角（单位：°）
}

// WaypointWorkType 航点工作类型枚举
type WaypointWorkType int

const (
	WaypointWorkTypeNone WaypointWorkType = 0 // 无
)

// Placemark 航点信息
type Placemark struct {
	// 航点信息
	Point                      Point                       `xml:"Point"`                                     // 航点经纬度
	Index                      int                         `xml:"wpml:index"`                                // 航点序号
	UseGlobalHeight            *BoolAsInt                  `xml:"wpml:useGlobalHeight,omitempty"`            // 是否使用全局高度
	EllipsoidHeight            *float64                    `xml:"wpml:ellipsoidHeight,omitempty"`            // 航点高度（WGS84椭球高度，单位：m）
	Height                     *float64                    `xml:"wpml:height,omitempty"`                     // 航点高度（EGM96海拔高度/相对起飞点高度/AGL相对地面高度，单位：m）
	ExecuteHeight              *float64                    `xml:"wpml:executeHeight,omitempty"`              // 航点执行高度（单位：m）
	UseGlobalSpeed             *BoolAsInt                  `xml:"wpml:useGlobalSpeed,omitempty"`             // 是否使用全局飞行速度
	WaypointSpeed              float64                     `xml:"wpml:waypointSpeed,omitempty"`              // 航点飞行速度（单位：m/s）
	UseGlobalHeadingParam      *BoolAsInt                  `xml:"wpml:useGlobalHeadingParam,omitempty"`      // 是否使用全局偏航角模式参数
	WaypointHeadingParam       *WaypointHeadingParam       `xml:"wpml:waypointHeadingParam,omitempty"`       // 偏航角模式参数
	UseGlobalTurnParam         *BoolAsInt                  `xml:"wpml:useGlobalTurnParam,omitempty"`         // 是否使用全局航点类型
	WaypointTurnParam          *WaypointTurnParam          `xml:"wpml:waypointTurnParam,omitempty"`          // 航点类型
	UseStraightLine            *BoolAsInt                  `xml:"wpml:useStraightLine,omitempty"`            // 该航段是否贴合直线
	GimbalPitchAngle           float64                     `xml:"wpml:gimbalPitchAngle,omitempty"`           // 航点云台俯仰角（单位：°）
	IsRisky                    *BoolAsInt                  `xml:"wpml:isRisky,omitempty"`                    // 是否为高风险航点
	WaypointWorkType           *WaypointWorkType           `xml:"wpml:waypointWorkType,omitempty"`           // 航点工作类型
	WaypointGimbalHeadingParam *WaypointGimbalHeadingParam `xml:"wpml:waypointGimbalHeadingParam,omitempty"` // 航点云台角度
	ActionGroup                *ActionGroup                `xml:"wpml:actionGroup,omitempty"`                // 动作组
}

func DefaultPlacemark(lng, lat float64) Placemark {
	trueBool := BoolAsInt(true)
	falseBool := BoolAsInt(false)
	height := 50.0
	ellipsoidHeight := 50.0
	return Placemark{
		IsRisky:               &falseBool,
		Point:                 Point{Coordinates: FormatCoordinates(lng, lat)},
		UseGlobalHeight:       &trueBool,
		Height:                &height,
		EllipsoidHeight:       &ellipsoidHeight,
		UseGlobalSpeed:        &trueBool,
		UseGlobalHeadingParam: &trueBool,
		WaypointHeadingParam:  nil,
		UseGlobalTurnParam:    &trueBool,
		WaypointTurnParam:     nil,
		UseStraightLine:       &falseBool,
		GimbalPitchAngle:      0,
		ActionGroup:           nil,
	}
}

// Point 航点经纬度
type Point struct {
	Coordinates string `xml:"coordinates"` // 经纬度（格式：经度,纬度）
}

// DroneEnumValue 飞行器机型主类型枚举
type DroneEnumValue int

const (
	DroneM350RTK  DroneEnumValue = 89 // 机型：M350 RTK
	DroneM300RTK  DroneEnumValue = 60 // 机型：M300 RTK
	DroneM30      DroneEnumValue = 67 // 机型：M30/M30T
	DroneM3Series DroneEnumValue = 77 // 机型：M3E/M3T/M3M
	DroneM3D      DroneEnumValue = 91 // 机型：M3D/M3TD
)

// DroneSubEnumValue 飞行器机型子类型枚举
type DroneSubEnumValue int

const (
	// SubM30 当主类型为67时的子类型
	SubM30  DroneSubEnumValue = 0 // 机型：M30双光
	SubM30T DroneSubEnumValue = 1 // 机型：M30T三光

	// SubM3E 当主类型为77时的子类型
	SubM3E DroneSubEnumValue = 0 // 机型：M3E
	SubM3T DroneSubEnumValue = 1 // 机型：M3T
	SubM3M DroneSubEnumValue = 2 // 机型：M3M

	// SubM3D 当主类型为91时的子类型
	SubM3D  DroneSubEnumValue = 0 // 机型：M3D
	SubM3TD DroneSubEnumValue = 1 // 机型：M3TD
)

// DroneInfo 共用元素：<wpml:droneInfo>
type DroneInfo struct {
	DroneEnumValue    DroneEnumValue    `xml:"wpml:droneEnumValue"`    // 飞行器机型主类型，必需元素
	DroneSubEnumValue DroneSubEnumValue `xml:"wpml:droneSubEnumValue"` // 飞行器机型子类型，必需元素
}

// InferenceByModel 根据机型推断飞行器类型
func (d *DroneInfo) InferenceByModel(model string) (ok bool) {
	switch model {
	case "M3E":
		d.DroneEnumValue = DroneM3Series
		d.DroneSubEnumValue = SubM3E
	case "M3T":
		d.DroneEnumValue = DroneM3Series
		d.DroneSubEnumValue = SubM3T
	case "M3M":
		d.DroneEnumValue = DroneM3Series
		d.DroneSubEnumValue = SubM3M
	default:
		return false
	}
	return true
}

// PayloadEnumValue 负载机型主类型枚举
type PayloadEnumValue int

const (
	PayloadH20       PayloadEnumValue = 42 // H20
	PayloadH20T      PayloadEnumValue = 43 // H20T
	PayloadM30Dual   PayloadEnumValue = 52 // M30双光相机
	PayloadM30Triple PayloadEnumValue = 53 // M30T三光相机
	PayloadM3E       PayloadEnumValue = 66 // M3E
	// ...其他枚举值需根据文档补充完整
)

// PayloadSubEnumValue 负载机型子类型枚举
type PayloadSubEnumValue int

const (
	PayloadSubM3E PayloadSubEnumValue = 0 // M3E
)

// PayloadInfo 共用元素：<wpml:payloadInfo>
type PayloadInfo struct {
	PayloadEnumValue     PayloadEnumValue    `xml:"wpml:payloadEnumValue"`     // 负载机型主类型，必需元素
	PayloadSubEnumValue  PayloadSubEnumValue `xml:"wpml:payloadSubEnumValue"`  // 负载机型子类型，必需元素
	PayloadPositionIndex int                 `xml:"wpml:payloadPositionIndex"` // 负载挂载位置，必需元素。0 对应主云台（M300 RTK、M350 RTK对应左前方），1、2 对应 M300 RTK 和 M350 RTK 的右前方和左后方
}

// ImageFormat 图像格式枚举
type ImageFormat string

const (
	ImageWide       ImageFormat = "wide"        // 存储广角镜头照片
	ImageZoom       ImageFormat = "zoom"        // 存储变焦镜头照片
	ImageIR         ImageFormat = "ir"          // 存储红外镜头照片
	ImageNarrowBand ImageFormat = "narrow_band" // 存储窄带镜头拍摄照片
	ImageVisible    ImageFormat = "visible"     // 可见光照片
)

// FocusMode 负载对焦模式枚举
type FocusMode string

const (
	FocusModeFirstPoint FocusMode = "firstPoint" // 首个航点自动对焦
	FocusModeCustom     FocusMode = "custom"     // 标定对焦值对焦
)

// MeteringMode 负载测光模式枚举
type MeteringMode string

const (
	MeteringModeAverage MeteringMode = "average" // 全局测光
	MeteringModeSpot    MeteringMode = "spot"    // 点测光
)

// ReturnMode 激光雷达回波模式枚举
type ReturnMode string

const (
	ReturnModeSingle ReturnMode = "singleReturnStrongest" // 单回波
	ReturnModeDual   ReturnMode = "dualReturn"            // 双回波
	ReturnModeTriple ReturnMode = "tripleReturn"          // 三回波
)

// ScanningMode 负载扫描模式枚举
type ScanningMode string

const (
	ScanningModeRepetitive    ScanningMode = "repetitive"    // 重复扫描
	ScanningModeNonRepetitive ScanningMode = "nonRepetitive" // 非重复扫描
)

// PayloadParam 负载参数：<wpml:payloadParam>
type PayloadParam struct {
	PayloadPositionIndex int           `xml:"wpml:payloadPositionIndex"`          // 负载挂载位置，必需元素
	FocusMode            FocusMode     `xml:"wpml:focusMode,omitempty"`           // 负载对焦模式
	MeteringMode         MeteringMode  `xml:"wpml:meteringMode,omitempty"`        // 负载测光模式
	DewarpingEnable      int           `xml:"wpml:dewarpingEnable,omitempty"`     // 是否开启畸变矫正 0:关 1:开
	ReturnMode           ReturnMode    `xml:"wpml:returnMode,omitempty"`          // 激光雷达回波模式
	SamplingRate         int           `xml:"wpml:samplingRate,omitempty"`        // 负载采样率（Hz）
	ScanningMode         ScanningMode  `xml:"wpml:scanningMode,omitempty"`        // 负载扫描模式
	ModelColoringEnable  int           `xml:"wpml:modelColoringEnable,omitempty"` // 真彩上色 0:关 1:开
	ImageFormat          []ImageFormat `xml:"wpml:imageFormat"`                   // 图片格式列表，必需元素
}

func DefaultPayloadParam() PayloadParam {
	return PayloadParam{
		PayloadPositionIndex: 0, // 默认挂载位置索引
		//ImageFormat:          []ImageFormat{ImageWide}, // 默认存储广角照片
		//FocusMode:            FocusModeFirstPoint,      // 默认首个航点自动对焦
		//MeteringMode:         MeteringModeAverage,      // 默认全局测光
		//ReturnMode:           ReturnModeSingle,         // 默认单回波
		//ScanningMode:         ScanningModeRepetitive,   // 默认重复扫描
		//SamplingRate:         240 * 1000,               // 默认采样率240kHz
	}
}

// HeadingMode 飞行器偏航角模式枚举
type HeadingMode string

const (
	HeadingFollowWayline    HeadingMode = "followWayline"    // 沿航线方向
	HeadingManually         HeadingMode = "manually"         // 手动控制
	HeadingFixed            HeadingMode = "fixed"            // 锁定当前偏航角
	HeadingSmoothTransition HeadingMode = "smoothTransition" // 自定义
	HeadingTowardPOI        HeadingMode = "towardPOI"        // 朝向兴趣点
)

// WaypointHeadingPathMode 偏航角转动方向枚举
type WaypointHeadingPathMode string

const (
	Clockwise        WaypointHeadingPathMode = "clockwise"        // 顺时针旋转飞行器偏航角
	CounterClockwise WaypointHeadingPathMode = "counterClockwise" // 逆时针旋转飞行器偏航角
	FollowBadArc     WaypointHeadingPathMode = "followBadArc"     // 沿最短路径旋转飞行器偏航角
)

// WaypointHeadingParam 航点偏航参数
type WaypointHeadingParam struct {
	WaypointHeadingMode     HeadingMode              `xml:"wpml:waypointHeadingMode"`               // 飞行器偏航角模式，必需元素
	WaypointHeadingAngle    *float64                 `xml:"wpml:waypointHeadingAngle,omitempty"`    // 偏航角度（仅smoothTransition模式需要）
	WaypointPoiPoint        *string                  `xml:"wpml:waypointPoiPoint,omitempty"`        // 兴趣点坐标（格式：纬度,经度,高度）
	WaypointHeadingPoiIndex *int                     `xml:"wpml:waypointHeadingPoiIndex,omitempty"` // 兴趣点索引
	WaypointHeadingPathMode *WaypointHeadingPathMode `xml:"wpml:waypointHeadingPathMode,omitempty"` // 偏航角转动方向
}

func DefaultWaypointHeadingParam() WaypointHeadingParam {
	angle := 0.0
	point := "0.000000,0.000000,0.000000"
	idx := 0
	return WaypointHeadingParam{
		WaypointHeadingMode:     HeadingFollowWayline,
		WaypointHeadingAngle:    &angle,
		WaypointPoiPoint:        &point,
		WaypointHeadingPoiIndex: &idx,
	}
}

// WaypointTurnMode 航点类型（航点转弯模式）枚举
type WaypointTurnMode string

const (
	CoordinateTurn                           WaypointTurnMode = "coordinateTurn"                           // 协调转弯，飞行器过点不停
	ToPointAndStopWithDiscontinuityCurvature WaypointTurnMode = "toPointAndStopWithDiscontinuityCurvature" // 直线飞行，飞行器到点停
	ToPointAndStopWithContinuityCurvature    WaypointTurnMode = "toPointAndStopWithContinuityCurvature"    // 曲线飞行，飞行器到点停
	ToPointAndPassWithContinuityCurvature    WaypointTurnMode = "toPointAndPassWithContinuityCurvature"    // 曲线飞行，飞行器过点不停
)

// WaypointTurnParam 航点转弯参数
type WaypointTurnParam struct {
	WaypointTurnMode        WaypointTurnMode `xml:"wpml:waypointTurnMode"`        // 航点转弯模式
	WaypointTurnDampingDist float64          `xml:"wpml:waypointTurnDampingDist"` // 转弯阻尼距离
}

// AutoRerouteInfo 自动绕行信息
type AutoRerouteInfo struct {
	MissionAutoRerouteMode      BoolAsInt `xml:"wpml:missionAutoRerouteMode"`      // 任务航线绕行模式
	TransitionalAutoRerouteMode BoolAsInt `xml:"wpml:transitionalAutoRerouteMode"` // 过渡航线绕行模式
}

// ActionGroupMode 动作组执行模式枚举
type ActionGroupMode string

const ActionGroupModeSequence ActionGroupMode = "sequence"

// ActionGroup 动作组
type ActionGroup struct {
	ActionGroupId         int             `xml:"wpml:actionGroupId"`         // 动作组ID，必需元素
	ActionGroupStartIndex int             `xml:"wpml:actionGroupStartIndex"` // 开始生效航点，必需元素
	ActionGroupEndIndex   int             `xml:"wpml:actionGroupEndIndex"`   // 结束生效航点，必需元素
	ActionGroupMode       ActionGroupMode `xml:"wpml:actionGroupMode"`       // 执行模式（enum: sequence）
	ActionTrigger         ActionTrigger   `xml:"wpml:actionTrigger"`         // 触发器
	Actions               []Action        `xml:"wpml:action"`                // 动作列表
}

func DefaultActionGroup(id, start, end int) ActionGroup {
	return ActionGroup{
		ActionGroupId:         id,
		ActionGroupStartIndex: start,
		ActionGroupEndIndex:   end,
		ActionGroupMode:       ActionGroupModeSequence,
		ActionTrigger:         ActionTrigger{TriggerType: TriggerReachPoint},
	}
}

// ActionTriggerType 动作触发器类型枚举
type ActionTriggerType string

const (
	TriggerReachPoint       ActionTriggerType = "reachPoint"
	TriggerBetweenPoints    ActionTriggerType = "betweenAdjacentPoints"
	TriggerMultipleTiming   ActionTriggerType = "multipleTiming"
	TriggerMultipleDistance ActionTriggerType = "multipleDistance"
)

// ActionTrigger 动作触发器
type ActionTrigger struct {
	TriggerType  ActionTriggerType `xml:"wpml:actionTriggerType"`            // 触发器类型
	TriggerParam float64           `xml:"wpml:actionTriggerParam,omitempty"` // 触发参数（单位：秒或米）
}

// ActionActuatorFuncType 动作类型枚举
type ActionActuatorFuncType string

const (
	ActionTakePhoto          ActionActuatorFuncType = "takePhoto"          // 单拍
	ActionStartRecord        ActionActuatorFuncType = "startRecord"        // 开始录像
	ActionStopRecord         ActionActuatorFuncType = "stopRecord"         // 结束录像
	ActionFocus              ActionActuatorFuncType = "focus"              // 对焦
	ActionZoom               ActionActuatorFuncType = "zoom"               // 变焦
	ActionCustomDir          ActionActuatorFuncType = "customDirName"      // 创建新文件夹
	ActionGimbalRotate       ActionActuatorFuncType = "gimbalRotate"       // 旋转云台
	ActionRotateYaw          ActionActuatorFuncType = "rotateYaw"          // 飞行器偏航
	ActionHover              ActionActuatorFuncType = "hover"              // 悬停等待
	ActionGimbalEvenlyRotate ActionActuatorFuncType = "gimbalEvenlyRotate" // 航段间均匀转动云台pitch角
	ActionAccurateShoot      ActionActuatorFuncType = "accurateShoot"      // 精准复拍动作
	ActionOrientedShoot      ActionActuatorFuncType = "orientedShoot"      // 定向拍照动作
	ActionPanoShot           ActionActuatorFuncType = "panoShot"           // 全景拍照动作
	ActionRecordPointCloud   ActionActuatorFuncType = "recordPointCloud"   // 点云录制操作
)

// Action 动作结构
type Action struct {
	ActionId     int                    `xml:"wpml:actionId"`                // 动作ID，必需元素
	ActionType   ActionActuatorFuncType `xml:"wpml:actionActuatorFunc"`      // 动作类型，必需元素
	ActionParams interface{}            `xml:"wpml:actionActuatorFuncParam"` // 动作参数（根据类型动态变化）
}

// TakePhotoParams 拍照动作参数
type TakePhotoParams struct {
	PayloadPositionIndex      int       `xml:"wpml:payloadPositionIndex"`      // 负载挂载位置
	FileSuffix                string    `xml:"wpml:fileSuffix"`                // 文件后缀
	PayloadLensIndex          []string  `xml:"wpml:payloadLensIndex"`          // 存储类型列表
	UseGlobalPayloadLensIndex BoolAsInt `xml:"wpml:useGlobalPayloadLensIndex"` // 是否使用全局设置
}

// StartRecordParams 录像动作参数
type StartRecordParams struct {
	PayloadPositionIndex      int       `xml:"wpml:payloadPositionIndex"`      // 负载挂载位置
	FileSuffix                string    `xml:"wpml:fileSuffix"`                // 文件后缀
	PayloadLensIndex          []string  `xml:"wpml:payloadLensIndex"`          // 存储类型列表
	UseGlobalPayloadLensIndex BoolAsInt `xml:"wpml:useGlobalPayloadLensIndex"` // 是否使用全局设置
}

type GimbalRotateParams struct {
	PayloadPositionIndex    int       `xml:"wpml:payloadPositionIndex"`    // 负载挂载位置
	GimbalHeadingYawBase    string    `xml:"wpml:gimbalHeadingYawBase"`    // 云台偏航角转动坐标系
	GimbalRotateMode        string    `xml:"wpml:gimbalRotateMode"`        // 云台转动模式
	GimbalPitchRotateEnable BoolAsInt `xml:"wpml:gimbalPitchRotateEnable"` // 是否使能云台Pitch转动
	GimbalPitchRotateAngle  float64   `xml:"wpml:gimbalPitchRotateAngle"`  // 云台Pitch转动角度
	GimbalRollRotateEnable  BoolAsInt `xml:"wpml:gimbalRollRotateEnable"`  // 是否使能云台Roll转动
	GimbalRollRotateAngle   float64   `xml:"wpml:gimbalRollRotateAngle"`   // 云台Roll转动角度
	GimbalYawRotateEnable   BoolAsInt `xml:"wpml:gimbalYawRotateEnable"`   // 是否使能云台Yaw转动
	GimbalYawRotateAngle    float64   `xml:"wpml:gimbalYawRotateAngle"`    // 云台Yaw转动角度
	GimbalRotateTimeEnable  BoolAsInt `xml:"wpml:gimbalRotateTimeEnable"`  // 是否使能云台转动时间
	GimbalRotateTime        float64   `xml:"wpml:gimbalRotateTime"`        // 云台完成转动用时
}

type ZoomParams struct {
	PayloadPositionIndex int     `xml:"wpml:payloadPositionIndex"` // 负载挂载位置
	FocalLength          float64 `xml:"wpml:focalLength"`          // 变焦焦距（单位：mm）
}

type HoverParams struct {
	HoverTime int `xml:"wpml:hoverTime"` // 悬停等待时间（单位：秒）
}
