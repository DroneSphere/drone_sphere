package app

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bytedance/sonic"
	"github.com/dronesphere/internal/adapter/eventhandler"
	"github.com/dronesphere/internal/adapter/http/dji"
	v1 "github.com/dronesphere/internal/adapter/http/v1"
	"github.com/dronesphere/internal/adapter/ws"
	"github.com/dronesphere/internal/service"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2"
	"github.com/lmittmann/tint"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/configs"
	"github.com/dronesphere/internal/repo"
	slogGorm "github.com/orandin/slog-gorm"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
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
	logger.Info("Initializing database connection", slog.Any("dsn", cfg.GetDBStr()))
	db, err := gorm.Open(mysql.Open(cfg.GetDBStr()), &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(err)
	}
	logger.Info("RDB connected")

	// MQTT
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTT.Broker)
	opts.SetClientID(cfg.MQTT.ClientID)
	opts.SetUsername(cfg.MQTT.Username)
	opts.SetPassword(cfg.MQTT.Password)
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
	opt, err := redis.ParseURL(cfg.GetRedisStr())
	if err != nil {
		panic(err)
	}
	rds := redis.NewClient(opt)
	logger.Info("Redis connected")

	// S3 Storage
	endpoint := "47.245.40.222:9000"
	accessKeyID := "LxsGNjx6YKIolXHMS8EA"
	secretAccessKey := "AGmpoYvM4dZM2lRS9M3EKOsH5XQJG3dWutk3xWqV"
	// Initialize minio client object.
	s3Client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		logger.Error("S3 client error", slog.Any("err", err))
	}
	logger.Info("S3 client connected")

	// Repos
	userRepo := repo.NewUserGormRepo(db, logger)
	droneRepo := repo.NewDroneGormRepo(db, rds, logger)
	saRepo := repo.NewAreaDefaultRepo(db, rds, logger)
	wlRepo := repo.NewWaylineGormRepo(db, s3Client, logger)
	jobRepo := repo.NewJobDefaultRepo(db, s3Client, rds, logger)
	modelRepo := repo.NewModelDefaultRepo(db, logger)
	gatewayRepo := repo.NewGatewayRepo(db, logger)
	resultRepo := repo.NewResultDefaultRepo(db, logger)

	// Services
	userSvc := service.NewUserSvc(userRepo, logger)
	droneSvc := service.NewDroneImpl(droneRepo, logger, client)
	saSvc := service.NewAreaImpl(saRepo, logger, client)
	wlSvc := service.NewWaylineImpl(wlRepo, logger)
	jobSvc := service.NewJobImpl(jobRepo, saRepo, droneRepo, logger)
	modelSvc := service.NewModelImpl(modelRepo, logger)
	gatewaySvc := service.NewGatewayImpl(gatewayRepo, logger)
	resultSvc := service.NewResultImpl(resultRepo, jobRepo, logger)

	// Service Container
	container := service.NewContainer(
		userSvc,
		droneSvc,
		saSvc,
		wlSvc,
		jobSvc,
		modelSvc,
		gatewaySvc,
		resultSvc,
		logger,
	)

	// Event Handlers
	eventhandler.NewHandler(eb, logger, client, droneSvc, modelRepo, gatewayRepo)

	// Servers
	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	// Routes
	v1.NewRouter(app, eb, logger, container)
	dji.NewRouter(app, eb, logger, droneSvc, wlSvc)
	ws.NewRouter(app, eb, logger, userSvc, droneSvc)

	// Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := <-sigChan
		logger.Info("Received signal", slog.Any("signal", s))

		if err := app.Shutdown(); err != nil {
			logger.Error("Failed to shutdown HTTP server", slog.Any("err", err))
		}
	}()

	// 启动 HTTP 服务器
	if err := app.Listen(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		logger.Error("Failed to start HTTP server", slog.Any("err", err))
	}

	wg.Wait()
}
