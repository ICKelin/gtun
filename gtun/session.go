package gtun

import (
	"sync"

	"github.com/hashicorp/yamux"
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
	conn   *yamux.Session
	region string
}

func newSession(conn *yamux.Session, region string) *Session {
	return &Session{
		conn:   conn,
		region: region,
	}
}

func (mgr *SessionManager) AddSession(region string, sess *Session) {
	mgr.sessions.Store(region, sess)
}

func (mgr *SessionManager) GetSession(region string) *Session {
	val, ok := mgr.sessions.Load(region)
	if !ok {
		return nil
	}
	return val.(*Session)
}

func (mgr *SessionManager) DeleteSession(region string) {
	mgr.sessions.Delete(region)
}
