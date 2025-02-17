package app

import (
	"context"
	"fmt"
	"github.com/dronesphere/internal/adapter/eventhandler"
	"github.com/dronesphere/internal/adapter/http/dji"
	v1 "github.com/dronesphere/internal/adapter/http/v1"
	"github.com/dronesphere/internal/adapter/ws"
	"github.com/dronesphere/internal/service"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2"
	"github.com/lmittmann/tint"
	"gorm.io/driver/postgres"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	// Redis
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
	saRepo := repo.NewSearchAreaGormRepo(db, rds, logger)
	algoRepo := repo.NewDetectAlgoGormRepo(db, logger)

	// Services
	userSvc := service.NewUserSvc(userRepo, logger)
	droneSvc := service.NewDroneImpl(droneRepo, logger, client)
	saSvc := service.NewSearchAreaImpl(saRepo, logger, client)
	algoSvc := service.NewDetectAlgoImpl(algoRepo, logger)

	// Event Handlers
	eventhandler.NewHandler(eb, logger, client, droneSvc)

	// Servers
	httpV1 := fiber.New()
	v1.NewRouter(httpV1, eb, logger, userSvc, droneSvc, saSvc, algoSvc)
	httpDJI := fiber.New()
	dji.NewRouter(httpDJI, eb, logger, droneSvc)
	wss := fiber.New()
	ws.NewRouter(wss, eb, logger, userSvc, droneSvc)

	var wg sync.WaitGroup

	// 启动所有服务器
	bootServers(&wg, logger, httpV1, httpDJI, wss)
	logger.Info("Servers all started")

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	logger.Info("Received shutdown signal, gracefully shutting down servers...")

	// 创建一个带有超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭所有服务器
	shutdownServers(ctx, logger, httpV1, httpDJI, wss)

	// 等待所有服务器关闭
	wg.Wait()
	logger.Info("All servers have been shut down. Exiting...")
}

func bootServers(wg *sync.WaitGroup, l *slog.Logger, apps ...*fiber.App) {
	port := 10086
	for _, app := range apps {
		wg.Add(1)
		go func(p int, a *fiber.App) {
			defer wg.Done()
			if err := a.Listen(fmt.Sprintf(":%d", p)); err != nil {
				l.Error("Server failed to start", slog.Any("err", err))
			}
		}(port, app)
		port++
	}
}

func shutdownServers(ctx context.Context, l *slog.Logger, apps ...*fiber.App) {
	for _, app := range apps {
		if err := app.ShutdownWithContext(ctx); err != nil {
			l.Error("Server shutdown error", slog.Any("err", err))
		} else {
			l.Info("Server gracefully stopped")
		}
	}
}
