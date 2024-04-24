package config

import (
	"github.com/ICKelin/gtun/src/internal/signature"
	"github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

var testConfig = `
access_token: "${ACCESS_TOKEN}"
accelerator:
  HK:
    routes:
      - scheme: "kcp"
        server: "${SERVER_IP}:3002"
        trace: "${SERVER_IP}:3003"
      - scheme: "mux"
        server: "${SERVER_IP}:3002"
        trace: "${SERVER_IP}:3003"
    proxy:
      tproxy_tcp: |
        {
          "read_timeout": 30,
          "write_timeout": 30,
          "listen_addr": ":8524"
        }
      tproxy_udp: |
        {
          "read_timeout": 30,
          "write_timeout": 30,
          "session_timeout": 30,
          "listen_addr": ":8524"
        }
log:
  days: 5
  level: debug
  path: /opt/apps/gtun/logs/gtun.log

`

func TestConfig(t *testing.T) {
	convey.Convey("test config", t, func() {
		convey.Convey("test with env var", func() {
			os.Setenv("SERVER_IP", "127.0.0.1")
			os.Setenv("ACCESS_TOKEN", "ICKelin:free")
			cfg, err := ParseBuffer([]byte(testConfig))
			convey.So(err, convey.ShouldBeNil)
			convey.So(cfg.Accelerator["HK"].Routes[0].Server, convey.ShouldEqual, "127.0.0.1:3002")
			convey.So(cfg.AccessToken, convey.ShouldEqual, "ICKelin:free")
		})

		convey.Convey("test signature", func() {
			convey.Convey("test config without signature", func() {
				signature.SetSignature("sig")
				_, err := ParseBuffer([]byte(testConfig))
				convey.So(err, convey.ShouldNotBeNil)
			})

			convey.Convey("test config with signature", func() {
				signature.SetSignature("sig")
				buf, err := signature.Sign([]byte(testConfig))
				convey.So(err, convey.ShouldBeNil)
				_, err = ParseBuffer(buf)
				convey.So(err, convey.ShouldBeNil)
			})
		})
	})
}
