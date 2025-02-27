package wpml

import (
	"testing"
)

func TestAppendPlacemark(t *testing.T) {
	t.Run("append to empty folder", func(t *testing.T) {
		b := &FolderBuilder{Folder: &Folder{Placemarks: []Placemark{}}}
		trueBool := BoolAsInt(true)
		p := Placemark{IsRisky: &trueBool, Point: Point{Coordinates: "1,2"}}

		result := b.AppendPlacemark(p)

		if result != b {
			t.Error("Expected to return the same builder instance")
		}
		if len(b.Folder.Placemarks) != 1 {
			t.Errorf("Expected 1 placemark, got %d", len(b.Folder.Placemarks))
		}
		if b.Folder.Placemarks[0].Index != 0 {
			t.Errorf("Expected index 0, got %d", b.Folder.Placemarks[0].Index)
		}
	})

	t.Run("append multiple placemarks", func(t *testing.T) {
		b := &FolderBuilder{Folder: &Folder{Placemarks: []Placemark{}}}
		p1 := Placemark{Point: Point{Coordinates: "1,2"}}
		p2 := Placemark{Point: Point{Coordinates: "3,4"}}

		b.AppendPlacemark(p1).AppendPlacemark(p2)

		if len(b.Folder.Placemarks) != 2 {
			t.Fatalf("Expected 2 placemarks, got %d", len(b.Folder.Placemarks))
		}
		if b.Folder.Placemarks[1].Index != 1 {
			t.Errorf("Expected index 1 for second placemark, got %d", b.Folder.Placemarks[1].Index)
		}
	})

	t.Run("preserve existing fields", func(t *testing.T) {
		b := &FolderBuilder{Folder: &Folder{Placemarks: []Placemark{}}}
		trueBool := BoolAsInt(true)
		h := 100.0
		p := Placemark{
			IsRisky:         &trueBool,
			EllipsoidHeight: &h,
			Point:           Point{Coordinates: "1,2"},
		}

		b.AppendPlacemark(p)

		result := b.Folder.Placemarks[0]
		if !*result.IsRisky || *result.EllipsoidHeight != 100 {
			t.Error("Placemark fields not preserved correctly")
		}
	})
}

func TestAppendDefaultPlacemark(t *testing.T) {
	t.Run("basic default values", func(t *testing.T) {
		b := &FolderBuilder{Folder: &Folder{Placemarks: []Placemark{}}}
		lng, lat := 100.5, 30.3

		b.AppendDefaultPlacemark(lng, lat)

		p := b.Folder.Placemarks[0]
		if p.Point.Coordinates != "100.500000,30.300000" {
			t.Errorf("Unexpected coordinates format: %s", p.Point.Coordinates)
		}
		if !*p.UseGlobalHeight || !*p.UseGlobalSpeed {
			t.Error("Default boolean flags not set correctly")
		}
	})

	t.Run("index assignment", func(t *testing.T) {
		b := &FolderBuilder{Folder: &Folder{Placemarks: make([]Placemark, 2)}}

		b.AppendDefaultPlacemark(0, 0)
		if b.Folder.Placemarks[2].Index != 2 {
			t.Errorf("Expected index 2 for third placemark, got %d", b.Folder.Placemarks[2].Index)
		}
	})

	t.Run("coordinate formatting", func(t *testing.T) {
		testCases := []struct {
			lng      float64
			lat      float64
			expected string
		}{
			{123.4567, 45.6789, "123.456700,45.678900"},
			{-122.123, 37.456, "-122.123000,37.456000"},
		}

		for _, tc := range testCases {
			b := &FolderBuilder{Folder: &Folder{Placemarks: []Placemark{}}}
			b.AppendDefaultPlacemark(tc.lng, tc.lat)
			if actual := b.Folder.Placemarks[0].Point.Coordinates; actual != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, actual)
			}
		}
	})
}
