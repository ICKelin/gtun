package gtun

import (
	"sync"

	"github.com/ICKelin/gtun/transport"
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

func (mgr *SessionManager) AddSession(region string, sess transport.Session) {
	mgr.sessions.Store(region, sess)
}

func (mgr *SessionManager) GetSession(region string) transport.Session {
	val, ok := mgr.sessions.Load(region)
	if !ok {
		return nil
	}
	return val.(transport.Session)
}

func (mgr *SessionManager) DeleteSession(region string) {
	mgr.sessions.Delete(region)
}
