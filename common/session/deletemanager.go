package session

import (
	"sync"
	"time"
)

var sessionsMu sync.RWMutex
var sessions = make(map[string]*Session)

var deleteManager = &SessionDeleteManager{}

// 代表一个被kick的session
type sessionElement struct {
	next       *sessionElement
	session    *Session
	removeTime int64
}

// 管理被kick的session的删除，被kick的session需要延迟一段时间删除，以便提醒客户端该session被kick
// 被延迟删除的session构成一条根据时间排好序的链表
type SessionDeleteManager struct {
	first *sessionElement
	last  *sessionElement
	mu    sync.Mutex
}

func (sessionDeleteManager *SessionDeleteManager) doRemove() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			sessionDeleteManager.mu.Lock()
			first, last := sessionDeleteManager.first, sessionDeleteManager.last
			sessionDeleteManager.mu.Unlock()
			elem := first
			now := time.Now().Unix()
			removed := false
			for elem != nil {
				if elem.removeTime > now {
					break
				}
				removed = true
				sessionsMu.Lock()
				delete(sessions, elem.session.SessionId)
				sessionsMu.Unlock()
				if elem == last {
					break
				}
				elem = elem.next
			}
			if removed {
				sessionDeleteManager.mu.Lock()
				if elem != sessionDeleteManager.first {
					sessionDeleteManager.first = elem
				} else {
					if sessionDeleteManager.last == sessionDeleteManager.first {
						sessionDeleteManager.first = nil
						sessionDeleteManager.last = nil
					} else {
						sessionDeleteManager.first = sessionDeleteManager.first.next
					}
				}
				sessionDeleteManager.mu.Unlock()
			}
		}
	}
}

func (sessionDeleteManager *SessionDeleteManager) add(session *Session) {
	removeTime := time.Now().Add(time.Minute).Unix()
	elem := &sessionElement{
		session:    session,
		removeTime: removeTime,
	}
	sessionDeleteManager.mu.Lock()
	defer sessionDeleteManager.mu.Unlock()
	if sessionDeleteManager.first == nil {
		sessionDeleteManager.first = elem
	}
	if sessionDeleteManager.last != nil {
		sessionDeleteManager.last.next = elem
	}
	sessionDeleteManager.last = elem
}
