package eventrecorder

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type fakeEventRecorder struct {
	calls []fakeCall
}

type fakeCall struct {
	regarding runtime.Object
	related   runtime.Object
	eventType string
	reason    string
	action    string
	note      string
}

func (f *fakeEventRecorder) Eventf(regarding runtime.Object, related runtime.Object, eventtype, reason, action, note string, args ...interface{}) {
	f.calls = append(f.calls, fakeCall{
		regarding: regarding,
		related:   related,
		eventType: eventtype,
		reason:    reason,
		action:    action,
		note:      fmt.Sprintf(note, args...),
	})
}

func TestThrottledRecorder_SuppressesUnchangedEvent(t *testing.T) {
	inner := &fakeEventRecorder{}
	r := NewStateAwareRecorder(inner)

	obj := &metav1.PartialObjectMetadata{}
	obj.SetNamespace("default")
	obj.SetName("demo")
	obj.SetUID("uid-1")

	r.Eventf(obj, nil, "Normal", "Observe", "Observe", "state is %s", "ready")
	r.Eventf(obj, nil, "Normal", "Observe", "Observe", "state is %s", "ready")

	if got := len(inner.calls); got != 1 {
		t.Fatalf("expected 1 published event, got %d", got)
	}
}

func TestThrottledRecorder_PublishesWhenReasonChanges(t *testing.T) {
	inner := &fakeEventRecorder{}
	r := NewStateAwareRecorder(inner)

	obj := &metav1.PartialObjectMetadata{}
	obj.SetNamespace("default")
	obj.SetName("demo")
	obj.SetUID("uid-1")

	r.Eventf(obj, nil, "Normal", "Observe", "Observe", "state is ready")
	r.Eventf(obj, nil, "Normal", "Updated", "Observe", "state is ready")

	if got := len(inner.calls); got != 2 {
		t.Fatalf("expected 2 published events, got %d", got)
	}
}

func TestThrottledRecorder_PublishesWhenMessageChanges(t *testing.T) {
	inner := &fakeEventRecorder{}
	r := NewStateAwareRecorder(inner)

	obj := &metav1.PartialObjectMetadata{}
	obj.SetNamespace("default")
	obj.SetName("demo")
	obj.SetUID("uid-1")

	r.Eventf(obj, nil, "Warning", "ReconcileFailed", "Reconcile", "error: %s", "timeout")
	r.Eventf(obj, nil, "Warning", "ReconcileFailed", "Reconcile", "error: %s", "forbidden")

	if got := len(inner.calls); got != 2 {
		t.Fatalf("expected 2 published events, got %d", got)
	}
}

func TestThrottledRecorder_ForgetForgetsState(t *testing.T) {
	inner := &fakeEventRecorder{}
	r := NewStateAwareRecorder(inner)

	obj := &metav1.PartialObjectMetadata{}
	obj.SetNamespace("default")
	obj.SetName("demo")
	obj.SetUID("uid-1")

	r.Eventf(obj, nil, "Normal", "Observe", "Observe", "state is ready")
	r.Eventf(obj, nil, "Normal", "Observe", "Observe", "state is ready")
	r.Forget(obj)
	r.Eventf(obj, nil, "Normal", "Observe", "Observe", "state is ready")

	if got := len(inner.calls); got != 2 {
		t.Fatalf("expected 2 published events, got %d", got)
	}
}
