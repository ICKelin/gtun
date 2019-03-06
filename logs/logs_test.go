package logs

import (
	"fmt"
	"testing"
)

func TestLogs(t *testing.T) {
	Info("this is error: ", fmt.Errorf("invalid cmd"))
}
