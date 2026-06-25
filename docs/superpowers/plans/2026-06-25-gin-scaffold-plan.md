# Gin Scaffold 项目脚手架实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) to implement this plan task-by-task.

**Goal:** 实现一个完整的 Go+Gin+Gorm+MySQL+Redis 后端脚手架模板，包含 JWT 双 Token 认证、Casbin RBAC、API 限流、风控黑名单、邮箱验证码、DDD 模块化架构，支持 Docker 部署。

**Architecture:** 模块化 DDD，每个业务模块（auth/user/blacklist）自含 domain/application/infrastructure/interfaces 四层，手动 New 函数 DI 组装，全局中间件独立于业务模块。

**Tech Stack:** Go 1.22+, Gin, GORM, MySQL 8.0+, Redis 7.2+, go-redis/v9, viper, zap, golang-jwt/v5, casbin/v2, bcrypt, validator/v10, pkg/errors

---

### 任务依赖关系

```
Task 1 (项目骨架)
  └─ Task 2 (共享包)
      └─ Task 3 (User Domain)
      └─ Task 4 (Blacklist Domain)
      └─ Task 5 (JWT + Redis)
      └─ Task 6 (SMTP)
          └─ Task 7 (全局中间件)
          └─ Task 8 (安全中间件)
              └─ Task 9 (Auth 业务层)
              └─ Task 10 (User 业务层)
              └─ Task 11 (Blacklist 业务层)
                  └─ Task 12 (Router + main.go)
                      └─ Task 13 (Docker)
```

---

### Task 1: 项目骨架与配置系统

**Files:**
- Create: `go.mod`
- Create: `config/config.go`
- Create: `config/config.yaml`
- Create: `config/viper.go`
- Create: `.env.example`
- Create: `.gitignore`
- Create: `Makefile`

**内容说明:**

`config/config.go` — 定义所有配置节的结构体，包含 App、Server、DB(MySQL)、Redis、JWT(双Token)、Casbin、RateLimit、Email(SMTP)、Log(zap) 九大节。密码/密钥等敏感字段标记为 `mapstructure:"-"` 以禁止 yaml 读取，转而从 .env 加载。

`config/viper.go` — 实现配置加载函数：先读 config.yaml 填充默认值 → 读 .env 文件覆盖敏感字段 → 系统环境变量最高优先级覆盖。

`config.yaml` — 提供可运行的默认值（MySQL/Redis 地址默认指向 docker-compose 的服务名）。

`.env.example` — 模板文件，列出所有必须的环境变量（DB_PASSWORD, JWT_SECRET, REDIS_PASSWORD, EMAIL_PASSWORD），注释说明其用途。

`Makefile` — run(build 并运行)、build(go build)、test(go test ./...)、lint(golangci-lint)、docker-up/docker-down、migrate 命令。

---

### Task 2: 共享基础设施包 (pkg)

**Files:**
- Create: `internal/pkg/errors/code.go`
- Create: `internal/pkg/errors/error.go`
- Create: `internal/pkg/response/response.go`
- Create: `internal/pkg/validator/validator.go`

**code.go** — 按模块分段定义错误码常量：
```go
// 系统通用 1xxxx
const (
    Success       = 0
    ErrBadRequest = 10001
    ErrNotFound   = 10002
    ErrRateLimit  = 10003
    ErrInternal   = 10004
    ErrDB         = 10005
    ErrRedis      = 10006
    ErrConfig     = 10007
)

// 认证授权 2xxxx
const (
    ErrUnauthorized    = 20001
    ErrTokenExpired    = 20002
    ErrTokenInvalid    = 20003
    ErrTokenBlacklisted = 20004
    ErrForbidden       = 20005
    ErrLoginFailed     = 20006
    ErrRefreshExpired  = 20007
    ErrCodeExpired     = 20008
    ErrCodeInvalid     = 20009
)

// 用户模块 3xxxx
const (
    ErrUserNotFound   = 30001
    ErrUsernameExist  = 30002
    ErrEmailExist     = 30003
    ErrUserDisabled   = 30004
    ErrPwdHashFailed  = 30005
    ErrPwdMismatch    = 30006
    ErrBlacklisted    = 30007
)
```

**error.go** — 自定义 `Error` 类型包装错误码 + 消息 + pkg/errors 堆栈，实现 `WithCode`、`WrapCode` 等工厂函数。

