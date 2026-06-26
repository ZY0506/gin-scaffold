package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// LoadConfig 加载配置，优先级：系统环境变量 > .env > config.yaml
func LoadConfig() (*Config, error) {
	// 1. 先加载 .env 到进程环境变量（使 ${VAR} 占位符和 AutomaticEnv 都能读取到）
	envFile := ".env"
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("加载 .env 文件失败: %w", err)
		}
	}

	// 2. 读取 config.yaml 默认配置
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
		fmt.Println("警告: config.yaml 未找到，将使用零值配置")
	}

	// 3. 系统环境变量覆盖（最高优先级）
	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 4. 解析 yaml 中 ${VAR} 格式的占位符
	for _, key := range v.AllKeys() {
		val := v.GetString(key)
		if len(val) > 3 && val[:2] == "${" && val[len(val)-1:] == "}" {
			envKey := val[2 : len(val)-1]
			envVal, exists := os.LookupEnv(envKey)
			if !exists {
				return nil, fmt.Errorf("环境变量 %s 未设置，请检查 .env 文件", envKey)
			}
			v.Set(key, envVal)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("配置反序列化失败: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置校验失败: %w", err)
	}

	return &cfg, nil
}
