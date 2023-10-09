package route

import (
	"github.com/ICKelin/optw/transport"
	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"net"
	"testing"
)

func TestManager_AddRoute(t *testing.T) {
	Convey("Test add route", t, func() {
		Convey("add first hop for region", func() {
			GetRouteManager().AddRoute("test_region1", &HopInfo{})
			So(len(GetRouteManager().regionHops), ShouldEqual, 1)
			So(len(GetRouteManager().regionHops["test_region1"]), ShouldEqual, 1)

		})

		Convey("add second hop", func() {
			GetRouteManager().AddRoute("test_region1", &HopInfo{})
			So(len(GetRouteManager().regionHops), ShouldEqual, 1)
			So(len(GetRouteManager().regionHops["test_region1"]), ShouldEqual, 2)
		})
	})
}

func TestManager_DeleteRoute(t *testing.T) {
	Convey("Test delete route", t, func() {
		Convey("delete un exist region", func() {
			GetRouteManager().DeleteRoute("not-found", nil)
		})

		Convey("delete exist region with empty hops", func() {
			GetRouteManager().regionHops["region1"] = make([]*HopInfo, 0)
			GetRouteManager().DeleteRoute("region1", nil)
		})

		Convey("delete exist region with hops", func() {
			newHopInfo := &HopInfo{Conn: &mockConn{}}
			gomonkey.ApplyMethod(newHopInfo.Conn, "RemoteAddr", func(hopInfo transport.Conn) net.Addr {
				return &net.TCPAddr{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: 10000,
					Zone: "",
				}
			})
			GetRouteManager().AddRoute("region1", newHopInfo)
			So(len(GetRouteManager().regionHops["region1"]), ShouldEqual, 1)
			GetRouteManager().DeleteRoute("region1", newHopInfo)
			So(len(GetRouteManager().regionHops["region1"]), ShouldEqual, 0)
		})

		Convey("delete not exist hops", func() {
			newHopInfo := &HopInfo{Conn: &mockConn{}}
			i := 0
			gomonkey.ApplyMethod(newHopInfo.Conn, "RemoteAddr", func(hopInfo transport.Conn) net.Addr {
				i += 1
				return &net.TCPAddr{
					IP:   net.IPv4(127, 0, byte(i), 1),
					Port: 10000,
					Zone: "",
				}
			})

			deleteHopInfo := &HopInfo{Conn: &mockConn{}}
			gomonkey.ApplyMethod(deleteHopInfo.Conn, "RemoteAddr", func(hopInfo transport.Conn) net.Addr {
				i += 1
				return &net.TCPAddr{
					IP:   net.IPv4(127, 0, byte(i), 1),
					Port: 10000,
					Zone: "",
				}
			})

			GetRouteManager().AddRoute("region1", newHopInfo)
			So(len(GetRouteManager().regionHops["region1"]), ShouldEqual, 1)

			GetRouteManager().DeleteRoute("region1", deleteHopInfo)
			So(len(GetRouteManager().regionHops["region1"]), ShouldEqual, 1)
		})
	})
}

func TestManager_Route(t *testing.T) {
	Convey("Test route", t, func() {
		Convey("nil next hop region", func() {
			hop := GetRouteManager().Route("not-found", "")
			So(hop, ShouldBeNil)

			GetRouteManager().regionHops["region1"] = make([]*HopInfo, 0)
			hop = GetRouteManager().Route("region1", "")
			So(hop, ShouldBeNil)
		})

		Convey("next hop is closed", func() {
			hop := &HopInfo{Conn: &mockConn{}}
			gomonkey.ApplyMethod(hop.Conn, "IsClosed", func(conn transport.Conn) bool {
				return true
			})
			gomonkey.ApplyMethod(hop.Conn, "RemoteAddr", func(hopInfo transport.Conn) net.Addr {
				return &net.TCPAddr{
					IP:   net.IPv4(127, 0, byte(0), 1),
					Port: 10000,
					Zone: "",
				}
			})

			GetRouteManager().AddRoute("region1", hop)
			routeHop := GetRouteManager().Route("region1", "")
			So(routeHop, ShouldBeNil)
		})

		Convey("best ip not match", func() {
			hop := &HopInfo{Conn: &mockConn{}}
			gomonkey.ApplyMethod(hop.Conn, "IsClosed", func(conn transport.Conn) bool {
				return false
			})
			gomonkey.ApplyMethod(hop.Conn, "RemoteAddr", func(hopInfo transport.Conn) net.Addr {
				return &net.TCPAddr{
					IP:   net.IPv4(127, 0, byte(0), 1),
					Port: 10000,
					Zone: "",
				}
			})

			gomonkey.ApplyMethod(GetRouteManager().raceManager, "GetBestNode", func(manager *RaceManager) string {
				return "192.168.1.1:9000"
			})
			GetRouteManager().AddRoute("region2", hop)
			routeHop := GetRouteManager().Route("region2", "")
			// use random
			So(routeHop, ShouldNotBeNil)
			So(routeHop.RemoteAddr().String(), ShouldEqual, "127.0.0.1:10000")
		})

		Convey("best ip match", func() {
			hop := &HopInfo{Conn: &mockConn{}}
			gomonkey.ApplyMethod(hop.Conn, "IsClosed", func(conn transport.Conn) bool {
				return false
			})
			gomonkey.ApplyMethod(hop.Conn, "RemoteAddr", func(hopInfo transport.Conn) net.Addr {
				return &net.TCPAddr{
					IP:   net.IPv4(127, 0, byte(0), 1),
					Port: 10000,
					Zone: "",
				}
			})

			gomonkey.ApplyMethod(GetRouteManager().raceManager, "GetBestNode", func(manager *RaceManager) string {
				return "127.0.0.1:10000"
			})

			GetRouteManager().AddRoute("region3", hop)
			routeHop := GetRouteManager().Route("region3", "")
			So(routeHop, ShouldNotBeNil)
			So(routeHop.RemoteAddr().String(), ShouldEqual, "127.0.0.1:10000")
		})

		Convey("random next hop", func() {})
	})
}

type mockConn struct{}

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
	//TODO implement me
	panic("implement me")
}

func (m *mockConn) RemoteAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

var _ transport.Conn = &mockConn{}
