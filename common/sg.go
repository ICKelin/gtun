package common

import "encoding/json"

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

const (
	CODE_SUCCESS       = 10000
	CODE_REGISTER_FAIL = 10001
	CODE_FAIL          = 99999
)

type S2GRegister struct {
	PublicIP       string `json:"public_ip"`
	ListenAddr     string `json:"listen_addr"`
	CIDR           string `json:"cidr"`
	Region         string `json:"region"`
	Token          string `json:"token"`
	Count          int    `json:"count"`
	MaxClientCount int    `json:"max_client_count"`
}

type S2GUpdate struct {
	Count int `json:"count"`
}

type G2SResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Response(data interface{}, err error) []byte {
	g2s := &G2SResponse{}

	if err != nil {
		g2s.Code = CODE_FAIL
		g2s.Message = err.Error()
		g2s.Data = data
	} else {
		g2s.Code = CODE_SUCCESS
		g2s.Message = "success"
		g2s.Data = data
	}

	bytes, _ := json.Marshal(g2s)
	return bytes
}
