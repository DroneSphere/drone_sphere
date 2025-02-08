package app

import (
	"fmt"
	"github.com/dronesphere/internal/adapter/eventhandler"
	v1 "github.com/dronesphere/internal/adapter/http/v1"
	"github.com/dronesphere/internal/service"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2"
	"github.com/lmittmann/tint"
	"gorm.io/driver/postgres"
	"log/slog"
	"os"
	"time"

	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/configs"
	"github.com/dronesphere/internal/repo"
	slogGorm "github.com/orandin/slog-gorm"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func Run(cfg *configs.Config) {
	// Log
	w := os.Stdout
	logger := slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	)

	// EventBus
	eb := EventBus.New()

	// Postgres
	gormLogger := slogGorm.New(
		slogGorm.WithHandler(logger.Handler()),                        // since v1.3.0
		slogGorm.WithTraceAll(),                                       // trace all messages
		slogGorm.SetLogLevel(slogGorm.DefaultLogType, slog.Level(32)), // Define the default logging level
	)
	dsn := "host=47.245.40.222 user=admin password=thF@AHgy3SUR dbname=drone_sphere port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		panic("failed to connect database")
	}

	// MQTT
	var broker = "47.245.40.222"
	var port = 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername("server")
	opts.SetPassword("server")
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		logger.Info("Received message", slog.Any("topic", msg.Topic()), slog.Any("message", string(msg.Payload())))
	})
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		logger.Info("Connected to MQTT broker")
	})
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logger.Error("Connection lost", slog.Any("err", err.Error()))
	})
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// TODO: Prepare Redis
	opt, err := redis.ParseURL("redis://:thF@AHgy3SUR@47.245.40.222:6379")
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
	rds := redis.NewClient(opt)
	logger.Info("Redis connected")

	// Repos
	userRepo := repo.NewUserGormRepo(db, logger)
	droneRepo := repo.NewDroneGormRepo(db, rds, logger)

	// Services
	userSvc := service.NewUserSvc(userRepo, logger)
	droneSvc := service.NewDroneImpl(droneRepo, logger, client)

	// Event Handlers
	eventhandler.NewHandler(eb, logger, client, droneSvc)

	// Servers
	http := fiber.New()
	v1.NewRouter(http, eb, logger, userSvc, droneSvc)

	// Run
	err = http.Listen(":10086")
	if err != nil {
		logger.Error(err.Error())
	}
}
