package gtun

import (
	"sync"

	"github.com/xtaci/smux"
)

var sessionMgr = &SessionManager{}

// SessionManager defines the session info add/delete/get actions
type SessionManager struct {
	sessions sync.Map
}

// GetSessionManager returs the singleton of session manager
func GetSessionManager() *SessionManager {
	return sessionMgr
}

// Session defines each opennotr_client to opennotr_server connection
type Session struct {
	conn    *smux.Session
	region  string
	rxbytes uint64
	txbytes uint64
}

func newSession(conn *smux.Session, region string) *Session {
	return &Session{
		conn:   conn,
		region: region,
	}
}

func (mgr *SessionManager) AddSession(vip string, sess *Session) {
	mgr.sessions.Store(vip, sess)
}

func (mgr *SessionManager) GetSession(vip string) *Session {
	val, ok := mgr.sessions.Load(vip)
	if !ok {
		return nil
	}
	return val.(*Session)
}

func (mgr *SessionManager) DeleteSession(vip string) {
	mgr.sessions.Delete(vip)
}
