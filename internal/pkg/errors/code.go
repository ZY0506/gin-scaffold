package errors

// 系统通用错误码 1xxxx
const (
	Success       = 0
	ErrBadRequest = 10001 // 请求参数错误
	ErrNotFound   = 10002 // 资源不存在
	ErrRateLimit  = 10003 // 请求频率过高
	ErrInternal   = 10004 // 服务器内部错误
	ErrDB         = 10005 // 数据库操作失败
	ErrRedis      = 10006 // 缓存操作失败
	ErrConfig     = 10007 // 配置错误
)

// 认证授权错误码 2xxxx
const (
	ErrUnauthorized     = 20001 // 未登录或登录已过期
	ErrTokenExpired     = 20002 // 令牌已过期
	ErrTokenInvalid     = 20003 // 无效的令牌
	ErrTokenBlacklisted = 20004 // 令牌已被注销
	ErrForbidden        = 20005 // 权限不足
	ErrLoginFailed      = 20006 // 用户名或密码错误
	ErrRefreshExpired   = 20007 // 刷新令牌已过期，请重新登录
	ErrCodeExpired      = 20008 // 验证码已过期
	ErrCodeInvalid      = 20009 // 验证码错误
)

// 用户模块错误码 3xxxx
const (
	ErrUserNotFound       = 30001 // 用户不存在
	ErrUsernameExist      = 30002 // 用户名已存在
	ErrEmailExist         = 30003 // 邮箱已被注册
	ErrUserDisabled       = 30004 // 账号已被禁用
	ErrPwdHashFailed      = 30005 // 密码加密失败
	ErrPwdMismatch        = 30006 // 密码错误
	ErrBlacklisted        = 30007 // 账号被加入黑名单
	ErrFileTypeNotAllowed = 30008 // 文件类型不允许
	ErrFileTooLarge       = 30009 // 文件大小超限
	ErrUploadFailed       = 30010 // 文件保存失败
)

// 管理员模块错误码 4xxxx
const (
	ErrAdminNotFound      = 40001 // 管理员不存在
	ErrAdminUsernameExist = 40002 // 管理员用户名已存在
)
