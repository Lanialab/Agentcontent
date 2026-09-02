package httpapi

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/grokbot-2/agentcontent/internal/ports"
)

type Hub struct {
	mu   sync.Mutex
	subs map[chan ports.TraceEvent]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: map[chan ports.TraceEvent]struct{}{}}
}

func (h *Hub) Emit(ev ports.TraceEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}

func (h *Hub) Subscribe() chan ports.TraceEvent {
	ch := make(chan ports.TraceEvent, 16)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(ch chan ports.TraceEvent) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
}

type envelope struct {
	Type  string          `json:"type"`
	At    string          `json:"at"`
	Path  string          `json:"path"`
	Kind  ports.TraceKind `json:"kind"`
	Label string          `json:"label"`
	Nodes []string        `json:"nodes"`
}

func encodeEvent(ev ports.TraceEvent) []byte {
	b, _ := json.Marshal(envelope{
		Type:  "pulse",
		At:    time.Now().UTC().Format(time.RFC3339Nano),
		Path:  ev.Path,
		Kind:  ev.Kind,
		Label: ev.Label,
		Nodes: ev.Nodes,
	})
	return b
}
