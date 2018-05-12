package dns

import (
	"testing"
)

func Test_LoadPolicyRule(t *testing.T) {
	selfdef := NewSelfDefine("./config/", "./config/resolve.conf")
	selfdef.Run()

	selfdef.buildInCache.String()
}
