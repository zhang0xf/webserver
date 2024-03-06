package session

import (
	rnd "crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"nw/common/seqqueue"
)

type Session struct {
	SessionId   string
	SessionData interface{}
	UserData    interface{}
	kickStatus  int32
	SeqQueue    *seqqueue.SeqQueue
}

func (session *Session) Kick(needSync bool, kickType int) {
	session.kickStatus = int32(kickType)
	session.SeqQueue.Stop(needSync)
	deleteManager.add(session)
}

func (session *Session) IsKicked() bool {
	return session.kickStatus > 0
}

func init() {
	go deleteManager.doRemove()
}

func GenerateSessionId() (string, error) {
	k := make([]byte, 16)
	if _, err := io.ReadFull(rnd.Reader, k); err != nil {
		return "", nil
	}
	return hex.EncodeToString(k), nil
}

func New(sessionData interface{}, userData interface{}) (*Session, error) {
	sessionId, err := GenerateSessionId()
	if err != nil {
		return nil, err
	}
	queue := seqqueue.New(userData, sessionData.(int))
	session := &Session{
		SessionId:   sessionId,
		SessionData: sessionData,
		UserData:    userData,
		SeqQueue:    queue,
	}
	return session, nil
}

func Add(sessionId string, session *Session) {
	sessionsMu.Lock()
	sessions[sessionId] = session
	sessionsMu.Unlock()
}

func Remove(sessionId string) {
	session, err := Get(sessionId)
	if err == errors.New("session missed") {
		return
	}
	RemoveSession(session)
}

func RemoveSession(session *Session) {
	if !session.IsKicked() {
		session.SeqQueue.Stop(false)
		sessionsMu.Lock()
		delete(sessions, session.SessionId)
		sessionsMu.Unlock()
	}
}

func Get(sessionId string) (*Session, error) {
	sessionsMu.RLock()
	session, ok := sessions[sessionId]
	sessionsMu.RUnlock()
	if !ok {
		return nil, errors.New("session missed")
	}
	return session, nil
}
