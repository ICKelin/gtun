package gtun

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
	"github.com/songgao/water"
)

type ClientConfig struct {
	serverAddr string
	authKey    string
}

type Client struct {
	serverAddr string
	authKey    string
	myip       string
	gw         string
	god        *God
}

func NewClient(cfg *ClientConfig) *Client {
	return &Client{
		serverAddr: cfg.serverAddr,
		authKey:    cfg.authKey,
		myip:       "",
		god:        NewGod(&GodConfig{}), // everything use default
	}
}

func (client *Client) Run(opts *Options) {
	for {
		server, err := client.god.Access()
		if err != nil && client.god.must {
			glog.ERROR("get server address fail: ", err)
			time.Sleep(time.Second * 3)
			continue
		}

		if server == "" {
			server = client.serverAddr
		}

		if server == "" {
			glog.ERROR("empty server")
			time.Sleep(time.Second * 3)
			continue
		}

		conn, err := conServer(server)
		if err != nil {
			glog.ERROR("connect to server fail: ", err)
			time.Sleep(time.Second * 3)
			continue
		}

		s2c, err := authorize(conn, client.authKey)
		if err != nil {
			glog.ERROR("auth fail: ", err)
			time.Sleep(time.Second * 3)
			continue
		}

		glog.INFO("connect to", server, "success", s2c.AccessIP)

		ifce, err := NewIfce(opts.tap)
		if err != nil {
			glog.ERROR(err)
			return
		}

		client.myip = s2c.AccessIP
		client.gw = s2c.Gateway
		wg := &sync.WaitGroup{}
		wg.Add(3)

		sndqueue := make(chan []byte)
		go ifaceRead(ifce, sndqueue)

		err = setupIface(ifce, s2c.AccessIP, s2c.Gateway)
		if err != nil {
			glog.ERROR(err)
			time.Sleep(time.Second * 3)
			continue
		}

		go func() {
			routes, err := downloadRoutes(s2c.RouteScriptUrl)
			if err != nil {
				glog.WARM(err)
			}
			insertRoute(routes, s2c.AccessIP, s2c.Gateway, ifce.Name())
		}()

		done := make(chan struct{})
		go heartbeat(sndqueue, done, wg)
		go snd(conn, sndqueue, done, wg)
		go rcv(conn, ifce, wg)

		wg.Wait()

		ifce.Close()
		glog.INFO("reconnecting")
	}
}

func (client *Client) needreload(nip string) bool {
	return client.myip == "" || client.myip != nip
}

func conServer(srv string) (conn net.Conn, err error) {
	tcp, err := net.DialTimeout("tcp", srv, time.Second*10)
	if err != nil {
		return nil, err
	}

	return tcp, nil
}

func authorize(conn net.Conn, key string) (s2cauthorize *common.S2CAuthorize, err error) {
	c2sauthorize := &common.C2SAuthorize{
		OS:       common.OSID(runtime.GOOS),
		Version:  common.Version(),
		AccessIP: "",
		Key:      key,
	}

	payload, err := json.Marshal(c2sauthorize)
	if err != nil {
		return nil, err
	}

	buff, _ := common.Encode(common.C2S_AUTHORIZE, payload)

	_, err = conn.Write(buff)
	if err != nil {
		return nil, err
	}

	cmd, resp, err := common.Decode(conn)
	if err != nil {
		return nil, err
	}

	if cmd != common.S2C_AUTHORIZE {
		err = fmt.Errorf("invalid authorize cmd %d", cmd)
		return nil, err
	}

	s2cauthorize = &common.S2CAuthorize{}
	err = json.Unmarshal(resp, s2cauthorize)
	if err != nil {
		return nil, err
	}

	return s2cauthorize, nil
}

func rcv(conn net.Conn, ifce *water.Interface, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	for {
		cmd, pkt, err := common.Decode(conn)
		if err != nil {
			glog.INFO(err)
			break
		}
		switch cmd {
		case common.S2C_HEARTBEAT:
			glog.DEBUG("heartbeat from srv")

		case common.C2C_DATA:
			_, err := ifce.Write(pkt)
			if err != nil {
				glog.ERROR(err)
			}

		default:
			glog.INFO("unimplement cmd", int(cmd), pkt)
		}
	}
}

func snd(conn net.Conn, sndqueue chan []byte, done chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()
	defer close(done)

	for {
		pkt := <-sndqueue
		conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
		_, err := conn.Write(pkt)
		conn.SetWriteDeadline(time.Time{})
		if err != nil {
			glog.ERROR(err)
			break
		}
	}
}

func heartbeat(sndqueue chan []byte, done chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-done:
			return

		case <-time.After(time.Second * 3):
			bytes, _ := common.Encode(common.C2S_HEARTBEAT, nil)
			sndqueue <- bytes
		}
	}
}

