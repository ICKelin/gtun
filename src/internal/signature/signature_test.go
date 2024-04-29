package signature

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSignature(t *testing.T) {
	convey.Convey("test sign and un sign", t, func() {
		convey.Convey("test sign and uns ign normal", func() {
			body := "{}"
			signature = "test"
			rs, err := Sign([]byte(body))
			convey.So(err, convey.ShouldBeNil)
			unSignBody, err := UnSign(rs)
			convey.So(err, convey.ShouldBeNil)
			convey.So(string(unSignBody), convey.ShouldEqual, body)
		})

		convey.Convey("test empty content un sign", func() {
			body := ""
			signature = "test"
			rs, err := Sign([]byte(body))
			convey.So(err, convey.ShouldBeNil)

			unSignBody, err := UnSign(rs)
			convey.So(err, convey.ShouldEqual, nil)
			convey.So(string(unSignBody), convey.ShouldEqual, body)
		})
	})
}
