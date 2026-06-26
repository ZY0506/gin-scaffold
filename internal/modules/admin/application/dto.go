package application

// === 管理员 DTO ===

// AdminLoginReq 管理员登录
type AdminLoginReq struct {
	Username string `json:"username" binding:"required,min=4,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// AdminCreateReq 创建管理员
type AdminCreateReq struct {
	Username string `json:"username" binding:"required,min=4,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Nickname string `json:"nickname" binding:"omitempty,max=64"`
}

// AdminUpdateReq 修改管理员
type AdminUpdateReq struct {
	Nickname string `json:"nickname" binding:"omitempty,max=64"`
	Password string `json:"password" binding:"omitempty,min=6,max=64"`
	Status   *int   `json:"status" binding:"omitempty,oneof=0 1"`
}

// AdminItemResp 管理员列表响应
type AdminItemResp struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Status      int    `json:"status"`
	LastLoginAt string `json:"last_login_at,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// === 操作日志 DTO ===

// OperationLogListReq 操作日志列表查询
type OperationLogListReq struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}
