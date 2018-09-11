package gtun

type tuple struct {
	proto   uint8
	srcPort uint16
	dstPort uint16
	srcIP   uint32
	dstIP   uint32
}

type session struct {
	origin tuple
	dst    tuple
}

type SessionManager struct {
}

func NewSessionManager() *SessionManager {
	return &SessionManager{}
}

func (ct *SessionManager) Get() {
	// TODO
}

func (ct *SessionManager) Set() {
	// TODO
}
