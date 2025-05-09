package entity

import (
	"github.com/dronesphere/internal/model/po"
)

type Wayline struct {
	po.Wayline
	Url string `json:"url"`
}
