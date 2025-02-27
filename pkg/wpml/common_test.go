package wpml

import (
	"testing"
	"time"
)

func TestDocument_GenerateXML(t *testing.T) {
	a := "Test Author"
	ti := time.Now().UnixMilli()
	doc := Document{
		Author:     &a,
		CreateTime: &ti,
		UpdateTime: &ti,
		Mission:    DefaultMissionConfig(DroneInfo{DroneEnumValue: DroneM30, DroneSubEnumValue: SubM30}, PayloadInfo{PayloadEnumValue: PayloadM30Dual, PayloadPositionIndex: 0}),
	}
	id := 0
	ty := TemplateTypeWaypoint
	doc.Folders = append(doc.Folders, Folder{
		TemplateType: &ty,
		TemplateID:   &id,
		Placemarks: []Placemark{
			DefaultPlacemark(23.98057, 115.987663),
			DefaultPlacemark(24.323345, 116.324532),
		}})

	xml, err := doc.GenerateXML()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(xml) == 0 {
		t.Fatalf("Expected non-empty XML, got empty string")
	}
}
