package gtun

import (
	"strings"
	"sync"

	"github.com/ICKelin/gtun/internal/logs"

	"github.com/ICKelin/optw/transport"
)

var sessionMgr = &SessionManager{
	sessions: make(map[string][]*Session),
}

type SessionMap map[string]*Session

type SessionManager struct {
	rwmu        sync.RWMutex
	sessions    map[string][]*Session
	raceManager *RaceManager
}

func GetSessionManager() *SessionManager {
	return sessionMgr
}

func (sessionMgr *SessionManager) SetRaceManager(raceManager *RaceManager) {
	sessionMgr.raceManager = raceManager
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

	// get bestNode from race module
	bestNode := mgr.raceManager.GetBestNode(region)
	bestIP := strings.Split(bestNode, ":")[0]
	for i := 0; i < len(regionSessions); i++ {
		sess := regionSessions[i]
		if sess.conn.IsClosed() {
			logs.Warn("%s %s is closed", region, sess.conn.RemoteAddr())
			continue
		}

		if len(bestIP) != 0 {
			sessIP := strings.Split(sess.conn.RemoteAddr().String(), ":")[0]
			if bestIP == sessIP {
				logs.Debug("best ip match %s", bestIP)
				return sess
			}
		}
	}

	logs.Warn("use random session")
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
