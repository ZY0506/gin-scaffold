package config

import (
	"fmt"
	"time"
)

type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Server    ServerConfig    `mapstructure:"server"`
	DB        DBConfig        `mapstructure:"db"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Casbin    CasbinConfig    `mapstructure:"casbin"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Email     EmailConfig     `mapstructure:"email"`
	Log       LogConfig       `mapstructure:"log"`
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("server.port 必须大于 0")
	}
	if c.DB.Host == "" || c.DB.Port <= 0 {
		return fmt.Errorf("db.host 和 db.port 必须配置")
	}
	if c.Redis.Host == "" || c.Redis.Port <= 0 {
		return fmt.Errorf("redis.host 和 redis.port 必须配置")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret 未配置，请在 .env 中设置 JWT_SECRET")
	}
	if c.JWT.AccessExpire <= 0 || c.JWT.RefreshExpire <= 0 {
		return fmt.Errorf("jwt.access_expire 和 refresh_expire 必须大于 0")
	}
	return nil
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Env     string `mapstructure:"env"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

func (s ServerConfig) Addr() string {
	return fmt.Sprintf(":%d", s.Port)
}

type DBConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	Charset  string `mapstructure:"charset"`
	MaxOpen  int    `mapstructure:"max_open"`
	MaxIdle  int    `mapstructure:"max_idle"`
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.Charset)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	AccessExpire  time.Duration `mapstructure:"access_expire"`
	RefreshExpire time.Duration `mapstructure:"refresh_expire"`
	Issuer        string        `mapstructure:"issuer"`
}

type CasbinConfig struct {
	ModelPath  string `mapstructure:"model_path"`
	PolicyPath string `mapstructure:"policy_path"`
}

type RateLimitConfig struct {
	Rate  float64 `mapstructure:"rate"`
	Burst int     `mapstructure:"burst"`
}

type EmailConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	FromName string `mapstructure:"from_name"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}
