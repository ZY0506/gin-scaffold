package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func LoadConfig() (*Config, error) {
	v := viper.New()

	// 1. 默认配置文件
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	// 读取 config.yaml
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
		// config.yaml 不存在可以接受，会用默认值
	}

	// 2. 环境变量覆盖
	v.SetEnvPrefix("")            // 无前缀
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 手动从环境变量读取 yaml 中 ${VAR} 格式的占位符并覆盖
	for _, key := range v.AllKeys() {
		val := v.GetString(key)
		if len(val) > 3 && val[:2] == "${" && val[len(val)-1:] == "}" {
			envKey := val[2 : len(val)-1]
			if envVal := os.Getenv(envKey); envVal != "" {
				v.Set(key, envVal)
			}
		}
	}

	// 3. 读 .env 文件
	envFile := ".env"
	if _, err := os.Stat(envFile); err == nil {
		v.SetConfigFile(envFile)
		v.SetConfigType("env")
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("读取 .env 文件失败: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("配置反序列化失败: %w", err)
	}

	return &cfg, nil
}
