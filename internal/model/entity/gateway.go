package entity

import (
	"time"

	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/pkg/misc"
	"github.com/jinzhu/copier"
)

// Gateway 网关设备实体
type Gateway struct {
	misc.BaseModel
	SN             string          `json:"sn"`               // 序列号
	Callsign       string          `json:"callsign"`         // 呼号
	Description    string          `json:"description"`      // 描述
	GatewayModelID uint            `json:"gateway_model_id"` // 网关型号ID
	GatewayModel   po.GatewayModel `json:"gateway_model"`    // 网关型号
	UserID         uint            `json:"user_id"`          // 用户ID
	Status         int             `json:"status"`           // 在线状态
	LastOnlineAt   time.Time       `json:"last_online_at"`   // 最后在线时间
}

// NewGatewayFromPO 从持久化对象创建网关实体
func NewGatewayFromPO(p *po.Gateway) *Gateway {
	if p == nil {
		return nil
	}

	var e Gateway
	if err := copier.Copy(&e, p); err != nil {
		return nil
	}

	e.ID = p.ID
	e.CreatedAt = p.CreatedTime
	e.UpdatedAt = p.UpdatedTime
	return &e
}

// ToPO 转换为持久化对象
func (g *Gateway) ToPO() *po.Gateway {
	var p po.Gateway
	if err := copier.Copy(&p, g); err != nil {
		return nil
	}
	p.CreatedTime = g.CreatedAt
	p.UpdatedTime = g.UpdatedAt
	return &p
}

// StatusText 获取状态文本描述
func (g *Gateway) StatusText() string {
	if g.Status == 1 {
		return "在线"
	}
	return "离线"
}
