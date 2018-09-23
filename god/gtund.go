package god

import (
	"encoding/json"
	"net"
	"time"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
)

type gtundConfig struct {
	Listener string `json:"gtund_listener"`
}

type gtund struct {
	listener string
}

func NewGtund(cfg *gtundConfig) *gtund {
	return &gtund{
		listener: cfg.Listener,
	}
}

func (d *gtund) Run() error {
	listener, err := net.Listen("tcp", d.listener)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go d.onConn(conn)
	}
}

func (d *gtund) onConn(conn net.Conn) {
	defer conn.Close()
	cmd, bytes, err := common.Decode(conn)
	if err != nil {
		glog.ERROR(err)
		return
	}

	if cmd != common.S2G_REGISTER {
		glog.ERROR("invalid cmd", cmd)
		return
	}

	reg := common.S2GRegister{}
	err = json.Unmarshal(bytes, &reg)
	if err != nil {
		glog.ERROR(err)
		return
	}

	// TODO: store gtund register infos
	// TODO: response register

	glog.INFO("register gtund from ", conn.RemoteAddr().String(), reg)
	for {
		cmd, _, err := common.Decode(conn)
		if err != nil {
			glog.ERROR(err)
			break
		}

		switch cmd {
		case common.S2G_HEARTBEAT:
			glog.DEBUG("heartbeat from gtund ", conn.RemoteAddr().String())
			bytes, err := common.Encode(common.G2S_HEARTBEAT, nil)
			if err != nil {
				glog.ERROR(err)
				continue
			}

			conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			conn.Write(bytes)
			conn.SetWriteDeadline(time.Time{})
		default:
			glog.WARM("unimplemented cmd", cmd)

		}
	}
}
