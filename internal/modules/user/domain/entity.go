package domain

import "time"

// 性别枚举
const (
	GenderUnknown = 0
	GenderMale    = 1
	GenderFemale  = 2
)

// 状态枚举
const (
	StatusActive   = 1 // 正常
	StatusDisabled = 0 // 禁用
)

// User 用户实体（务实合一，同时作为 GORM Model 和 Domain Entity）
// Password 字段使用 json:"-" 确保不会通过 API 响应泄露
type User struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Username    string     `gorm:"uniqueIndex;size:32;not null" json:"username"`
	Email       string     `gorm:"uniqueIndex;size:128;not null" json:"email"`
	Password    string     `gorm:"size:256;not null" json:"-"`
	Nickname    string     `gorm:"size:64" json:"nickname"`
	Avatar      string     `gorm:"size:256" json:"avatar"`
	Gender      int        `gorm:"default:0" json:"gender"`
	Birthday    *time.Time `json:"birthday,omitempty"`
	Status      int        `gorm:"default:1" json:"status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP string     `gorm:"size:45" json:"last_login_ip,omitempty"` // size:45 兼容 IPv6 最大长度
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (u *User) IsDisabled() bool {
	return u.Status == StatusDisabled
}
