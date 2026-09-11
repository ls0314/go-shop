package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 全局配置结构体
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	ES       ESConfig       `yaml:"elasticsearch"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type RabbitMQConfig struct {
	DSN string `yaml:"dsn"`
}

type ESConfig struct {
	Addresses []string `yaml:"addresses"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `yaml:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Dbname   string `yaml:"dbname"`
	Sslmode  string `yaml:"sslmode"`
}

// GlobalConfig 全局配置实例
var GlobalConfig Config

// LoadConfig 加载配置文件
func LoadConfig(configPath string) error {
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析 YAML 配置
	err = yaml.Unmarshal(data, &GlobalConfig)
	if err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	applyEnvOverrides()
	return nil
}

// GetDSN 获取数据库连接字符串
func (db *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.Dbname, db.Sslmode)
}

func applyEnvOverrides() {
	GlobalConfig.Server.Port = getEnv("DEMO_SHOP_SERVER_PORT", GlobalConfig.Server.Port)

	GlobalConfig.Database.Host = getEnv("DEMO_SHOP_DB_HOST", GlobalConfig.Database.Host)
	GlobalConfig.Database.Port = getEnv("DEMO_SHOP_DB_PORT", GlobalConfig.Database.Port)
	GlobalConfig.Database.User = getEnv("DEMO_SHOP_DB_USER", GlobalConfig.Database.User)
	GlobalConfig.Database.Password = getEnv("DEMO_SHOP_DB_PASSWORD", GlobalConfig.Database.Password)
	GlobalConfig.Database.Dbname = getEnv("DEMO_SHOP_DB_NAME", GlobalConfig.Database.Dbname)
	GlobalConfig.Database.Sslmode = getEnv("DEMO_SHOP_DB_SSLMODE", GlobalConfig.Database.Sslmode)

	GlobalConfig.Redis.Addr = getEnv("DEMO_SHOP_REDIS_ADDR", GlobalConfig.Redis.Addr)
	GlobalConfig.Redis.Password = getEnv("DEMO_SHOP_REDIS_PASSWORD", GlobalConfig.Redis.Password)

	GlobalConfig.RabbitMQ.DSN = getEnv("DEMO_SHOP_RABBITMQ_DSN", GlobalConfig.RabbitMQ.DSN)

	// ES 地址是列表,env 用逗号分隔:DEMO_SHOP_ES_ADDRESSES=http://es:9200,http://es2:9200
	if es := os.Getenv("DEMO_SHOP_ES_ADDRESSES"); es != "" {
		GlobalConfig.ES.Addresses = strings.Split(es, ",")
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
