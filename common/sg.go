package common

// S => gtund(gtun server)
// G => god
const (
	S2G_REGISTER = iota
	G2S_REGISTER

	S2G_HEARTBEAT
	G2S_HEARTBEAT

	S2G_UPDATE_CLIENT_COUNT
	G2S_UPDATE_CLIENT_COUNT
)

type S2GRegister struct {
	PublicIP       string `json:"public_ip"`
	Port           int    `json:"listen_port"`
	CIDR           string `json:"cidr"`
	Region         string `json:"region"`
	Token          string `json:"token"`
	Count          int    `json:"count"`
	MaxClientCount int    `json:"max_client_count"`
	IsWindows      bool   `json:"is_windows"`
}

type S2GUpdate struct {
	Count int `json:"count"`
}
