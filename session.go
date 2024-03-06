package nw

import (
	"math/rand"
	"sync"
	"time"
)

// Session the session interface to bind to a Conn
type Session interface {
	GetId() uint32                 // an unique id must be assigned for each Session
	GetConn() Conn                 // get corresponding Conn
	OnOpen(conn Conn)              // called when Conn is opened
	OnClose(conn Conn)             // called when Conn is closed
	OnRecv(conn Conn, data []byte) // called when Conn receives data, ATTENTION: data is a slice and must be consumed immediately
}

// SessionManager manage all the sessions for a server network service
type SessionManager interface {
	AddSession(id uint32, session Session)
	RemoveSession(id uint32)
	GetSession(id uint32) Session
	GetSessionCount() int
	Clear()
	Range(f func(id uint32, session Session) bool)
}

type DefaultSessionManager struct {
	sessions   map[uint32]Session // key may not be equal to session.GetId(), let application layer decide
	sessionsMu sync.RWMutex
	safeMode   bool
}

func NewDefaultSessionManager(safeMode bool) *DefaultSessionManager {
	return &DefaultSessionManager{
		sessions: make(map[uint32]Session),
		safeMode: safeMode,
	}
}

func (this *DefaultSessionManager) AddSession(id uint32, session Session) {
	if !this.safeMode {
		this.sessions[id] = session
		return
	}
	this.sessionsMu.Lock()
	this.sessions[id] = session
	this.sessionsMu.Unlock()
}

func (this *DefaultSessionManager) RemoveSession(id uint32) {
	if !this.safeMode {
		delete(this.sessions, id)
		return
	}
	this.sessionsMu.Lock()
	delete(this.sessions, id)
	this.sessionsMu.Unlock()
}

func (this *DefaultSessionManager) GetSession(id uint32) Session {
	if !this.safeMode {
		return this.sessions[id]
	}
	this.sessionsMu.RLock()
	session := this.sessions[id]
	this.sessionsMu.RUnlock()
	return session
}

func (this *DefaultSessionManager) GetRoundSession() (Session, bool) {
	var keys = make([]uint32, 0)
	this.sessionsMu.RLock()
	for key := range this.sessions {
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		this.sessionsMu.RUnlock()
		return nil, false
	}
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	index := r.Intn(len(keys))
	session := this.sessions[keys[index]]
	this.sessionsMu.RUnlock()
	return session, true
}

func (this *DefaultSessionManager) Range(f func(id uint32, session Session) bool) {
	if this.safeMode {
		this.sessionsMu.RLock()
	}
	for id, session := range this.sessions {
		if f(id, session) {
			break
		}
	}
	if this.safeMode {
		this.sessionsMu.RUnlock()
	}
}

func (this *DefaultSessionManager) GetSessionCount() int {
	if this.safeMode {
		this.sessionsMu.RLock()
	}
	count := len(this.sessions)
	if this.safeMode {
		this.sessionsMu.RUnlock()
	}
	return count
}

func (this *DefaultSessionManager) Clear() {
	if this.safeMode {
		this.sessionsMu.RLock()
	}
	for id := range this.sessions {
		delete(this.sessions, id)
	}
	if this.safeMode {
		this.sessionsMu.RUnlock()
	}
}

func (this *DefaultSessionManager) GetAllSessionId() []uint32 {
	this.sessionsMu.RLock()
	ids := make([]uint32, 0)
	for k := range this.sessions {
		ids = append(ids, k)
	}
	this.sessionsMu.RUnlock()
	return ids
}