**response.go** — 统一响应结构 `{code int, msg string, data any}`，提供 `Success`, `Error`, `Page` 三个构造函数，以及 `WriteJSON(c, resp)` 写入 gin.Context。

**validator.go** — 自定义 validator 注册（密码强度：`password` tag 校验至少8位+字母+数字组合）。

---

### Task 3: 用户模块 Domain + Infrastructure

**Files:**
- Create: `internal/modules/user/domain/entity.go`
- Create: `internal/modules/user/domain/repository.go`
- Create: `internal/modules/user/infrastructure/model.go`
- Create: `internal/modules/user/infrastructure/repo.go`

**entity.go** — User 结构体（务实合一风格，带 gorm/json tag），字段：ID, Username, Email, Password(json:"-"), Nickname, Avatar, Gender(int 0/1/2), Birthday(*time.Time), Status(int 1/0), LastLoginAt(*time.Time), LastLoginIP(string), CreatedAt, UpdatedAt。

**repository.go** — UserRepository 接口：Create, FindByID, FindByUsername, FindByEmail, Update, Delete, List(page, size, conditions) ([]User, int64 total, error), UpdateLoginInfo(id, ip)。

**model.go** — 实体与 GORM Model 合一，不需要额外结构体转换。

**repo.go** — GormUserRepo 结构体实现 UserRepository 接口。

---

### Task 4: 风控黑名单模块 Domain + Infrastructure

**Files:**
- Create: `internal/modules/blacklist/domain/entity.go`
- Create: `internal/modules/blacklist/domain/repository.go`
- Create: `internal/modules/blacklist/infrastructure/repo.go`

**entity.go** — Blacklist 结构体：ID, UserID(*uint), IP, Reason, BlockedBy(uint), BlockedAt, ExpiredAt(*time.Time), IsActive(bool)，带 gorm tag。

**repository.go** — BlacklistRepository 接口：Create, FindByUserID, FindByIP, FindActive(userID, ip) (*Blacklist, error), List(page, size) ([]Blacklist, int64, error), Deactivate(id) error。

**repo.go** — GormBlacklistRepo 实现。

---

### Task 5: JWT 双 Token + Redis 黑名单

**Files:**
- Create: `internal/modules/auth/infrastructure/jwt.go`
- Create: `internal/modules/auth/infrastructure/blacklist.go`

**jwt.go** — JWTService 结构体：
- `GeneratePair(userID uint, role string) (accessToken, refreshToken string, err error)` — 生成双 Token
- `ValidateToken(tokenString string) (*Claims, error)` — 验证 Token 并解析 Claims
- Claims 包含：UserID, Role, TokenType(access/refresh), JTI(UUID), 标准注册项

Access Token 有效期 15 分钟，Refresh Token 有效期 7 天。

**blacklist.go** — RedisTokenBlacklist 结构体：
- `Add(ctx, jti string, ttl time.Duration) error` — 将 jti 加入 Redis Set
- `Exists(ctx, jti string) (bool, error)` — 检查是否在黑名单中

---

### Task 6: 邮箱验证码基础设施

**Files:**
- Create: `internal/modules/auth/domain/email.go`
- Create: `internal/modules/auth/infrastructure/smtp.go`

**email.go** — EmailSender 接口 + VerificationCode 接口：
- `EmailSender`: Send(to, subject, body string) error
- `CodeStore`: Set(ctx, email, code string, ttl time.Duration) error; Get(ctx, email string) (string, error); Del(ctx, email string) error

**smtp.go** — SMTPSender 实现 EmailSender（net/smtp），RedisCodeStore 实现 CodeStore（go-redis）。

---

### Task 7: 全局中间件

**Files:**
- Create: `internal/middleware/recovery.go`
- Create: `internal/middleware/cors.go`
- Create: `internal/middleware/logger.go`
- Create: `internal/middleware/ratelimit.go`

**recovery.go** — 自定义 Recovery 中间件，用 zap 记录 panic 堆栈，返回统一错误响应。

**cors.go** — 跨域中间件，允许配置允许的来源、方法、头部。

**logger.go** — 请求日志中间件，zap 记录 method、path、status、latency、client IP。

**ratelimit.go** — 基于 `golang.org/x/time/rate` 的令牌桶限流中间件，支持配置速率和突发大小，超出返回 10003 错误码。支持可选的 Redis 分布式版。

