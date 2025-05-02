package sse

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type SSEHandler struct {
	mu      sync.Mutex
	clients map[chan string]struct{}
}

func NewSSEHandler() *SSEHandler {
	return &SSEHandler{
		clients: make(map[chan string]struct{}),
	}
}

func (h *SSEHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *SSEHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Formattare il log in JSON
	data := map[string]interface{}{
		"time":  r.Time.Format(time.RFC3339),
		"level": r.Level.String(),
		"msg":   r.Message,
	}

	// Estrarre gli attributi aggiuntivi dal record
	r.Attrs(func(a slog.Attr) bool {
		data[a.Key] = a.Value.Any()
		return true
	})

	jsonData, _ := json.Marshal(data)

	// Invia il log a tutti i client registrati
	for client := range h.clients {
		select {
		case client <- string(jsonData):
		default:
			// Se il canale è bloccato, rimuoviamo il client
			close(client)
			delete(h.clients, client)
		}
	}
	return nil
}

// WithAttrs e WithGroup implementano slog.Handler ma non fanno nulla in questo caso
func (h *SSEHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *SSEHandler) WithGroup(name string) slog.Handler       { return h }

// Aggiunge un nuovo client SSE
func (h *SSEHandler) AddClient(ch chan string) {
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
}

// Rimuove un client SSE
func (h *SSEHandler) RemoveClient(ch chan string) {
	h.mu.Lock()
	delete(h.clients, ch)
	close(ch)
	h.mu.Unlock()
}

// SSE endpoint per servire i log
func (h *SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Imposta gli header per SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Canale per inviare i log al client
	clientChan := make(chan string, 10)
	h.AddClient(clientChan)
	defer h.RemoveClient(clientChan)

	// Invia i log al client finché non chiude la connessione
	for log := range clientChan {
		_, err := w.Write([]byte("data: " + log + "\n\n"))
		if err != nil {
			break
		}
		w.(http.Flusher).Flush()
	}
}
