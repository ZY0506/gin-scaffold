package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/ZY0506/gin-scaffold/config"
	"github.com/ZY0506/gin-scaffold/internal/middleware"
	adminApp "github.com/ZY0506/gin-scaffold/internal/modules/admin/application"
	adminDomain "github.com/ZY0506/gin-scaffold/internal/modules/admin/domain"
	adminInfra "github.com/ZY0506/gin-scaffold/internal/modules/admin/infrastructure"
	adminHandler "github.com/ZY0506/gin-scaffold/internal/modules/admin/interfaces"
	authApp "github.com/ZY0506/gin-scaffold/internal/modules/auth/application"
	authInfra "github.com/ZY0506/gin-scaffold/internal/modules/auth/infrastructure"
	authHandler "github.com/ZY0506/gin-scaffold/internal/modules/auth/interfaces"
	blApp "github.com/ZY0506/gin-scaffold/internal/modules/blacklist/application"
	blInfra "github.com/ZY0506/gin-scaffold/internal/modules/blacklist/infrastructure"
	blHandler "github.com/ZY0506/gin-scaffold/internal/modules/blacklist/interfaces"
	userApp "github.com/ZY0506/gin-scaffold/internal/modules/user/application"
	userInfra "github.com/ZY0506/gin-scaffold/internal/modules/user/infrastructure"
	userHandler "github.com/ZY0506/gin-scaffold/internal/modules/user/interfaces"
	"github.com/ZY0506/gin-scaffold/internal/router"
)

func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 初始化 zap 日志
	zapLogger, err := initLogger(&cfg.Log)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer func() { _ = zapLogger.Sync() }()

	// 3. 连接 MySQL
	db, err := initDB(cfg)
	if err != nil {
		zapLogger.Fatal("连接数据库失败", zap.Error(err))
	}

	// 4. 连接 Redis
	rdb := initRedis(cfg)

	// 5. 创建 Repo 层实例
	userRepo := userInfra.NewGormUserRepo(db)
	blRepo := blInfra.NewGormBlacklistRepo(db)
	adminRepo := adminInfra.NewGormAdminRepo(db)
	opLogRepo := adminInfra.NewGormOperationLogRepo(db)

	// 6. 创建基础设施服务
	jwtSvc := authInfra.NewJWTService(&cfg.JWT)
	tokenBlacklist := authInfra.NewRedisTokenBlacklist(rdb)
	codeStore := authInfra.NewRedisCodeStore(rdb)
	emailSender := authInfra.NewSMTPSender(&cfg.Email)

	// 7. 创建 Casbin 服务
	casbinSvc, err := authInfra.NewCasbinService(&cfg.Casbin, db)
	if err != nil {
		zapLogger.Fatal("初始化 Casbin 失败", zap.Error(err))
	}

	// 8. 创建应用服务
	authSvc := authApp.NewAuthService(cfg, userRepo, blRepo, jwtSvc, tokenBlacklist, codeStore, emailSender)
	userSvc := userApp.NewUserService(userRepo)
	blSvc := blApp.NewBlacklistService(blRepo)
	adminSvc := adminApp.NewAdminService(adminRepo)
	opLogSvc := adminApp.NewOperationLogService(opLogRepo)

	// 9. 创建 Handler
	authH := authHandler.NewAuthHandler(authSvc)
	userH := userHandler.NewUserHandler(userSvc, cfg.Upload.SaveDir, int64(cfg.Upload.MaxSizeMB)*1024*1024, cfg.Upload.AllowedExts)
	blH := blHandler.NewBlacklistHandler(blSvc)
	adminH := adminHandler.NewAdminHandler(adminSvc, opLogSvc, jwtSvc)

	// 10. 创建中间件
	authMW := middleware.JWTAuth(jwtSvc, tokenBlacklist)
	casbinMW := middleware.CasbinRBAC(casbinSvc.Enforcer)

	// 操作日志中间件 — 记录管理端写操作
	opLogMW := middleware.WithOperationLog(func(c *gin.Context) {
		entry := &adminDomain.OperationLog{
			AdminID:    middleware.GetUserID(c),
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: c.Writer.Status(),
			ClientIP:   c.ClientIP(),
		}
		go func(l *adminDomain.OperationLog) {
			if err := opLogSvc.Create(context.Background(), l); err != nil {
				zapLogger.Warn("写入操作日志失败", zap.Error(err))
			}
		}(entry)
	})

	// 11. 初始化 Gin 引擎 + 全局中间件
	ginEngine := gin.New()
	ginEngine.Use(
		middleware.Recovery(zapLogger),
		middleware.CORS(),
		middleware.Logger(zapLogger),
		middleware.RateLimit(&cfg.RateLimit),
	)

	// 12. 创建管理端路由结构体
	userAdminRouter := userHandler.NewAdminRouter(userH)
	blAdminRouter := blHandler.NewAdminRouter(blH)
	adminRouter := adminHandler.NewAdminRouter(adminH)

	// 13. 注册路由
	router.Register(ginEngine, authH, authMW, userH, userAdminRouter, blAdminRouter,
		adminRouter, adminH, authMW, casbinMW, opLogMW)

	// 14. 静态文件服务
	ginEngine.Static("/uploads", "./storage")

	// 15. 种子管理员（首次启动时创建默认管理员账号）
	seedDefaultAdmin(db, zapLogger)

	// 16. 启动服务
	zapLogger.Info("服务启动", zap.String("addr", cfg.Server.Addr()))
	if err := ginEngine.Run(cfg.Server.Addr()); err != nil {
		zapLogger.Fatal("服务启动失败", zap.Error(err))
	}
}

func initLogger(cfg *config.LogConfig) (*zap.Logger, error) {
	var zapCfg zap.Config
	if cfg.Level == "debug" {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}
	return zapCfg.Build()
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DB.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdle)
	return db, nil
}

func initRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
}

// seedDefaultAdmin 首次启动时创建默认管理员
func seedDefaultAdmin(db *gorm.DB, logger *zap.Logger) {
	var count int64
	db.Model(&adminDomain.Admin{}).Count(&count)
	if count > 0 {
		return
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		logger.Warn("管理员密码加密失败", zap.Error(err))
		return
	}

	admin := &adminDomain.Admin{
		Username: "admin",
		Password: string(hashedPwd),
		Nickname: "超级管理员",
		Status:   1,
	}

	if err := db.Create(admin).Error; err != nil {
		logger.Warn("创建默认管理员失败", zap.Error(err))
		return
	}

	logger.Info("默认管理员已创建",
		zap.String("username", "admin"),
		zap.String("password", "admin123"),
	)
}