---

### Task 8: 安全中间件

**Files:**
- Create: `internal/middleware/auth.go`
- Create: `internal/middleware/casbin.go`
- Create: `internal/modules/auth/infrastructure/casbin.go`

**auth.go** — JWT 鉴权中间件：从 Authorization Header(Bearer token) 提取 Token，调用 JWTService.ValidateToken，解析 Claims 注入 gin.Context（Set "user_id"/"role"），黑名单效验。

**casbin.go** — Casbin 鉴权中间件：从 gin.Context 取 user role 和请求路径/方法，调用 casbin.Enforce。

**casbin.go** (infrastructure) — CasbinService 结构体：初始化 Casbin Enforcer（GORM adapter），加载 model.conf（嵌入的资源配置）。提供预置角色策略：admin 和 user。

---

### Task 9: Auth 模块业务层

**Files:**
- Create: `internal/modules/auth/application/dto.go`
- Create: `internal/modules/auth/application/service.go`
- Create: `internal/modules/auth/interfaces/handler.go`
- Create: `internal/modules/auth/interfaces/router.go`

**dto.go** — 请求/响应 DTO：
- SendCodeReq{Email}
- RegisterReq{Email, Code, Username, Password}
- LoginReq{Account(用户名或邮箱), Password}
- RefreshTokenReq{RefreshToken}
- AuthTokenResp{AccessToken, RefreshToken, ExpiresIn}
- MeResp(User 信息，不暴露 Password)

**service.go** — AuthService 编排登录/注册/刷新/登出/发送验证码/重置密码用例，注入 UserRepository + BlacklistRepository + JWTService + CodeStore + EmailSender。

**handler.go** — Gin Handlers，各方法调用 AuthService，通过 response 包返回。

**router.go** — RegisterAuthRoutes 函数，路由分组注册（公开路由 + 需 JWT 的受保护路由）。

---

### Task 10: User 模块业务层

**Files:**
- Create: `internal/modules/user/application/dto.go`
- Create: `internal/modules/user/application/service.go`
- Create: `internal/modules/user/interfaces/handler.go`
- Create: `internal/modules/user/interfaces/router.go`

**dto.go** — UpdateProfileReq{Nickname, Avatar, Gender, Birthday}, ChangePasswordReq{OldPwd, NewPwd}, AdminCreateUserReq, AdminUpdateUserReq, UserPageResp。

**service.go** — UserService：
- 客户端：GetProfile, UpdateProfile, ChangePassword（校验旧密码）
- 管理端：List, GetByID, CreateByAdmin, UpdateByAdmin, ToggleStatus

**handler.go / router.go** — 客户端路由（/auth/* 下）和管理端路由（/admin/users）。

---

### Task 11: Blacklist 模块业务层

**Files:**
- Create: `internal/modules/blacklist/application/dto.go`
- Create: `internal/modules/blacklist/application/service.go`
- Create: `internal/modules/blacklist/interfaces/handler.go`
- Create: `internal/modules/blacklist/interfaces/router.go`

**service.go** — BlacklistService：Add(userID or IP), List, Deactivate 三个用例。

**handler.go / router.go** — 管理端路由 /admin/blacklist。

---

### Task 12: 路由注册与程序入口

**Files:**
- Create: `internal/router/router.go`
- Create: `cmd/server/main.go`

**router.go** — 全局路由初始化函数：注册全局中间件 → 挂载各模块路由 → 健康检查端点 /health。

**main.go** — 入口函数：
1. viper 加载配置
2. zap 初始化日志
3. GORM 连接 MySQL + AutoMigrate
4. go-redis 连接 Redis
5. 初始化各模块基础设施（Repo/Service）
6. Casbin 初始化
7. 手动 DI 组装各模块应用服务
8. 注册路由
9. gin.Start 启动服务

---

### Task 13: Docker 容器化

**Files:**
- Create: `deployments/Dockerfile`
- Create: `deployments/docker-compose.yml`

**Dockerfile** — 多阶段构建：golang:1.22-alpine 构建 → alpine:3.19 运行。

**docker-compose.yml** — 三个服务：
- app：构建当前目录，端口 8080:8080，env_file .env，depends_on mysql(healthcheck) + redis
- mysql：8.0，挂载持久化卷，初始化数据库
- redis：7.2-alpine，挂载持久化卷
