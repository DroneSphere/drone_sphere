package wpml

import "testing"

func TestEndToEnd(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		// 准备测试数据
		drone := DroneInfo{}
		if ok := drone.InferenceByModel("M3E"); !ok {
			t.Fatal("InferenceByModel failed")
		}
		payload := PayloadInfo{
			PayloadEnumValue:     PayloadM3E,
			PayloadSubEnumValue:  PayloadSubM3E,
			PayloadPositionIndex: 0,
		}

		builder := NewBuilder().Init("郭苏睿").SetDefaultMissionConfig(drone, payload)
		fBuilder := builder.Template.CreateFolder(TemplateTypeWaypoint, 0)
		type Mark struct {
			Lng    float64
			Lat    float64
			Height float64
		}
		marks := []Mark{
			{Lng: 117.132273581859, Lat: 36.6669406735414, Height: 30},
			{Lng: 117.132818772702, Lat: 36.6668672842279, Height: 30},
			{Lng: 117.132573075258, Lat: 36.6665315078118, Height: 30},
		}
		for _, mark := range marks {
			fBuilder.AppendDefaultPlacemark(mark.Lng, mark.Lat)
		}
		templateXML, err := builder.Template.GenerateXML()
		if err != nil {
			t.Errorf("XML generation failed: %v", err)
		}
		t.Logf("Generated Template XML with Waypoints:\n%s", templateXML)

		builder.GenerateWayline()
		waylineXML, err := builder.Wayline.GenerateXML()
		if err != nil {
			t.Errorf("XML generation failed: %v", err)
		}
		t.Logf("Generated Wayline XML:\n%s", waylineXML)

		filename := "demo.kmz"
		if err := GenerateKMZ(filename, templateXML, waylineXML); err != nil {
			t.Errorf("KMZ generation failed: %v", err)
		}
	})
}
