package dto

type ProductTopo struct {
	Domain       string `json:"domain"`
	Type         int    `json:"type"`
	SubType      int    `json:"sub_type"`
	DeviceSecret string `json:"device_secret"`
	Nonce        string `json:"nonce"`
	ThingVersion string `json:"thing_version"`
}
type UpdateTopoPayload struct {
	ProductTopo
	SubDevices []struct {
		SN string `json:"sn"`
		ProductTopo
		Index string `json:"index"`
	} `json:"sub_devices"`
}
