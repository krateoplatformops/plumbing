package eventrecorder

import (
	"fmt"
	"hash/fnv"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	record "k8s.io/client-go/tools/events"
)

// ThrottledRecorder publishes an event only when its effective state changes
// for the same regarding/related object stream.
type ThrottledRecorder struct {
	mu         sync.Mutex
	inner      record.EventRecorder
	signatures map[string]uint64
}

var _ record.EventRecorder = (*ThrottledRecorder)(nil)

// NewStateAwareRecorder returns a recorder that suppresses duplicate events.
func NewStateAwareRecorder(inner record.EventRecorder) *ThrottledRecorder {
	return &ThrottledRecorder{
		inner:      inner,
		signatures: make(map[string]uint64),
	}
}

// NewThrottledRecorder is kept for backward compatibility.
// publishN is ignored because deduplication is state-based, not count-based.
func NewThrottledRecorder(inner record.EventRecorder, _ int) *ThrottledRecorder {
	return NewStateAwareRecorder(inner)
}

func (t *ThrottledRecorder) Eventf(regarding runtime.Object, related runtime.Object, eventType, reason, action, note string, args ...interface{}) {
	if t == nil || t.inner == nil {
		return
	}

	message := fmt.Sprintf(note, args...)
	key := eventKey(regarding, related, eventType, reason, action)
	sig := hashString(message)

	t.mu.Lock()
	prev, ok := t.signatures[key]
	if ok && prev == sig {
		t.mu.Unlock()
		return
	}
	t.signatures[key] = sig
	t.mu.Unlock()

	t.inner.Eventf(regarding, related, eventType, reason, action, "%s", message)
}

// Reset clears all tracked signatures.
func (t *ThrottledRecorder) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.signatures = make(map[string]uint64)
}

// Forget removes tracked signatures for a specific regarding object.
func (t *ThrottledRecorder) Forget(regarding runtime.Object) {
	prefix := objectKey(regarding) + "|"

	t.mu.Lock()
	defer t.mu.Unlock()

	for k := range t.signatures {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(t.signatures, k)
		}
	}
}

func eventKey(regarding runtime.Object, related runtime.Object, eventType, reason, action string) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s", objectKey(regarding), objectKey(related), eventType, reason, action)
}

func objectKey(obj runtime.Object) string {
	if obj == nil {
		return "<nil>"
	}

	gvk := obj.GetObjectKind().GroupVersionKind().String()
	if accessor, err := meta.Accessor(obj); err == nil {
		return fmt.Sprintf("%T|%s|%s|%s|%s", obj, gvk, accessor.GetNamespace(), accessor.GetName(), accessor.GetUID())
	}

	return fmt.Sprintf("%T|%s", obj, gvk)
}

func hashString(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}
