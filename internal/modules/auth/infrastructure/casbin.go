package infrastructure

import (
	"os"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	gormAdapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"

	"github.com/ZY0506/gin-scaffold/config"
)

// CasbinService Casbin 权限管理服务
type CasbinService struct {
	Enforcer *casbin.SyncedEnforcer
}

// NewCasbinService 创建 Casbin 服务，并自动初始化策略
func NewCasbinService(cfg *config.CasbinConfig, db *gorm.DB) (*CasbinService, error) {
	// 初始化 GORM 适配器（策略存储在 casbin_rule 表）
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

	enforcer.EnableAutoSave(true)

	// 首次启动时，如果数据库中没有策略，从 policy.csv 导入初始策略
	if err := seedPolicies(enforcer, cfg.PolicyPath); err != nil {
		return nil, err
	}

	return &CasbinService{Enforcer: enforcer}, nil
}

// seedPolicies 检查数据库中是否有策略，为空则从 CSV 文件加载初始策略
func seedPolicies(enforcer *casbin.SyncedEnforcer, policyPath string) error {
	// 检查 policy.csv 是否存在
	if _, err := os.Stat(policyPath); os.IsNotExist(err) {
		return nil
	}

	// 检查是否已有策略（已有则跳过初始化）
	policies, err := enforcer.GetPolicy()
	if err != nil {
		return err
	}
	if len(policies) > 0 {
		return nil
	}

	// 用 file-adapter 从 CSV 加载策略到模型
	fa := fileadapter.NewAdapter(policyPath)
	if err := fa.LoadPolicy(enforcer.GetModel()); err != nil {
		return err
	}

	// 将加载的策略保存到数据库
	return enforcer.SavePolicy()
}
