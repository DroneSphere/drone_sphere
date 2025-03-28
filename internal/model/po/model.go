package po

import (
	"fmt"

	"github.com/dronesphere/internal/pkg/misc"
	"gorm.io/gorm"
)

type GatewayModel struct {
	misc.BaseModel
	// 型号名称，DJI 文档中收录的标准名称
	Name string `json:"name"`
	// 描述，自定义的型号描述
	Description string `json:"description,omitempty"`
	// 领域，DJI 文档指定
	Domain int `json:"domain"`
	// 主型号，DJI 文档指定
	Type int `json:"type"`
	// 子型号，DJI 文档指定
	SubType int `json:"sub_type"`
}

type DroneModel struct {
	misc.BaseModel
	// 型号名称
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// 领域
	Domain int `json:"domain"`
	// 主类型
	Type int `json:"type"`
	// 自类型
	SubType int `json:"sub_type"`
	// 对应的网关ID
	GatewayID uint `json:"gateway_id"`
	// 可搭载云台
	Gimbals        []GimbalModel  `json:"gimbals,omitempty" gorm:"many2many:drone_gimbal;"`
	Payloads       []PayloadModel `json:"payloads,omitempty" gorm:"many2many:drone_payload;"`
	IsRTKAvailable bool           `json:"is_rtk_available" gorm:"default:false"`
}

type GimbalModel struct {
	misc.BaseModel
	// 云台名称
	Name string `json:"name"`
	// 描述
	Description string `json:"description"`
	// 产品线名称，FPV相机、云台相机和机场相机
	Product string `json:"product"`
	// 领域
	Domain int `json:"domain"`
	// 型号
	Type int `json:"type"`
	// 子型号
	SubType int `json:"sub_type"`
	// 相机位置索引
	Gimbalindex int `json:"gimbalindex"`
	// 对应的无人机
	Drones             []DroneModel `json:"drones,omitempty" gorm:"many2many:drone_gimbal;"`
	IsThermalAvailable bool         `json:"is_thermal_available" gorm:"default:false"`
}

type PayloadModel struct {
	misc.BaseModel
	Name           string       `json:"name"`
	Description    string       `json:"description,omitempty"`
	Category       string       `json:"category"`
	Drones         []DroneModel `json:"drones,omitempty" gorm:"many2many:drone_payload;"`
	IsRTKAvailable bool         `json:"is_rtk_available" gorm:"default:false"`
}

type DroneVariation struct {
	misc.BaseModel
	// 变体名称，可自动生成或自定义
	Name string `json:"name"`
	// 变体描述
	Description string `json:"description,omitempty"`
	// 关联的无人机型号ID
	DroneModelID uint `json:"drone_model_id" gorm:"index"`
	// 关联的无人机型号
	DroneModel DroneModel `json:"drone_model" gorm:"foreignKey:DroneModelID"`
	// 关联的云台型号，多对多关系
	Gimbals []GimbalModel `json:"gimbals" gorm:"many2many:variation_gimbal;"`
	// 关联的负载型号，多对多关系
	Payloads []PayloadModel `json:"payloads" gorm:"many2many:variation_payload;"`
	// 是否为有效配置（部分组合可能在物理上不兼容）
	IsValid bool `json:"is_valid" gorm:"default:true"`
}

func (dv *DroneVariation) SupportsRTK() bool {
	if len(dv.Gimbals) == 0 {
		return false
	}
	return dv.DroneModel.IsRTKAvailable || dv.Gimbals[0].IsThermalAvailable
}

func (dv *DroneVariation) SupportsThermal() bool {
	if len(dv.Gimbals) == 0 {
		return false
	}
	return dv.Gimbals[0].IsThermalAvailable
}

// 生成一个无人机型号的所有可能配置组合
func GenerateDroneVariations(db *gorm.DB, droneModelID uint) ([]DroneVariation, error) {
	var variations []DroneVariation
	drone := DroneModel{}

	if err := db.Preload("Gimbals").Preload("Payloads").First(&drone, droneModelID).Error; err != nil {
		return variations, err
	}

	// // 如果没有云台或负载，创建一个基础变体
	// if len(drone.Gimbals) == 0 && len(drone.Payloads) == 0 {
	// 	variations = append(variations, DroneVariation{
	// 		DroneModelID: drone.ID,
	// 		Name:         drone.Name + " 基础配置",
	// 	})
	// 	return variations, nil
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
		for _, payloadCombo := range payloadCombos {
			// 跳过完全没有配置的情况
			if len(gimbalCombo) == 0 && len(payloadCombo) == 0 {
				continue
			}

			name := fmt.Sprintf("%s - %s", drone.Name, gimbalCombo[0].Name)
			if len(payloadCombo) > 0 {
				name = fmt.Sprintf("%s (携带%s)", name, payloadCombo[0].Name)
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

// // 检查无人机型号是否存在特定云台和负载组合的变体
// func HasVariation(db *gorm.DB, droneModelID uint, gimbalIDs, payloadIDs []uint) (bool, *DroneVariation, error) {
// 	// 先查询该无人机型号的所有变体
// 	variations, err := GetDroneVariations(db, droneModelID)
// 	if err != nil {
// 		return false, nil, err
// 	}

// 	// 检查每个变体是否匹配给定的云台和负载组合
// 	for _, variation := range variations {
// 		if matchesIDLists(variation.Gimbals, gimbalIDs) && matchesIDLists(variation.Payloads, payloadIDs) {
// 			return true, &variation, nil
// 		}
// 	}

// 	return false, nil, nil
// }

// // 辅助函数：检查模型列表是否匹配ID列表
// func matchesIDLists[T interface{ GetID() uint }](models []T, ids []uint) bool {
// 	if len(models) != len(ids) {
// 		return false
// 	}

// 	// 获取模型ID列表
// 	var modelIDs []uint
// 	for _, model := range models {
// 		modelIDs = append(modelIDs, model.GetID())
// 	}

// 	// 对两个列表进行排序
// 	slices.Sort(modelIDs)
// 	slices.Sort(ids)

// 	// 比较两个列表
// 	return slices.Equal(modelIDs, ids)
// }

// // 为BaseModel添加GetID方法，以便上面的泛型函数能够使用
// func (m *misc.BaseModel) GetID() uint {
// 	return m.ID
// }
