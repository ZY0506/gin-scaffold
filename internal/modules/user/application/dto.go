package application

// 管理端创建用户
type AdminCreateUserReq struct {
	Username string `json:"username" binding:"required,min=4,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,password"`
	Nickname string `json:"nickname" binding:"omitempty,max=64"`
	Avatar   string `json:"avatar" binding:"omitempty,max=256"`
	Gender   int    `json:"gender" binding:"omitempty,oneof=0 1 2"`
	Status   *int   `json:"status" binding:"omitempty,oneof=0 1"`
}

// 管理端更新用户
type AdminUpdateUserReq struct {
	Nickname string `json:"nickname" binding:"omitempty,max=64"`
	Avatar   string `json:"avatar" binding:"omitempty,max=256"`
	Gender   *int   `json:"gender" binding:"omitempty,oneof=0 1 2"`
	Status   *int   `json:"status" binding:"omitempty,oneof=0 1"`
}

// 用户列表查询
type UserListReq struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" binding:"omitempty"`
	Status   *int   `form:"status" binding:"omitempty,oneof=0 1"`
}

// 用户列表响应
type UserItemResp struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Gender      int    `json:"gender"`
	Status      int    `json:"status"`
	LastLoginAt string `json:"last_login_at,omitempty"`
	LastLoginIP string `json:"last_login_ip,omitempty"`
	CreatedAt   string `json:"created_at"`
}
