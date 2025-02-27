package wpml

import (
	"testing"
	"time"
)

func TestNewWaylineBuilder(t *testing.T) {
	author := "test_author"
	preTime := time.Now().UnixMilli()
	builder := NewWaylineBuilder(author)
	postTime := time.Now().UnixMilli()

	// 测试基础字段
	if *builder.Doc.Author != author {
		t.Errorf("Expected author %s, got %s", author, *builder.Doc.Author)
	}

	// 测试时间戳合理性
	if *builder.Doc.CreateTime < preTime || *builder.Doc.CreateTime > postTime {
		t.Error("CreateTime not in expected time range")
	}
	if builder.Doc.UpdateTime != builder.Doc.CreateTime {
		t.Error("UpdateTime should equal CreateTime on initialization")
	}

	// 验证模板文档为空
	if *builder.Template.Author != "" {
		t.Error("Template document should be empty for new builder")
	}
}

func TestFromTemplate(t *testing.T) {
	templateBuilder := NewTemplateBuilder("template_author")
	missionConfig := DefaultMissionConfig(DroneInfo{DroneEnumValue: DroneM30, DroneSubEnumValue: SubM30}, PayloadInfo{PayloadEnumValue: PayloadM30Dual, PayloadPositionIndex: 0})
	templateBuilder.SetMissionConfig(missionConfig)
	templateBuilder.CreateFolder(TemplateTypeWaypoint, 1)

	waylineBuilder := FromTemplate(*templateBuilder)

	// 验证文档内容
	if waylineBuilder.Doc.Author != templateBuilder.Doc.Author {
		t.Error("Author not copied correctly")
	}
	if waylineBuilder.Doc.Mission.DroneInfo != missionConfig.DroneInfo {
		t.Error("Mission config not preserved")
	}

	// 验证文件夹被清空
	if len(waylineBuilder.Doc.Folders) != 0 {
		t.Error("Folders should be cleared")
	}

	// 验证模板文档
	if waylineBuilder.Template.Author != templateBuilder.Doc.Author {
		t.Error("Template document not stored correctly")
	}
}

func TestWaylineTypeCheckers(t *testing.T) {
	builder := NewWaylineBuilder("test")
	if builder.IsTemplate() {
		t.Error("IsTemplate should return false")
	}
	if !builder.IsWayline() {
		t.Error("IsWayline should return true")
	}
}

func TestWaylineCreateFolder(t *testing.T) {
	builder := NewWaylineBuilder("test")

	t.Run("first folder creation", func(t *testing.T) {
		ty := TemplateTypeWaypoint
		folder := Folder{TemplateType: &ty}
		fb := builder.CreateFolder(folder, 1)

		if len(builder.Doc.Folders) != 1 {
			t.Fatal("Folder not added to document")
		}
		// 测试文件夹参数
		createdFolder := builder.Doc.Folders[0]
		if *createdFolder.WaylineID != 1 {
			t.Error("WaylineID not set correctly")
		}
		// 测试返回的构建器
		if *fb.Folder.WaylineID != 1 {
			t.Error("FolderBuilder not initialized with correct folder")
		}
	})

	t.Run("multiple folders", func(t *testing.T) {
		ty := TemplateTypeWaypoint
		folder := Folder{TemplateType: &ty}
		builder.CreateFolder(folder, 2)
		if len(builder.Doc.Folders) != 2 {
			t.Error("Second folder not added")
		}
	})
}

func TestWaylineGenerateXML(t *testing.T) {
	t.Run("success generation", func(t *testing.T) {
		builder := NewWaylineBuilder("test")
		builder.Doc.Mission = DefaultMissionConfig(DroneInfo{DroneEnumValue: DroneM30, DroneSubEnumValue: SubM30}, PayloadInfo{PayloadEnumValue: PayloadM30Dual, PayloadPositionIndex: 0})
		_, err := builder.GenerateXML()
		if err != nil {
			t.Errorf("XML generation failed: %v", err)
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		// 假设当没有Mission时会返回错误
		builder := NewWaylineBuilder("test") // 没有设置MissionConfig
		_, err := builder.GenerateXML()
		if err == nil {
			t.Error("Expected error from empty mission config")
		}
	})
}
