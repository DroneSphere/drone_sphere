package po

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type GatewayModel struct {
	ID          uint      `json:"gateway_model_id" gorm:"primaryKey;column:gateway_model_id"`
	CreatedTime time.Time `json:"created_time" gorm:"autoCreateTime;column:created_time"`
	UpdatedTime time.Time `json:"updated_time" gorm:"autoUpdateTime;column:updated_time"`
	State       int       `json:"state" gorm:"default:0;column:state"` // -1: deleted, 0: active
	// 型号名称，DJI 文档中收录的标准名称
	Name string `json:"gateway_model_name" gorm:"column:gateway_model_name"`
	// 描述，自定义的型号描述
	Description string `json:"gateway_model_description,omitempty" gorm:"column:gateway_model_description"`
	// 领域，DJI 文档指定
	Domain int `json:"gateway_model_domain" gorm:"column:gateway_model_domain"`
	// 主型号，DJI 文档指定
	Type int `json:"gateway_model_type" gorm:"column:gateway_model_type"`
	// 子型号，DJI 文档指定
	SubType int `json:"gateway_model_sub_type" gorm:"column:gateway_model_sub_type"`
}

// TableName 指定 GatewayModel 表名为 tb_gateway_models
func (gm GatewayModel) TableName() string {
	return "tb_gateway_models"
}

type DroneModel struct {
	ID          uint      `json:"drone_model_id" gorm:"primaryKey;column:drone_model_id"`
	CreatedTime time.Time `json:"created_time" gorm:"autoCreateTime;column:created_time"`
	UpdatedTime time.Time `json:"updated_time" gorm:"autoUpdateTime;column:updated_time"`
	State       int       `json:"state" gorm:"default:0;column:state"` // -1: deleted, 0: active
	// 型号名称
	Name        string `json:"drone_model_name" gorm:"column:drone_model_name"`
	Description string `json:"drone_model_description,omitempty" gorm:"column:drone_model_description"`
	// 领域
	Domain int `json:"drone_model_domain" gorm:"column:drone_model_domain"`
	// 主类型
	Type int `json:"drone_model_type" gorm:"column:drone_model_type"`
	// 自类型
	SubType int `json:"drone_model_sub_type" gorm:"column:drone_model_sub_type"`
	// 对应的网关ID
	GatewayID uint `json:"gateway_model_id" gorm:"column:gateway_model_id"`
	// 可搭载云台
	Gimbals        []GimbalModel  `json:"gimbals,omitempty" gorm:"many2many:tb_drone_gimbal;"`
	Payloads       []PayloadModel `json:"payloads,omitempty" gorm:"many2many:tb_drone_payload;"`
	IsRTKAvailable bool           `json:"is_rtk_available" gorm:"default:false;column:is_rtk_available"`
}

// TableName 指定 DroneModel 表名为 tb_drone_models
func (dm DroneModel) TableName() string {
	return "tb_drone_models"
}

type GimbalModel struct {
	ID          uint      `json:"gimbal_model_id" gorm:"primaryKey;column:gimbal_model_id"`
	CreatedTime time.Time `json:"created_time" gorm:"autoCreateTime;column:created_time"`
	UpdatedTime time.Time `json:"updated_time" gorm:"autoUpdateTime;column:updated_time"`
	State       int       `json:"state" gorm:"default:0;column:state"` // -1: deleted, 0: active
	// 云台名称
	Name string `json:"gimbal_model_name" gorm:"column:gimbal_model_name"`
	// 描述
	Description string `json:"gimbal_model_description" gorm:"column:gimbal_model_description"`
	// 产品线名称，FPV相机、云台相机和机场相机
	Product string `json:"gimbal_model_product" gorm:"column:gimbal_model_product"`
	// 领域
	Domain int `json:"gimbal_model_domain" gorm:"column:gimbal_model_domain"`
	// 型号
	Type int `json:"gimbal_model_type" gorm:"column:gimbal_model_type"`
	// 子型号
	SubType int `json:"gimbal_model_sub_type" gorm:"column:gimbal_model_sub_type"`
	// 相机位置索引
	Gimbalindex int `json:"gimbalindex" gorm:"column:gimbalindex"`
	// 对应的无人机
	Drones             []DroneModel `json:"drones,omitempty" gorm:"many2many:tb_drone_gimbal;"`
	IsThermalAvailable bool         `json:"is_thermal_available" gorm:"default:false"`
}

// TableName 指定 GimbalModel 表名为 tb_gimbal_models
func (gm GimbalModel) TableName() string {
	return "tb_gimbal_models"
}

