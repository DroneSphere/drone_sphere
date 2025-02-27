package wpml

import "time"

type TemplateBuilder struct {
	Doc Document
}

// NewTemplateBuilder 创建一个新的 TemplateBuilder
func NewTemplateBuilder(author string) *TemplateBuilder {
	ti := time.Now().UnixMilli()
	return &TemplateBuilder{
		Doc: Document{
			Author:     &author,
			CreateTime: &ti,
			UpdateTime: &ti,
		},
	}
}

// FromDocument 从 Document 生成一个 TemplateBuilder
func FromDocument(doc Document) *TemplateBuilder {
	return &TemplateBuilder{
		Doc: doc,
	}
}

func (b *TemplateBuilder) IsTemplate() bool {
	return true
}

func (b *TemplateBuilder) IsWayline() bool {
	return false
}

// SetMissionConfig 设置 MissionConfig
func (b *TemplateBuilder) SetMissionConfig(config MissionConfig) *TemplateBuilder {
	b.Doc.Mission = config
	return b
}

// CreateFolder 创建 Folder
func (b *TemplateBuilder) CreateFolder(t TemplateType, id int) *FolderBuilder {
	var folder Folder
	switch t {
	case TemplateTypeWaypoint:
		folder = DefaultWaypointFolder(id)
	}
	b.Doc.Folders = append(b.Doc.Folders, folder)
	return &FolderBuilder{
		Folder: &b.Doc.Folders[len(b.Doc.Folders)-1],
	}
}

// GenerateXML 输出 XML 文件
func (b *TemplateBuilder) GenerateXML() (string, error) {
	return b.Doc.GenerateXML()
}
