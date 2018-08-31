package gtund

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrameInvalid(t *testing.T) {
	type TestCase struct {
		input   []byte
		invalid bool
	}

	var units = []TestCase{
		TestCase{
			input:   []byte{},
			invalid: true,
		},
		TestCase{
			input:   []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x08, 0x00},
			invalid: false,
		},
		TestCase{
			input:   []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x08, 0x00, 0x04},
			invalid: false,
		},
	}

	for _, unit := range units {
		got := Frame(unit.input).Invalid()
		assert.Equal(t, unit.invalid, got, fmt.Sprintf("expected %v got %v", unit.invalid, got))
	}
}

func TestIPV4(t *testing.T) {
	type TestCase struct {
		input []byte
		ipv4  bool
	}

	var units = []TestCase{
		TestCase{
			input: []byte{},
			ipv4:  false,
		},
		TestCase{
			input: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x08, 0x00},
			ipv4:  true,
		},
		TestCase{
			input: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x08, 0x00, 0x04},
			ipv4:  true,
		},
	}

	for _, unit := range units {
		got := Frame(unit.input).IsIPV4()
		assert.Equal(t, unit.ipv4, got, fmt.Sprintf("%v expected %v got %v", unit.input, unit.ipv4, got))
	}
}