type PayloadModel struct {
	ID          uint      `json:"payload_model_id" gorm:"primaryKey;column:payload_model_id"`
	CreatedTime time.Time `json:"created_time" gorm:"autoCreateTime;column:created_time"`
	UpdatedTime time.Time `json:"updated_time" gorm:"autoUpdateTime;column:updated_time"`
	State       int       `json:"state" gorm:"default:0"`
	Name        string    `json:"payload_model_name"`
	Description string    `json:"payload_model_description,omitempty"`
	Category    string    `json:"category"`
	// 添加主类型字段
	Type int `json:"payload_model_type" gorm:""`
	// 添加子类型字段
	SubType int `json:"payload_model_sub_type" gorm:""`
	// 与无人机型号的多对多关系
	Drones         []DroneModel `json:"drones,omitempty" gorm:"many2many:tb_drone_payload"`
	IsRTKAvailable bool         `json:"is_rtk_available" gorm:"default:false"`
}

type DroneVariation struct {
	ID          uint      `json:"drone_variation_id" gorm:"primaryKey;column:drone_variation_id"`
	CreatedTime time.Time `json:"created_time" gorm:"autoCreateTime;column:created_time"`
	UpdatedTime time.Time `json:"updated_time" gorm:"autoUpdateTime;column:updated_time"`
	State       int       `json:"state" gorm:"default:0;column:drone_variation_state"` // -1: deleted, 0: active	// 变体名称，可自动生成或自定义
	Name        string    `json:"drone_variation_name" gorm:"column:drone_variation_name"`
	// 变体描述
	Description string `json:"drone_variation_description,omitempty" gorm:"column:drone_variation_description"`
	// 关联的无人机型号ID
	DroneModelID uint `json:"drone_model_id" gorm:"index;column:drone_model_id"` // 无人机型号ID
	// 关联的无人机型号
	DroneModel DroneModel `json:"drone_model" gorm:"foreignKey:DroneModelID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"` // 无人机型号信息
	// 关联的云台型号，多对多关系
	Gimbals []GimbalModel `json:"gimbals" gorm:"many2many:tb_variation_gimbal;"`
	// 关联的负载型号，多对多关系
	Payloads []PayloadModel `json:"payloads" gorm:"many2many:tb_variation_payload;"`
	// 是否为有效配置（部分组合可能在物理上不兼容）
	IsValid bool `json:"is_valid" gorm:"default:true"`
}

// TableName 指定 DroneVariation 表名为 tb_drone_variations
func (dv DroneVariation) TableName() string {
	return "tb_drone_variations"
}

func (dv *DroneVariation) SupportsRTK() bool {
	for _, payload := range dv.Payloads {
		if payload.IsRTKAvailable {
			return true
		}
	}
	return dv.DroneModel.IsRTKAvailable
}

func (dv *DroneVariation) SupportsThermal() bool {
	for _, gimbal := range dv.Gimbals {
		if gimbal.IsThermalAvailable {
			return true
		}
	}
	return false
}

// 生成一个无人机型号的所有可能配置组合
func GenerateDroneVariations(db *gorm.DB, droneModelID uint) ([]DroneVariation, error) {
	var variations []DroneVariation
	drone := DroneModel{}

	// 加载无人机型号及其支持的云台和负载
	if err := db.Preload("Gimbals").Preload("Payloads").First(&drone, droneModelID).Error; err != nil {
		return variations, err
	}

	// 同时加载该无人机支持的所有负载型号
	// var supportedPayloads []PayloadModel
	// if err := db.Model(&PayloadModel{}).
	// 	Joins("JOIN payload_drone_support ON tb_payload_models.id = payload_drone_support.payload_model_id").
	// 	Where("payload_drone_support.drone_model_id = ?", droneModelID).
	// 	Find(&supportedPayloads).Error; err != nil {
	// 	return variations, err
	// }

	// // 合并普通关联的负载和特别支持的负载，去重
	// payloadMap := make(map[uint]PayloadModel)
	// for _, p := range drone.Payloads {
	// 	payloadMap[p.ID] = p
	// }
	// for _, p := range supportedPayloads {
	// 	if _, exists := payloadMap[p.ID]; !exists {
	// 		payloadMap[p.ID] = p
	// 		drone.Payloads = append(drone.Payloads, p)
	// 	}
	// }

	// 对云台生成单选组合（包括无云台选项）
	gimbalCombos := [][]GimbalModel{{}} // 包含"无云台"选项
	for _, gimbal := range drone.Gimbals {
		singleGimbal := []GimbalModel{gimbal}
		gimbalCombos = append(gimbalCombos, singleGimbal)
	}

	// 对负载生成所有可能组合，如果没有负载则至少包含一个空组合
	var payloadCombos [][]PayloadModel
	if len(drone.Payloads) == 0 {
		payloadCombos = [][]PayloadModel{{}} // 无负载情况
	} else {
		payloadCombos = generateModelCombinations(drone.Payloads)
		// 添加空负载组合
		payloadCombos = append(payloadCombos, []PayloadModel{})
	}

	// 为每个组合创建变体
	for _, gimbalCombo := range gimbalCombos {
		if len(payloadCombos) == 0 {
			// 如果没有负载，则只生成一个变体
			variation := DroneVariation{
				DroneModelID: drone.ID,
				Gimbals:      gimbalCombo,
				Payloads:     []PayloadModel{},
				Name:         fmt.Sprintf("%s - %s", drone.Name, gimbalCombo[0].Name),
			}
			variations = append(variations, variation)
			continue
		}
		for _, payloadCombo := range payloadCombos {
			// 跳过完全没有配置的情况
			if len(gimbalCombo) == 0 && len(payloadCombo) == 0 {
				continue
			}

			// 生成变体名称
			var name string
			if len(gimbalCombo) > 0 {
				name = fmt.Sprintf("%s - %s", drone.Name, gimbalCombo[0].Name)
				if len(payloadCombo) > 0 {
					name = fmt.Sprintf("%s (携带%s)", name, payloadCombo[0].Name)
				}
			} else if len(payloadCombo) > 0 {
				name = fmt.Sprintf("%s (携带%s)", drone.Name, payloadCombo[0].Name)
			} else {
				name = fmt.Sprintf("%s - 基础版", drone.Name)
			}

			variation := DroneVariation{
				DroneModelID: drone.ID,
				Gimbals:      gimbalCombo,
				Payloads:     payloadCombo,
				Name:         name,
			}

			variations = append(variations, variation)
		}
	}

	// 保存到数据库
	for _, variation := range variations {
		if err := db.Create(&variation).Error; err != nil {
			return nil, err
		}
	}

	return variations, nil
}

