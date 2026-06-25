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

// 重置密码
type ResetPasswordReq struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,password"`
}
