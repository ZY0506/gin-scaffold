# Gin Scaffold 项目设计文档

> 通用 Go 后端脚手架模板，采用 DDD 模块化架构

## 技术栈

| 领域 | 选型 |
|------|------|
| Web 框架 | gin-gonic/gin |
| ORM | gorm.io/gorm + gorm.io/driver/mysql |
| 缓存/黑名单 | redis/go-redis/v9 |
| 配置 | spf13/viper |
| JWT | golang-jwt/jwt/v5 (双 Token) |
| RBAC | casbin/casbin/v2 + casbin/gorm-adapter/v3 |
| 限流 | golang.org/x/time/rate (令牌桶) |
| 参数校验 | go-playground/validator/v10 |
| 日志 | go.uber.org/zap |
| 密码 | golang.org/x/crypto/bcrypt |
| 错误栈 | pkg/errors |
| 邮件 | net/smtp (SMTP 通用方案) |
| 测试 | testing + stretchr/testify |
| DB | MySQL 8.0+ / Redis 7.2+ |

## 目录结构 (模块化 DDD)

```
gin-scaffold/
├── cmd/
│   └── server/
│       └── main.go                # 程序入口：配置加载 → DI → 启动服务
├── internal/
│   ├── modules/
│   │   ├── auth/                  # 认证模块（登录/注册/Token管理）
│   │   │   ├── domain/           # Token 实体、EmailSender 接口
│   │   │   ├── application/      # 登录/注册/登出/刷新令牌用例
│   │   │   ├── infrastructure/   # JWT 实现、Casbin 初始化、Redis 黑名单、SMTP 发送
│   │   │   └── interfaces/       # Gin Handlers、DTO、路由注册
│   │   ├── user/                  # 用户模块
│   │   │   ├── domain/           # User 实体、UserRepository 接口
│   │   │   ├── application/      # 个人信息修改/密码修改/注销等用例
│   │   │   ├── infrastructure/   # GORM UserRepository 实现
│   │   │   └── interfaces/       # Gin Handlers、DTO、路由注册 (包括 admin)
│   │   └── blacklist/            # 风控黑名单模块
│   │       ├── domain/           # Blacklist 实体、BlacklistRepository 接口
│   │       ├── application/      # 封禁/解封/检查用例
│   │       ├── infrastructure/   # GORM BlacklistRepository 实现
│   │       └── interfaces/       # Gin Handlers、DTO、路由注册 (admin)
│   ├── middleware/
│   │   ├── cors.go               # 跨域中间件
│   │   ├── ratelimit.go          # 令牌桶限流中间件（单机）
│   │   ├── ratelimit_redis.go    # 分布式令牌桶限流中间件（可选）
│   │   ├── auth.go               # JWT 鉴权中间件
│   │   ├── casbin.go             # RBAC 鉴权中间件
│   │   ├── recovery.go           # panic 恢复中间件
│   │   └── logger.go             # zap 请求日志中间件
│   ├── pkg/
│   │   ├── errors/               # 统一错误码（按模块分段）
│   │   ├── response/             # 统一 API 响应格式 {code, msg, data}
│   │   └── validator/            # 自定义验证器（密码强度等）
│   └── router/
│       └── router.go             # 全局路由注册
├── config/
│   ├── config.go                 # YAML 配置结构体
│   ├── config.yaml               # 默认配置文件
│   └── viper.go                  # viper 初始化 + 加载逻辑
├── deployments/
│   ├── Dockerfile                # 多阶段构建
│   └── docker-compose.yml        # app + mysql + redis
├── .env.example                  # 环境变量模板
├── .gitignore
├── go.mod
├── go.sum
└── Makefile
```

## 架构分层与依赖原则

### 模块内三层

```
interfaces/ (Gin Handlers, DTO)
    ↓ 调用
application/ (UseCase 编排)
    ↓ 调用接口（依赖倒置）
domain/ (Entity, Repository 接口) ← infrastructure/ 实现
```

- `domain` 层零外部依赖（不导入 gin/gorm/redis）
- `infrastructure` 实现 domain 定义的接口
- DI 通过手动 New 函数 + main.go 组装

### 请求全链路

```
Client → Recovery → CORS → RateLimit → Logger → Router
                                                      │
                                            ┌─────────┴─────────┐
                                            │       公开路由      │
                                            │  /send-code, /login│
                                            │                    │
                                            │   受保护路由（JWT)   │
                                            │  /auth/*, /admin/* │
                                            │                    │
                                            │  Admin 路由 (+Casbin)│
                                            │  /admin/*          │
                                            └────────────────────┘
```

## 配置设计

### 加载优先级：系统环境变量 > .env > config.yaml

```go
type Config struct {
    App       AppConfig
    Server    ServerConfig       // Port, ReadTimeout, WriteTimeout
    DB        DBConfig           // MySQL DSN
    Redis     RedisConfig        // Redis 地址/密码/DB
    JWT       JWTConfig          // Secret(.env), AccessExpire, RefreshExpire, Issuer
    Casbin    CasbinConfig       // ModelPath
    RateLimit RateLimitConfig    // Rate, Burst
    Email     EmailConfig        // SMTP Host/Port/Username/Password(.env)
    Log       LogConfig          // Level, Filename, MaxSize, etc.
}
```

## 统一 API 响应

```json
{
    "code": 0,
    "msg": "ok",
    "data": { ... }
}
```

```go
Success(data any) Response
Error(code int, msg string) Response
Page(data any, total, page, size int64) Response
```

