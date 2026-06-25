package application

// 发送验证码
type SendCodeReq struct {
	Email string `json:"email" binding:"required,email"`
}

// 注册
type RegisterReq struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required,len=6"`
	Username string `json:"username" binding:"required,min=4,max=32"`
	Password string `json:"password" binding:"required,password"`
}

// 登录
type LoginReq struct {
	Account  string `json:"account" binding:"required"` // 用户名或邮箱
	Password string `json:"password" binding:"required"`
}

// 刷新 Token
type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Token 响应
type AuthTokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// 修改密码
type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,password"`
}

// 重置密码
type ResetPasswordReq struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,password"`
}

// 修改个人信息
type UpdateProfileReq struct {
	Nickname string `json:"nickname" binding:"omitempty,max=64"`
	Avatar   string `json:"avatar" binding:"omitempty,max=256"`
	Gender   *int   `json:"gender" binding:"omitempty,oneof=0 1 2"`
	Birthday string `json:"birthday" binding:"omitempty"` // 格式: 2006-01-02
}

// 用户信息响应
type UserInfoResp struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Gender      int    `json:"gender"`
	Birthday    string `json:"birthday,omitempty"`
	Status      int    `json:"status"`
	LastLoginAt string `json:"last_login_at,omitempty"`
	LastLoginIP string `json:"last_login_ip,omitempty"`
	CreatedAt   string `json:"created_at"`
}
