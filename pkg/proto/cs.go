package proto

const (
	CmdAuth = iota
	CmdHeartbeat
	CmdData
)

type C2SAuth struct {
	Key string `json:"key"`
}

type S2CAuth struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ProxyProtocol struct {
	Protocol string `json:"protocol"`
	SrcIP    string `json:"sip"`
	SrcPort  string `json:"sport"`
	DstIP    string `json:"dip"`
	DstPort  string `json:"dport"`
}
