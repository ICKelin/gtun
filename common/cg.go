package common

type C2GRegister struct {
	IsWindows bool   `json:"is_windows"`
	AuthToken string `json:"auth_token"`
}

type G2SRgister struct {
	GtundAddress string `json:"gtund_addr"`
}
