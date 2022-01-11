package gtun

import (
	"sync"

	"github.com/ICKelin/optw/transport"
)

var sessionMgr = &SessionManager{
	sessions: make(map[string][]*Session),
}

type SessionMap map[string]*Session

type SessionManager struct {
	rwmu     sync.RWMutex
	sessions map[string][]*Session
}

func GetSessionManager() *SessionManager {
	return sessionMgr
}

type Session struct {
	conn   transport.Conn
	region string
}

func newSession(conn transport.Conn, region string) *Session {
	return &Session{
		conn:   conn,
		region: region,
	}
}

func (mgr *SessionManager) AddSession(region string, sess *Session) {
	mgr.rwmu.Lock()
	defer mgr.rwmu.Unlock()
	regionSessions := mgr.sessions[region]
	if regionSessions == nil {
		regionSessions = make([]*Session, 0)
	}

	regionSessions = append(regionSessions, sess)
	mgr.sessions[region] = regionSessions
}

func (mgr *SessionManager) GetSession(region, dip string) *Session {
	mgr.rwmu.RLock()
	defer mgr.rwmu.RUnlock()

	regionSessions, ok := mgr.sessions[region]
	if !ok {
		return nil
	}

	if len(regionSessions) <= 0 {
		return nil
	}

	hash := 0
	for _, c := range dip {
		hash += int(c)
	}
	return regionSessions[hash%len(regionSessions)]
}

func (mgr *SessionManager) DeleteSession(region string, sess *Session) {
	mgr.rwmu.Lock()
	defer mgr.rwmu.Unlock()
	regionSessions := mgr.sessions[region]
	if regionSessions == nil {
		return
	}

	newSessions := make([]*Session, 0, len(regionSessions))
	for _, s := range regionSessions {
		if s.conn.RemoteAddr().String() == sess.conn.RemoteAddr().String() {
			continue
		}

		newSessions = append(newSessions, s)
	}

	mgr.sessions[region] = newSessions
}
