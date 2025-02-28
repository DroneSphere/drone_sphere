package v1

type CreateWaylineRequest struct {
	Points  []PointRequest `json:"points" validate:"required"`
	DroneSN string         `json:"drone_sn" validate:"required"`
	Height  float64        `json:"height"`
}

type PointRequest struct {
	Index int     `json:"index"`
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
}

type WaylineItemResult struct {
	ID         uint   `json:"id"`
	DroneModel string `json:"drone_model"`
	DroneSN    string `json:"drone_sn"`
	UploadUser string `json:"upload_user"`
	S3Key      string `json:"s3_key"`
	CreatedAt  string `json:"created_at"`
}
