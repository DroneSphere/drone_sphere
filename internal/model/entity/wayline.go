package entity

import "github.com/dronesphere/internal/model/vo"

type Wayline struct {
	ID         uint          `json:"id" gorm:"primary_key"`
	UploadUser string        `json:"upload_user"`
	Area       SearchArea    `json:"area"`
	Drone      Drone         `json:"drone"`
	Points     []vo.GeoPoint `json:"points"`
	S3Key      string        `json:"s3_key"`
	S3Url      string        `json:"s3_url"`
}
