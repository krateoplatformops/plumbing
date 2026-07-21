package sse

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestSSEHandler_Handle(t *testing.T) {
	sseHandler := NewSSEHandler()

	clientChan := make(chan string, 5)
	sseHandler.AddClient(clientChan)

	logger := slog.New(sseHandler)

	logger.Info("Test log", slog.String("key", "value"))

	select {
	case logMsg := <-clientChan:
		if !strings.Contains(logMsg, `"msg":"Test log"`) {
			t.Errorf("message not received: %s", logMsg)
		}
		if !strings.Contains(logMsg, `"key":"value"`) {
			t.Errorf("custom attribute is missing: %s", logMsg)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout")
	}

	sseHandler.RemoveClient(clientChan)
}

func TestSSEHandler_AddRemoveClient(t *testing.T) {
	sseHandler := NewSSEHandler()
	clientChan := make(chan string, 5)

	sseHandler.AddClient(clientChan)
	if len(sseHandler.clients) != 1 {
		t.Errorf("unable to register client")
	}

	sseHandler.RemoveClient(clientChan)
	if len(sseHandler.clients) != 0 {
		t.Errorf("unable to unregister client")
	}
}
