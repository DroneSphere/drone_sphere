package wpml

import (
	"testing"
	"time"
)

func TestNewTemplateBuilder(t *testing.T) {
	author := "test_author"
	preTime := time.Now().UnixMilli()
	builder := NewTemplateBuilder(author)
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
}

func TestTemplateFromDocument(t *testing.T) {
	a := "existing_doc"
	ti := time.Now().UnixMilli()
	doc := Document{
		Author:     &a,
		CreateTime: &ti,
		Mission:    MissionConfig{},
	}
	builder := FromDocument(doc)

	if builder.Doc.Author != doc.Author {
		t.Error("Document not properly initialized")
	}
	if builder.Doc.CreateTime != doc.CreateTime {
		t.Error("CreateTime not preserved")
	}
}

func TestTemplateTypeCheckers(t *testing.T) {
	builder := NewTemplateBuilder("test")
	if !builder.IsTemplate() {
		t.Error("IsTemplate should return true")
	}
	if builder.IsWayline() {
		t.Error("IsWayline should return false")
	}
}

func TestTemplateSetMissionConfig(t *testing.T) {
	builder := NewTemplateBuilder("test")
	config := MissionConfig{}

	result := builder.SetMissionConfig(config)

	// 测试链式调用
	if result != builder {
		t.Error("Should return self for chaining")
	}
}

func TestTemplateCreateFolder(t *testing.T) {
	builder := NewTemplateBuilder("test")

	t.Run("first folder creation", func(t *testing.T) {
		fb := builder.CreateFolder(TemplateTypeWaypoint, 1)

		if len(builder.Doc.Folders) != 1 {
			t.Fatal("Folder not added to document")
		}
		// 测试文件夹参数
		folder := builder.Doc.Folders[0]
		if *folder.TemplateType != TemplateTypeWaypoint || *folder.TemplateID != 1 {
			t.Error("Folder parameters mismatch")
		}
		// 测试返回的构建器
		if *fb.Folder.TemplateID != 1 {
			t.Error("FolderBuilder not initialized with correct folder")
		}
	})

	t.Run("multiple folders", func(t *testing.T) {
		builder.CreateFolder(TemplateTypeWaypoint, 2)
		if len(builder.Doc.Folders) != 2 {
			t.Error("Second folder not added")
		}
	})
}

func TestTemplateGenerateXML(t *testing.T) {
	t.Run("success generation", func(t *testing.T) {
		builder := NewTemplateBuilder("test").SetMissionConfig(MissionConfig{DroneInfo: DroneInfo{DroneEnumValue: DroneM3D}})
		_, err := builder.GenerateXML()
		if err != nil {
			t.Errorf("XML generation failed: %v", err)
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		// 假设当没有Mission时会返回错误
		builder := NewTemplateBuilder("test") // 没有设置MissionConfig
		_, err := builder.GenerateXML()
		if err == nil {
			t.Error("Expected error from empty mission config")
		}
	})
}
