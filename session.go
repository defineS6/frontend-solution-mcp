package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

type SessionStore struct {
	mu       sync.Mutex
	messages map[string][]ChatMessage
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		messages: make(map[string][]ChatMessage),
	}
}

func (s *SessionStore) MessagesForRequest(sessionID string, userMessage ChatMessage) (string, []ChatMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		generated, err := generateSessionID()
		if err != nil {
			return "", nil, err
		}
		sessionID = generated
	}

	history := cloneMessages(s.messages[sessionID])
	messages := append(history, userMessage)
	return sessionID, messages, nil
}

func (s *SessionStore) AppendExchange(sessionID string, userMessage ChatMessage, assistantMessage ChatMessage) []ChatMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	history := append(cloneMessages(s.messages[sessionID]), userMessage, assistantMessage)
	s.messages[sessionID] = history
	return cloneMessages(history)
}

func (s *SessionStore) History(sessionID string) []ChatMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	return cloneMessages(s.messages[sessionID])
}

func cloneMessages(messages []ChatMessage) []ChatMessage {
	if len(messages) == 0 {
		return nil
	}
	cloned := make([]ChatMessage, len(messages))
	copy(cloned, messages)
	return cloned
}

func generateSessionID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("生成会话 ID 失败：%w", err)
	}
	return "session-" + hex.EncodeToString(bytes[:]), nil
}
