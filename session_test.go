package main

import (
	"strings"
	"testing"
)

func TestSessionStoreCreatesAndContinuesSession(t *testing.T) {
	store := NewSessionStore()

	userA := ChatMessage{Role: "user", Content: "第一个需求"}
	sessionID, messages, err := store.MessagesForRequest("", userA)
	if err != nil {
		t.Fatalf("MessagesForRequest() error = %v", err)
	}
	if !strings.HasPrefix(sessionID, "session-") {
		t.Fatalf("sessionID = %q", sessionID)
	}
	if len(messages) != 1 || messages[0] != userA {
		t.Fatalf("messages = %#v", messages)
	}

	assistantA := ChatMessage{Role: "assistant", Content: "第一个方案"}
	history := store.AppendExchange(sessionID, userA, assistantA)
	if len(history) != 2 {
		t.Fatalf("history len = %d", len(history))
	}

	userB := ChatMessage{Role: "user", Content: "继续细化"}
	nextSessionID, nextMessages, err := store.MessagesForRequest(sessionID, userB)
	if err != nil {
		t.Fatalf("MessagesForRequest() error = %v", err)
	}
	if nextSessionID != sessionID {
		t.Fatalf("nextSessionID = %q, want %q", nextSessionID, sessionID)
	}
	if len(nextMessages) != 3 {
		t.Fatalf("nextMessages len = %d", len(nextMessages))
	}
	if nextMessages[0] != userA || nextMessages[1] != assistantA || nextMessages[2] != userB {
		t.Fatalf("nextMessages = %#v", nextMessages)
	}
}
