package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	cnt := `
		{
			"gtund_config":{
				"gtund_listener": ":9876"
			},
			"gtun_config":{
				"gtun_listener":":9875"
			},
			"database":{
				"url":"mongodb://127.0.0.1:27017",
				"dbname":"gtun"
			}
		}
	`

	config, err := parseConfig([]byte(cnt))
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, config)
	t.Log(config.String())
}