func ifaceRead(ifce *water.Interface, sndqueue chan []byte) {
	packet := make([]byte, 65536)
	for {
		n, err := ifce.Read(packet)
		if err != nil {
			glog.ERROR(err)
			break
		}

		bytes, _ := common.Encode(common.C2C_DATA, packet[:n])
		sndqueue <- bytes
	}
}

func clearIfConfig(ifce *water.Interface, ip string, gw string) {
	switch runtime.GOOS {
	case "linux":
		args := strings.Split(fmt.Sprintf("addr del %s/24 dev %s", ip, ifce.Name()), " ")
		glog.DEBUG(exec.Command("ip", args...).CombinedOutput())

	case "darwin":

	case "windows":
	}
}

func setupIface(ifce *water.Interface, ip string, gw string) (err error) {
	type CMD struct {
		cmd  string
		args []string
	}

	cmdlist := make([]*CMD, 0)

	switch runtime.GOOS {
	case "linux":
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{ifce.Name(), "up"}})
		args := strings.Split(fmt.Sprintf("addr add %s/24 dev %s", ip, ifce.Name()), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ip", args: args})

	case "darwin":
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{ifce.Name(), "up"}})

		args := strings.Split(fmt.Sprintf("%s %s %s", ifce.Name(), ip, ip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: args})

		args = strings.Split(fmt.Sprintf("add -net %s/24 %s", gw, ip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "route", args: args})

	case "windows":
		args := strings.Split(fmt.Sprintf("interface ip set address name=\"%s\" addr=%s source=static mask=255.255.255.0 gateway=%s", ifce.Name(), ip, gw), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "netsh", args: args})

		args = strings.Split(fmt.Sprintf("delete 0.0.0.0 %s", gw), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "route", args: args})
	}

	for _, c := range cmdlist {
		output, err := exec.Command(c.cmd, c.args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("run %s error %s", c, string(output))
		}
	}

	return nil
}

func releaseDevice(device, ip, gateway string) (err error) {
	type CMD struct {
		cmd  string
		args []string
	}

	cmdlist := make([]*CMD, 0)

	switch runtime.GOOS {
	case "linux":
		args := strings.Split(fmt.Sprintf("%s down", device), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: args})

	case "darwin":
		gw := strings.Split(gateway, ".")
		if len(gw) != 4 {
			break
		}

		s := strings.Join(gw[:3], ".")
		args := strings.Split(fmt.Sprintf("delete -net %s/24 %s", s, ip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "route", args: args})
		args = strings.Split(fmt.Sprintf("%s delete %s", device, ip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: args})
	}

	for _, c := range cmdlist {
		output, _ := exec.Command(c.cmd, c.args...).CombinedOutput()
		if err != nil {
			fmt.Printf("run %s error %s\n", c, string(output))
		}
	}

	return nil
}

func downloadRoutes(url string) ([]string, error) {
	routes := make([]string, 0)

	glog.INFO("downloading route file from:", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		// may need to validate ip/cidr format
		routes = append(routes, string(line))
	}
	glog.INFO("downloaded route file from:", url)
	return routes, nil
}

func insertRoute(routedIPS []string, devIP, gw string, devName string) {
	// Windows platform route add need iface index args.
	ifceIndex := -1
	ifce, err := net.InterfaceByName(devName)
	if err != nil {
		if runtime.GOOS == "windows" {
			return
		}
	} else {
		ifceIndex = ifce.Index
	}
	glog.INFO("inserting routes")
	for _, address := range routedIPS {
		execRoute(address, devName, devIP, gw, ifceIndex)
	}
	glog.INFO("inserted routes, routes count:", len(routedIPS))
}

type CMD struct {
	cmd  string
	args []string
}

func execRoute(address, device, tunip, gateway string, ifceIndex int) {
	cmd := &CMD{}

	switch runtime.GOOS {
	case "linux":
		args := strings.Split(fmt.Sprintf("ro add %s dev %s", address, device), " ")
		cmd = &CMD{cmd: "ip", args: args}

	case "darwin":
		args := strings.Split(fmt.Sprintf("add -net %s %s", address, tunip), " ")
		cmd = &CMD{cmd: "route", args: args}

	case "windows":
		args := strings.Split(fmt.Sprintf("add %s %s if %d", address, gateway, ifceIndex), " ")
		cmd = &CMD{cmd: "route", args: args}

	default:
		return
	}

	output, err := exec.Command(cmd.cmd, cmd.args...).CombinedOutput()
	if err != nil {
		glog.DEBUG("add", address, "fail:", string(output))
	}
	glog.DEBUG(string(output))
}
