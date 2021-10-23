package kcp

type KCPConfig struct {
	// fec args
	FecDataShards   int `json:"dataShards"`
	FecParityShards int `json:"parityShards"`
	// nodelay config args
	Nodelay  int `json:"nodelay"`
	Interval int `json:"interval"`
	Resend   int `json:"resend"`
	Nc       int `json:"nc"`
	// windows size
	SndWnd     int  `json:"sndwnd"`
	RcvWnd     int  `json:"rcvwnd"`
	Mtu        int  `json:"mtu"`
	AckNoDelay bool `json:"ackNoDelay"`
	Rcvbuf     int  `json:"rcvBuf"`
	SndBuf     int  `json:"sndBuf"`
}
