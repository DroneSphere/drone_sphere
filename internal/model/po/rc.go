package po

type RTRC struct {
	SN           string `redis:"sn" json:"sn"`
	OnlineStatus bool   `redis:"online_status" json:"online_status"`
}
