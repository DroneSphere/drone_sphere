package po

import (
	"github.com/dronesphere/internal/model/vo"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Wayline struct {
	gorm.Model
	AreaID     uint                             `gorm:"area_id"`
	DroneID    uint                             `gorm:"drone_id"`
	UploadUser string                           `gorm:"upload_user"`
	Points     datatypes.JSONSlice[vo.GeoPoint] `gorm:"points"`
	S3Key      string                           `gorm:"s3_key"`
	S3Url      string                           `gorm:"s3_url"`
}

func (Wayline) TableName() string {
	return "wayline"
}
