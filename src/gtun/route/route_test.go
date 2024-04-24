package route

import (
	"github.com/ICKelin/gtun/src/internal/logs"
	"github.com/ICKelin/optw/transport"
	gomonkey "github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"net"
	"strings"
	"testing"
	"time"
)

func TestRoute(t *testing.T) {
	convey.Convey("test route", t, func() {
		convey.Convey("not exist region", func() {
			conn := GetRouteManager().Route("unknown region", "127.0.0.1")
			convey.So(conn, convey.ShouldBeNil)
		})

		convey.Convey("not route to host", func() {
			GetRouteManager().routeTable["region1"] = make([]*routeItem, 0)
			conn := GetRouteManager().Route("unknown region", "127.0.0.1")
			convey.So(conn, convey.ShouldBeNil)
			delete(GetRouteManager().routeTable, "region1")
		})

		convey.Convey("best node", func() {
			p := gomonkey.ApplyPrivateMethod(tm, "getRegionBestTarget", func(tm *traceManager, region string) (traceTarget, bool) {
				return traceTarget{
					traceAddr:  "",
					serverAddr: "127.0.0.1:5201",
					scheme:     "kcp",
				}, true
			})
			defer p.Reset()

			convey.Convey("best node match", func() {
				newRoute := &routeItem{
					region:     "",
					scheme:     "kcp",
					serverAddr: "127.0.0.1:5201",
					Conn:       &mockConn{},
				}
				GetRouteManager().addRoute("region1", newRoute)
				defer GetRouteManager().deleteRoute("region1", newRoute)

				conn := GetRouteManager().Route("region1", "127.0.0.1:8954")
				convey.So(conn, convey.ShouldNotBeNil)
			})

			convey.Convey("best node not match", func() {
				gomonkey.ApplyFunc(logs.Warn, func(f interface{}, v ...interface{}) {
					convey.So(strings.Contains(f.(string), "use random hop"), convey.ShouldBeTrue)
				})
				newRoute := &routeItem{
					region:     "",
					scheme:     "kcp",
					serverAddr: "127.0.0.1:5202",
					Conn:       &mockConn{},
				}
				GetRouteManager().addRoute("region1", newRoute)
				defer GetRouteManager().deleteRoute("region1", newRoute)

				conn := GetRouteManager().Route("region1", "127.0.0.1:8954")
				convey.So(conn, convey.ShouldNotBeNil)
			})
		})
	})
}

type mockConn struct {
}

func (m *mockConn) LocalAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (m *mockConn) SetDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (m *mockConn) OpenStream() (transport.Stream, error) {
	//TODO implement me
	panic("implement me")
}

func (m *mockConn) AcceptStream() (transport.Stream, error) {
	//TODO implement me
	panic("implement me")
}

func (m *mockConn) Close() {
	//TODO implement me
	panic("implement me")
}

func (m *mockConn) IsClosed() bool {
	return false
}

func (m *mockConn) RemoteAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}
