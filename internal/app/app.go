package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dronesphere/internal/adapter/eventhandler"
	"github.com/dronesphere/internal/adapter/http/dji"
	v1 "github.com/dronesphere/internal/adapter/http/v1"
	"github.com/dronesphere/internal/adapter/ws"
	"github.com/dronesphere/internal/model/dto"
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
	opts.SetClientID(cfg.MQTT.ClientID + "-" + fmt.Sprintf("%d", time.Now().Unix()))
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
	droneSvc := service.NewDroneImpl(droneRepo, modelRepo, logger, client)
	saSvc := service.NewAreaImpl(saRepo, logger, client)
	wlSvc := service.NewWaylineImpl(wlRepo, logger)
	jobSvc := service.NewJobImpl(jobRepo, saRepo, droneRepo, modelRepo, wlRepo, wlSvc, logger)
	modelSvc := service.NewModelImpl(modelRepo, logger)
	gatewaySvc := service.NewGatewayImpl(gatewayRepo, logger)
	resultSvc := service.NewResultImpl(resultRepo, jobRepo, droneRepo, logger)

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
	eventhandler.NewHandler(eb, logger, client, droneSvc, gatewaySvc, modelRepo, gatewayRepo)

	// 初始化各服务
	httpV1 := fiber.New()
	v1.NewRouter(httpV1, eb, logger, container, cfg)

	httpDJI := fiber.New()
	dji.NewRouter(httpDJI, eb, logger, droneSvc, wlSvc)

	wss := fiber.New()
	ws.NewRouter(wss, eb, logger, userSvc, droneSvc)

	ctx := context.Background()
	drones, _, err := droneRepo.SelectAll(ctx, "", "", 0, 0, 0)
	if err != nil {
		logger.Error("查询无人机列表失败", slog.Any("err", err))
		return
	}
	for _, drone := range drones {
		topic := fmt.Sprintf("thing/product/%s/osd", drone.SN)

		token := client.Subscribe(topic, 0, func(c mqtt.Client, m mqtt.Message) {
			logger.Info("接收无人机 OSD 消息", slog.Any("topic", m.Topic()), slog.Any("message", string(m.Payload())))
			var p struct {
				dto.MessageCommon
				Data dto.DroneMessageProperty `json:"data"`
			}
			if err := json.Unmarshal(m.Payload(), &p); err != nil {
				logger.Error("解析无人机心跳消息失败", slog.Any("topic", m.Topic()), slog.Any("error", err))
				return
			}

			logger.Info("接收无人机 OSD 消息", slog.Any("topic", m.Topic()), slog.Any("payload", p))

			if err := droneSvc.UpdateStateBySN(ctx, drone.SN, p.Data); err != nil {
				logger.Error("更新无人机实时数据失败", slog.Any("err", err))
				return
			}
			logger.Info("更新无人机实时数据成功", slog.Any("droneSN", drone.SN))
		})
		if token.Wait() && token.Error() != nil {
			logger.Error("无人机 OSD 订阅失败", slog.Any("topic", topic), slog.Any("err", token.Error()))
			return
		} else {
			logger.Info("无人机 OSD 订阅成功", slog.Any("topic", topic))
		}
	}

	var wg sync.WaitGroup
	// 启动所有服务器
	bootServers(cfg, &wg, logger, httpV1, httpDJI, wss)
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

func bootServers(cfg *configs.Config, wg *sync.WaitGroup, l *slog.Logger, apps ...*fiber.App) {
	port := cfg.Server.Port
	// 创建互斥锁，确保服务器逐个启动
	var mu sync.Mutex
	var startupWg sync.WaitGroup

	for i, app := range apps {
		wg.Add(1)
		startupWg.Add(1)

		go func(index int, p int, a *fiber.App) {
			defer wg.Done()

			// 使用互斥锁确保服务器按顺序启动
			mu.Lock()
			serverName := fmt.Sprintf("Server %d", index+1)
			l.Info("Starting server", slog.String("server", serverName), slog.Int("port", p))

			// 启动服务器
			go func() {
				// 服务器启动后释放锁，允许下一个服务器启动
				defer mu.Unlock()
				defer startupWg.Done()

				// 短暂延迟，确保服务器有时间进行初始化
				time.Sleep(100 * time.Millisecond)
				l.Info("Server started", slog.String("server", serverName), slog.Int("port", p))
			}()

			if err := a.Listen(fmt.Sprintf(":%d", p)); err != nil {
				l.Error("Server failed to start", slog.String("server", serverName), slog.Int("port", p), slog.Any("err", err))
			}
		}(i, port, app)

		port++
	}

	// 等待所有服务器完成启动过程
	startupWg.Wait()
	l.Info("All servers started sequentially")
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
