package configs

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"db"`
	MQTT   MQTTConfig   `mapstructure:"mqtt"`
	Redis  RedisConfig  `mapstructure:"redis"`
}

func (c *Config) GetDBStr() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		c.DB.Host,
		c.DB.Username,
		c.DB.Password,
		c.DB.DBName,
		c.DB.Port,
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

	// 反序列化配置文件到结构体
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
