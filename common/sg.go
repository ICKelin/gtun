package common

// S => gtund(gtun server)
// G => gtun
const (
	S2G_REGISTER = iota
	G2S_REGISTER

	S2G_HEARTBEAT
	G2S_HEARTBEAT
)

type S2GRegister struct {
	PublicIP   string `json:"public_ip"`
	ListenPort string `json:"listen_port"`
	CIDR       string `json:"cidr"`
}

type G2SRegister struct {
}
