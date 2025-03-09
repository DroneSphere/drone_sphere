package service

import (
	"context"
	"errors"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/vo"
	"github.com/dronesphere/pkg/wpml"
	"github.com/google/uuid"
	"log/slog"
	"os"
)

type (
	WaylineSvc interface {
		Create(ctx context.Context, points []vo.GeoPoint, droneSN string, height float64) (entity.Wayline, error)
		FetchAll(ctx context.Context) ([]entity.Wayline, error)
		Download(ctx context.Context, key string) (string, error)
	}

	WaylineRepo interface {
		SaveToS3(ctx context.Context, path, key string) (string, error)
		Save(ctx context.Context, wayline entity.Wayline) (entity.Wayline, error)
		FetchKMZByKey(ctx context.Context, key string) (string, error)
		FetchAll(ctx context.Context) ([]entity.Wayline, error)
	}
)

type WaylineImpl struct {
	wr WaylineRepo
	dr DroneRepo
	l  *slog.Logger
}

func NewWaylineImpl(wr WaylineRepo, dr DroneRepo, l *slog.Logger) WaylineSvc {
	return &WaylineImpl{
		wr: wr,
		dr: dr,
		l:  l,
	}
}

// Create TODO: 生成航线
func (s *WaylineImpl) Create(ctx context.Context, points []vo.GeoPoint, droneSN string, height float64) (entity.Wayline, error) {
	s.l.Info("Create", slog.Any("points", points), slog.Any("droneSN", droneSN), slog.Any("height", height))
	//drone, err := s.dr.FetchStateBySN(ctx, droneSN)
	//if err != nil {
	//	return entity.Wayline{}, err
	//}
	//s.l.Info("Create", slog.Any("drone", drone))
	// 准备 wpml 需要的 DroneInfo 和 PayloadInfo
	droneInfo := wpml.DroneInfo{}
	if ok := droneInfo.InferenceByModel("M3E"); !ok {
		s.l.Error("InferenceByModel failed")
		return entity.Wayline{}, errors.New("InferenceByModel failed")
	}
	payload := wpml.PayloadInfo{
		PayloadEnumValue:     wpml.PayloadM3E,
		PayloadSubEnumValue:  wpml.PayloadSubM3E,
		PayloadPositionIndex: 0,
	}
	// wpml 构建工具初始化
	// 生成航线模板文件
	const author = "System"
	builder := wpml.NewBuilder().Init(author).SetDefaultMissionConfig(droneInfo, payload)
	fBuilder := builder.Template.CreateFolder(wpml.TemplateTypeWaypoint, 0)
	for _, e := range points {
		fBuilder.AppendDefaultPlacemark(e.Lng, e.Lat)
	}
	templateXML, err := builder.Template.GenerateXML()
	if err != nil {
		return entity.Wayline{}, err
	}
	// 生成航线执行文件
	builder.GenerateWayline()
	waylineXML, err := builder.Wayline.GenerateXML()
	if err != nil {
		return entity.Wayline{}, err
	}
	// 生成 KMZ 文件
	// 随机UUID作为文件名
	filename := uuid.New().String() + ".kmz"
	if err := wpml.GenerateKMZ(filename, templateXML, waylineXML); err != nil {
		return entity.Wayline{}, err
	}
	s.l.Info("Create", slog.Any("filename", filename))
	// 清理本地文件
	defer func() {
		if err := os.Remove(filename); err != nil {
			s.l.Error("Remove failed", slog.Any("filename", filename))
		}
	}()

	// 上传到 S3
	url, err := s.wr.SaveToS3(ctx, filename, filename)
	if err != nil {
		s.l.Error("SaveToS3 failed", slog.Any("filename", filename))
		return entity.Wayline{}, err
	}
	s.l.Info("Create", slog.Any("url", url))

	// 保存到数据库
	en := entity.Wayline{
		UploadUser: author,
		Area:       entity.SearchArea{},
		Drone:      entity.Drone{},
		Points:     points,
		S3Key:      filename,
		S3Url:      url,
	}
	if _, err := s.wr.Save(ctx, en); err != nil {
		s.l.Error("Save failed", slog.Any("en", en))
		return entity.Wayline{}, err
	}
	s.l.Info("Create", slog.Any("en", en))

	return en, nil
}

func (s *WaylineImpl) FetchAll(ctx context.Context) ([]entity.Wayline, error) {
	return s.wr.FetchAll(ctx)
}

func (s *WaylineImpl) Download(ctx context.Context, key string) (string, error) {
	return s.wr.FetchKMZByKey(ctx, key)
}
