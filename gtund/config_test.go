package gtund

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	var cfg = `
	{
		"god_config":{
			"god_hb_interval": 10,
			"god_conn_timeout":10,
			"god_addr": "127.0.0.1:9876"
		}
	}
	`

	config, err := parseConfig([]byte(cfg))
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, config)

	t.Log(config.String())
}
