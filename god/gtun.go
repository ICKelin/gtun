package god

import (
	"net/http"
)

type gtunConfig struct {
	Listener string `json:"gtun_listener"`
}

type gtun struct {
	listener string
}

func NewGtun(cfg *gtunConfig) *gtun {
	return &gtun{
		listener: cfg.Listener,
	}
}

func (c *gtun) Run() error {
	http.HandleFunc("/gtun/register", c.onRegister)
	return http.ListenAndServe(c.listener, nil)
}

func (c *gtun) onRegister(w http.ResponseWriter, r *http.Request) {
	// TODO
}
