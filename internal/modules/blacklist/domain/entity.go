package domain

import "time"

// Blacklist 风控黑名单
type Blacklist struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    *uint      `gorm:"index" json:"user_id,omitempty"`       // 可空（按IP封禁时无用户ID）
	IP        string     `gorm:"size:45;index" json:"ip"`              // 封禁IP，size:45 兼容IPv6
	Reason    string     `gorm:"size:256;not null" json:"reason"`      // 封禁原因
	BlockedBy uint       `gorm:"not null" json:"blocked_by"`           // 操作管理员ID
	BlockedAt time.Time  `json:"blocked_at"`                           // 封禁时间
	ExpiredAt *time.Time `gorm:"index" json:"expired_at,omitempty"`    // 到期时间（空=永久）
	IsActive  bool       `gorm:"default:true;index" json:"is_active"`  // 是否生效
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (b *Blacklist) IsExpired() bool {
	if b.ExpiredAt == nil {
		return false // 永久封禁
	}
	return time.Now().After(*b.ExpiredAt)
}
