package wpml

import (
	"encoding/xml"
	"fmt"
)

func FormatCoordinates(lng, lat float64) string {
	return FormatFloat(lng) + "," + FormatFloat(lat)
}

func FormatFloat(f float64) string {
	return fmt.Sprintf("%.13f", f)
}

type BoolAsInt bool

func (b BoolAsInt) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	value := 0
	if b {
		value = 1
	}
	return e.EncodeElement(value, start)
}
