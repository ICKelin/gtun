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
	ListenAddr string `json:"listen_addr"`
	CIDR       string `json:"cidr"`
	Region     string `json:"region"`
}

type G2SRegister struct {
}