## 错误码分段

| 范围 | 模块 | 说明 |
|------|------|------|
| 0 | 通用 | 成功 |
| 10001–10099 | 系统通用 | 参数错误、内部错误、限流、DB/Redis 异常 |
| 20001–20099 | 认证授权 | Token 相关、登录失败、权限不足 |
| 30001–30099 | 用户模块 | 用户不存在、账号重复、禁用 |

## 双 Token 认证机制

- **Access Token**: 15 分钟有效，携带 JWT ID (jti)
- **Refresh Token**: 7 天有效，轮换策略（旧 Refresh Token 使用后加入黑名单）
- **Token 黑名单**: Redis Set，以 jti 为 key，TTL 等于 Token 剩余有效期
- **登录流程**: 校验密码 → 风控检查 → 生成双 Token
- **刷新流程**: 校验 Refresh Token → 校验黑名单 → 轮换生成新的双 Token → 旧 Refresh 加入黑名单
- **登出流程**: Access Token 加入黑名单（剩余有效期内不可用）

## 用户表设计

```go
type User struct {
    ID          uint
    Username    string    // 唯一，32位
    Email       string    // 唯一，128位
    Password    string    // bcrypt 哈希，json 隐藏
    Nickname    string    // 昵称，64位
    Avatar      string    // 头像 URL，256位
    Gender      int       // 0:未知 1:男 2:女
    Birthday    *time.Time
    Status      int       // 1:正常 0:禁用
    LastLoginAt *time.Time
    LastLoginIP string    // IPV6 兼容（45位）
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
// 实体与 GORM Model 合一，domain 层直接使用
```

## 风控黑名单

```go
type Blacklist struct {
    ID        uint
    UserID    *uint       // 可空（按IP封禁时为空）
    IP        string      // 封禁 IP
    Reason    string      // 封禁原因
    BlockedBy uint        // 操作管理员 ID
    BlockedAt time.Time
    ExpiredAt *time.Time  // 到期时间（空=永久）
    IsActive  bool        // 是否生效（解封设为 false）
}
```

### 登录风控检查

```
1. 查用户 → 用户存在且状态正常
2. 查 Blacklist WHERE (UserID = ? OR IP = ?) AND IsActive = true AND (ExpiredAt IS NULL OR ExpiredAt > NOW())
3. 命中 → 拒绝登录
4. 未命中 → 继续密码校验
```

## 接口路由

### 客户端 API `/api/v1/`

| 路由 | 方法 | 说明 | 鉴权 |
|------|------|------|------|
| /auth/send-code | POST | 发送邮箱验证码 | 公开 |
| /register | POST | 用户注册（含验证码） | 公开 |
| /login | POST | 登录 | 公开 |
| /auth/refresh | POST | 刷新双 Token | 公开(需Refresh Token) |
| /auth/logout | POST | 登出 | JWT |
| /auth/me | GET | 当前用户信息 | JWT |
| /auth/password | PUT | 修改密码 | JWT |
| /auth/reset-password | POST | 重置密码 | 公开(需验证码) |
| /auth/profile | PUT | 修改个人信息 | JWT |
| /auth/account | DELETE | 注销账户(硬删除) | JWT |

### 管理端 API `/api/v1/admin/`

| 路由 | 方法 | 说明 | 鉴权 |
|------|------|------|------|
| /admin/users | GET | 用户列表(分页) | JWT+Casbin(admin) |
| /admin/users/:id | GET | 用户详情 | JWT+Casbin(admin) |
| /admin/users | POST | 创建用户 | JWT+Casbin(admin) |
| /admin/users/:id | PUT | 更新用户 | JWT+Casbin(admin) |
| /admin/users/:id/status | PATCH | 启用/禁用用户 | JWT+Casbin(admin) |
| /admin/blacklist | POST | 添加黑名单 | JWT+Casbin(admin) |
| /admin/blacklist | GET | 黑名单列表(分页) | JWT+Casbin(admin) |
| /admin/blacklist/:id | DELETE | 解封 | JWT+Casbin(admin) |

## 邮箱验证码

- **验证码**: 6 位数字，5 分钟有效
- **存储**: Redis `verification_code:{email}` → `{code: "123456"}`
- **邮件发送**: SMTP 通用方案（接口 + 默认 SMTPSender 实现）

## 密码强度校验

- 至少 8 位
- 只允许字母和数字
- 必须同时包含至少 1 个字母 + 1 个数字
- 通过自定义 validator 规则实现

## Casbin RBAC 权限模型

```
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)
```

### 预置策略

```csv
p, admin, /api/v1/admin/*, (GET|POST|PUT|PATCH|DELETE)
p, user, /api/v1/auth/*, (GET|POST|PUT|DELETE)
```

## Docker 容器化

- **Dockerfile**: 多阶段构建 (golang:1.22 builder → alpine:3.19 runtime)
- **docker-compose**: app + mysql:8.0 + redis:7.2，含健康检查
- **.env**: 敏感信息通过 env_file 注入

## Makefile 命令

```makefile
run             # 本地热重载运行 (air)
build            # 编译二进制
test             # 运行测试
migrate          # 数据库迁移
docker-up        # docker-compose up -d
docker-down      # docker-compose down
lint             # golangci-lint
```

## 项目初始化流程

1. 克隆/复制模板
2. 复制 `.env.example` → `.env` 填入敏感配置
3. `docker-compose up -d` 启动 MySQL + Redis
4. `make run` 启动 Go 服务（自动执行迁移）
