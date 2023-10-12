package chatbot

import (
	"time"

	"github.com/sashabaranov/go-openai"
)

type SessionManager struct {
	DefaultRole string
	Sessions    map[string]*Session
}

func NewSessionManager(defaultRole string) *SessionManager {
	sessMgr := &SessionManager{}
	sessMgr.DefaultRole = defaultRole
	sessMgr.Sessions = make(map[string]*Session)
	return sessMgr
}

func (m *SessionManager) GetSession(id string) *Session {
	if s, ok := m.Sessions[id]; ok {
		return s
	}

	s := NewSession(id, m.DefaultRole)
	m.Sessions[id] = s
	return s
}

func (m *SessionManager) ClearExpiredSessions(expiryPeriod time.Duration) []string {
	now := time.Now()
	ids := make([]string, 0)
	for _, s := range m.Sessions {
		if len(s.Messages) == 0 {
			continue
		}

		if now.Sub(s.LastUpdateDate) > expiryPeriod {
			s.Clear()
			ids = append(ids, s.ID)
		}
	}
	return ids
}

type Session struct {
	ID             string
	Role           string
	Messages       []openai.ChatCompletionMessage
	LastUpdateDate time.Time
}

func NewSession(id, role string) *Session {
	s := &Session{}
	s.Messages = make([]openai.ChatCompletionMessage, 0, 10)
	s.LastUpdateDate = time.Now()
	s.ID = id
	s.Role = role
	return s
}

func (s *Session) Clear() {
	s.Messages = s.Messages[:0]
	s.LastUpdateDate = time.Now()
}

func (s *Session) ShortID() string {
	const _SHORT_ID_LEN = 12
	if len(s.ID) <= _SHORT_ID_LEN {
		return s.ID
	}

	return s.ID[:_SHORT_ID_LEN]
}

func (s *Session) AddMessage(msg *openai.ChatCompletionMessage) {
	s.Messages = append(s.Messages, *msg)
	s.LastUpdateDate = time.Now()
}

func (s *Session) ChangeRole(role string) {
	s.Clear()
	s.Role = role
}