// 递归生成所有可能的GimbalModel组合
func generateModelCombinations[T any](models []T) [][]T {
	result := [][]T{{}} // 包含空组合

	for _, model := range models {
		var newCombos [][]T
		for _, combo := range result {
			newCombo := make([]T, len(combo))
			copy(newCombo, combo)
			newCombo = append(newCombo, model)
			newCombos = append(newCombos, newCombo)
		}
		result = append(result, newCombos...)
	}

	return result[1:] // 排除空组合
}

// 查询指定无人机型号的默认变体
func GetDefaultVariation(db *gorm.DB, droneModelID uint) (*DroneVariation, error) {
	var variation DroneVariation
	err := db.Where("drone_model_id = ? AND is_default = ?", droneModelID, true).
		Preload("DroneModel").
		Preload("Gimbals").
		Preload("Payloads").
		First(&variation).Error
	if err != nil {
		return nil, err
	}
	return &variation, nil
}

// 查询指定无人机型号的所有变体
func GetDroneVariations(db *gorm.DB, droneModelID uint) ([]DroneVariation, error) {
	var variations []DroneVariation
	err := db.Where("drone_model_id = ?", droneModelID).
		Preload("DroneModel").
		Preload("Gimbals").
		Preload("Payloads").
		Find(&variations).Error
	return variations, err
}

// 查询所有变体信息
func GetAllDroneVariations(db *gorm.DB) ([]DroneVariation, error) {
	var variations []DroneVariation
	err := db.Preload("DroneModel").
		Preload("Gimbals").
		Preload("Payloads").
		Find(&variations).Error
	return variations, err
}

// 根据条件查询变体
func FindDroneVariations(db *gorm.DB, query map[string]interface{}) ([]DroneVariation, error) {
	var variations []DroneVariation
	err := db.Where(query).
		Preload("DroneModel").
		Preload("Gimbals").
		Preload("Payloads").
		Find(&variations).Error
	return variations, err
}

// 分页查询变体
func GetDroneVariationsWithPagination(db *gorm.DB, page, pageSize int) ([]DroneVariation, int64, error) {
	var variations []DroneVariation
	var total int64

	// 获取总数
	if err := db.Model(&DroneVariation{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Preload("DroneModel").
		Preload("Gimbals").
		Preload("Payloads").
		Offset(offset).
		Limit(pageSize).
		Find(&variations).Error

	return variations, total, err
}

// 设置为默认变体
func (dv *DroneVariation) SetAsDefault(db *gorm.DB) error {
	// 先将所有同一个无人机型号的变体设为非默认
	if err := db.Model(&DroneVariation{}).
		Where("drone_model_id = ?", dv.DroneModelID).
		Update("is_default", false).Error; err != nil {
		return err
	}

	// 然后将当前变体设为默认
	return db.Model(dv).Update("is_default", true).Error
}

// 获取变体的摘要信息
func (dv *DroneVariation) GetSummary() string {
	modelName := "未知"
	if dv.DroneModel.Name != "" {
		modelName = dv.DroneModel.Name
	}

	return fmt.Sprintf("%s (云台: %d, 负载: %d)",
		modelName,
		len(dv.Gimbals),
		len(dv.Payloads))
}

// 删除一个无人机型号的所有变体
func DeleteDroneVariations(db *gorm.DB, droneModelID uint) error {
	return db.Where("drone_model_id = ?", droneModelID).Delete(&DroneVariation{}).Error
}
