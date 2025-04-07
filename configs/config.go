package configs

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	DB       DBConfig       `mapstructure:"db"`
	MQTT     MQTTConfig     `mapstructure:"mqtt"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Platform PlatformConfig `mapstructure:"platform"` // 新增平台配置
}

func (c *Config) GetDBStr() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		c.DB.Username,
		c.DB.Password,
		c.DB.Host,
		c.DB.Port,
		c.DB.DBName,
	)
}

func (c *Config) GetRedisStr() string {
	return fmt.Sprintf(
		"redis://:%s@%s:%s/%d",
		c.Redis.Password,
		c.Redis.Host,
		c.Redis.Port,
		c.Redis.DB,
	)
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type MQTTConfig struct {
	Broker   string `mapstructure:"broker"`
	Port     string `mapstructure:"port"`
	ClientID string `mapstructure:"client_id"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// PlatformConfig 平台配置
type PlatformConfig struct {
	Name        string `mapstructure:"name"`         // 平台名称
	Workspace   string `mapstructure:"workspace"`    // 工作空间名称
	WorkspaceID string `mapstructure:"workspace_id"` // 工作空间ID
	Desc        string `mapstructure:"desc"`         // 平台描述
	Thing       struct {
		Host     string `mapstructure:"host"`     // 物联网平台连接地址
		Username string `mapstructure:"username"` // 物联网平台用户名
		Password string `mapstructure:"password"` // 物联网平台密码
	} `mapstructure:"thing"`
	API struct {
		Host  string `mapstructure:"host"`  // API服务地址
		Token string `mapstructure:"token"` // API访问令牌
	} `mapstructure:"api"`
	WS struct {
		Host  string `mapstructure:"host"`  // WebSocket服务地址
		Token string `mapstructure:"token"` // WebSocket访问令牌
	} `mapstructure:"ws"`
}

func LoadConfig() (*Config, error) {
	// Development 环境下加载.env文件，默认为development
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	if os.Getenv("APP_ENV") == "development" {
		err := godotenv.Load(".env")
		if err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	// 设置环境变量
	_ = viper.BindEnv("server.port", "SERVER_PORT")
	_ = viper.BindEnv("db.host", "DB_HOST")
	_ = viper.BindEnv("db.port", "DB_PORT")
	_ = viper.BindEnv("db.username", "DB_USER")
	_ = viper.BindEnv("db.password", "DB_PASSWORD")
	_ = viper.BindEnv("db.dbname", "DB_NAME")
	_ = viper.BindEnv("mqtt.broker", "MQTT_BROKER")
	_ = viper.BindEnv("mqtt.port", "MQTT_PORT")
	_ = viper.BindEnv("mqtt.client_id", "MQTT_CLIENT_ID")
	_ = viper.BindEnv("mqtt.username", "MQTT_USERNAME")
	_ = viper.BindEnv("mqtt.password", "MQTT_PASSWORD")
	_ = viper.BindEnv("redis.host", "REDIS_HOST")
	_ = viper.BindEnv("redis.port", "REDIS_PORT")
	_ = viper.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = viper.BindEnv("redis.db", "REDIS_DB")

	// 平台相关环境变量
	_ = viper.BindEnv("platform.name", "PLATFORM_NAME")
	_ = viper.BindEnv("platform.workspace", "PLATFORM_WORKSPACE")
	_ = viper.BindEnv("platform.workspace_id", "PLATFORM_WORKSPACE_ID")
	_ = viper.BindEnv("platform.desc", "PLATFORM_DESC")
	_ = viper.BindEnv("platform.thing.host", "PLATFORM_THING_HOST")
	_ = viper.BindEnv("platform.thing.username", "PLATFORM_THING_USERNAME")
	_ = viper.BindEnv("platform.thing.password", "PLATFORM_THING_PASSWORD")
	_ = viper.BindEnv("platform.api.host", "PLATFORM_API_HOST")
	_ = viper.BindEnv("platform.api.token", "PLATFORM_API_TOKEN")
	_ = viper.BindEnv("platform.ws.host", "PLATFORM_WS_HOST")
	_ = viper.BindEnv("platform.ws.token", "PLATFORM_WS_TOKEN")

	// 反序列化配置文件到结构体
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
