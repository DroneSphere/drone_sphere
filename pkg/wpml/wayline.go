package wpml

import "time"

type WaylineBuilder struct {
	Doc      Document
	Template Document
}

// NewWaylineBuilder 创建一个新的 WaylineBuilder
func NewWaylineBuilder(author string) *WaylineBuilder {
	ti := time.Now().UnixMilli()
	return &WaylineBuilder{
		Doc: Document{
			Author:     &author,
			CreateTime: &ti,
			UpdateTime: &ti,
		},
	}
}

// FromTemplate 从 Template 生成一个 WaylineBuilder
func FromTemplate(t TemplateBuilder) *WaylineBuilder {
	doc := t.Doc
	doc.Author = nil
	doc.CreateTime = nil
	doc.UpdateTime = nil
	for i := range doc.Folders {
		f := &doc.Folders[i]
		// 添加 WaylineID
		f.WaylineID = &i
		hMode := ExecuteHeightModeWGS84
		f.ExecuteHeightMode = &hMode
		for j := range f.Placemarks {
			p := &f.Placemarks[j]
			// 处理高度
			if *p.UseGlobalHeight {
				p.ExecuteHeight = f.GlobalHeight
			} else {
				p.ExecuteHeight = p.EllipsoidHeight
			}
			p.EllipsoidHeight = nil
			p.Height = nil
			p.UseGlobalHeight = nil
			// 处理速度
			if *p.UseGlobalSpeed {
				p.WaypointSpeed = f.AutoFlightSpeed
			}
			p.UseGlobalSpeed = nil
			// 处理 GlobalHeadingParam
			if *p.UseGlobalHeadingParam {
				p.WaypointHeadingParam = f.GlobalWaypointHeadingParam
			}
			p.UseGlobalHeadingParam = nil
			// 处理 WaypointTurnParam
			if *p.UseGlobalTurnParam {
				//p.WaypointTurnParam.WaypointTurnMode = *f.GlobalWaypointTurnMode
				p.WaypointTurnParam = &WaypointTurnParam{
					WaypointTurnMode: *f.GlobalWaypointTurnMode,
				}
			}
			p.UseGlobalTurnParam = nil
			// 处理 StraightLine
			if *p.UseStraightLine {
				p.UseStraightLine = f.GlobalUseStraightLine
			}
			p.UseStraightLine = nil
			// 处理 GimbalPitchMode
			if *f.GimbalPitchMode == GimbalPitchModeManual {
				p.WaypointGimbalHeadingParam = &WaypointGimbalHeadingParam{
					WaypointGimbalPitchAngle: 0,
					WaypointGimbalYawAngle:   0,
				}
			}
			workType := WaypointWorkTypeNone
			p.WaypointWorkType = &workType
		}
		// 擦除不需要的字段
		f.TemplateType = nil
		f.WaylineCoordinateSysParam = nil
		f.PayloadParam = nil
		f.GlobalWaypointTurnMode = nil
		f.GlobalUseStraightLine = nil
		f.GimbalPitchMode = nil
		f.GlobalHeight = nil
		f.GlobalWaypointHeadingParam = nil
	}
	return &WaylineBuilder{
		Doc:      doc,
		Template: t.Doc,
	}
}

func (b *WaylineBuilder) IsTemplate() bool {
	return false
}

func (b *WaylineBuilder) IsWayline() bool {
	return true
}

// CreateFolder 创建 Folder
func (b *WaylineBuilder) CreateFolder(folder Folder, id int) *FolderBuilder {
	folder.WaylineID = &id
	b.Doc.Folders = append(b.Doc.Folders, folder)
	return &FolderBuilder{
		Folder: &b.Doc.Folders[len(b.Doc.Folders)-1],
	}
}

// GenerateXML 输出 XML 文件
func (b *WaylineBuilder) GenerateXML() (string, error) {
	return b.Doc.GenerateXML()
}
