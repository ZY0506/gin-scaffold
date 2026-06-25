package infrastructure

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormAdapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"

	"github.com/ZY0506/gin-scaffold/config"
)

// CasbinService Casbin 权限管理服务
// 使用 GORM 适配器持久化策略，支持运行时动态加载策略
type CasbinService struct {
	Enforcer *casbin.SyncedEnforcer
}

// NewCasbinService 创建 Casbin 服务
// 1. 使用 GORM 适配器连接数据库（策略存储在 casbin_rule 表）
// 2. 从文件加载 RBAC 模型
// 3. 返回线程安全的 SyncedEnforcer
func NewCasbinService(cfg *config.CasbinConfig, db *gorm.DB) (*CasbinService, error) {
	// 初始化 GORM 适配器
	adapter, err := gormAdapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}

	// 从文件加载 RBAC 模型
	m, err := model.NewModelFromFile(cfg.ModelPath)
	if err != nil {
		return nil, err
	}

	// 创建线程安全的 SyncedEnforcer
	enforcer, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}

	// 启用自动保存（策略修改后自动持久化到数据库）
	enforcer.EnableAutoSave(true)

	return &CasbinService{Enforcer: enforcer}, nil
}
