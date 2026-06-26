package application

// 添加黑名单
type AddBlacklistReq struct {
	UserID    *uint  `json:"user_id" binding:"omitempty"` // 封禁用户（按需）
	IP        string `json:"ip" binding:"omitempty,ip"`   // 封禁IP（按需）
	Reason    string `json:"reason" binding:"required,min=1,max=256"`
	ExpiredAt string `json:"expired_at" binding:"omitempty"` // 到期时间，空=永久
}

// 修改黑名单
type UpdateBlacklistReq struct {
	Reason    string `json:"reason" binding:"required,min=1,max=256"`
	ExpiredAt string `json:"expired_at" binding:"omitempty"` // 到期时间，空=永久
}

// 黑名单列表查询
type BlacklistListReq struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// 黑名单列表响应项
type BlacklistItemResp struct {
	ID        uint   `json:"id"`
	UserID    *uint  `json:"user_id,omitempty"`
	IP        string `json:"ip"`
	Reason    string `json:"reason"`
	BlockedBy uint   `json:"blocked_by"`
	BlockedAt string `json:"blocked_at"`
	ExpiredAt string `json:"expired_at,omitempty"`
	IsActive  bool   `json:"is_active"`
}
