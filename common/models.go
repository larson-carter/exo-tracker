package models

type Peer struct {
	ID                 string                 `json:"node_id"`
	IP                 string                 `json:"ip"`
	Port               int                    `json:"port"` // Change to int
	DeviceCapabilities map[string]interface{} `json:"device_capabilities"`
}
