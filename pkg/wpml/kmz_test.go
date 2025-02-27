package wpml

import (
	"testing"
)

func TestGenerateKMZ(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		// 准备测试数据
		filename := "./test_output.zip"
		templateContent := "<kml>test template</kml>"
		waylineContent := "<wpml>test wayline</wpml>"

		// 执行函数
		err := GenerateKMZ(filename, templateContent, waylineContent)
		if err != nil {
			t.Fatalf("GenerateKMZ failed: %v", err)
		}
	})
}
