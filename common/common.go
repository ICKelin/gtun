package common

type C2SAuthorize struct {
	AccessIP string `json:"access_ip"`
	Key      string `json:"key"`
}
type S2CAuthorize struct {
	Status   string `json:"status"`
	AccessIP string `json:"access_ip"`
}
