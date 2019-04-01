package common

// S => gtund(gtun server)
// C => gtun(gtun client)
const (
	C2C_DATA = byte(0x00)

	C2S_DATA = byte(0x01)
	S2C_DATA = byte(0x02)

	C2S_HEARTBEAT = byte(0x03)
	S2C_HEARTBEAT = byte(0x04)

	C2S_AUTHORIZE = byte(0x05)
	S2C_AUTHORIZE = byte(0x06)
)

var os = map[string]int{
	"linux":   1,
	"darwin":  2,
	"windows": 3,
}

func OSID(goos string) int {
	return os[goos]
}

type C2SAuthorize struct {
	Version string `json:"version"`
	OS      int    `json:"os"`
	Key     string `json:"key"`
}

type S2CAuthorize struct {
	Status         string   `json:"status"`
	AccessIP       string   `json:"access_ip"`
	Nameservers    []string `json:"nameservers"`
	Gateway        string   `json:"gateway"`
	RouteScriptUrl string   `json:"route_script_url"`
}
