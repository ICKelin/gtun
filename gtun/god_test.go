package gtun

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGod(t *testing.T) {
	g := NewGod(&GodConfig{
		GodAddr:  "http://127.0.0.1:2002",
		GodToken: "abcdefg",
	})
	body, err := g.Access()
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, body)
}
