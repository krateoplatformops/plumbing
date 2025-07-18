package event

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	record "k8s.io/client-go/tools/events"
)

// Mock object for testing
type mockObject struct {
	runtime.Object
}

func TestTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		actual   Type
		expected string
	}{
		{
			name:     "TypeNormal",
			actual:   TypeNormal,
			expected: "Normal",
		},
		{
			name:     "TypeWarning",
			actual:   TypeWarning,
			expected: "Warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.actual))
		})
	}
}

func TestWithRelated(t *testing.T) {
	relatedObj := &mockObject{}

	event := Event{
		Type:    TypeNormal,
		Reason:  "TestReason",
		Message: "Test message",
		Action:  "TestAction",
	}

	// Apply the option
	opt := WithRelated(relatedObj)
	opt(&event)

	assert.Equal(t, relatedObj, event.Related)
}

func TestNormal(t *testing.T) {
	tests := []struct {
		name     string
		reason   Reason
		action   Action
		message  string
		opts     []EventOption
		expected Event
	}{
		{
			name:    "Normal event without options",
			reason:  "TestReason",
			action:  "TestAction",
			message: "Test message",
			opts:    nil,
			expected: Event{
				Type:    TypeNormal,
				Reason:  "TestReason",
				Action:  "TestAction",
				Message: "Test message",
				Related: nil,
			},
		},
		{
			name:    "Normal event with related object",
			reason:  "CreateSuccess",
			action:  "Create",
			message: "Successfully created resource",
			opts:    []EventOption{WithRelated(&mockObject{})},
			expected: Event{
				Type:    TypeNormal,
				Reason:  "CreateSuccess",
				Action:  "Create",
				Message: "Successfully created resource",
				Related: &mockObject{},
			},
		},
		{
			name:    "Normal event with empty message",
			reason:  "EmptyMessage",
			action:  "Test",
			message: "",
			opts:    nil,
			expected: Event{
				Type:    TypeNormal,
				Reason:  "EmptyMessage",
				Action:  "Test",
				Message: "",
				Related: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normal(tt.reason, tt.action, tt.message, tt.opts...)

			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Reason, result.Reason)
			assert.Equal(t, tt.expected.Action, result.Action)
			assert.Equal(t, tt.expected.Message, result.Message)

			if tt.expected.Related == nil {
				assert.Nil(t, result.Related)
			} else {
				assert.NotNil(t, result.Related)
				assert.IsType(t, tt.expected.Related, result.Related)
			}
		})
	}
}

func TestWarning(t *testing.T) {
	tests := []struct {
		name     string
		reason   Reason
		action   Action
		err      error
		opts     []EventOption
		expected Event
	}{
		{
			name:   "Warning event without options",
			reason: "FailedCreate",
			action: "Create",
			err:    errors.New("creation failed"),
			opts:   nil,
			expected: Event{
				Type:    TypeWarning,
				Reason:  "FailedCreate",
				Action:  "Create",
				Message: "creation failed",
				Related: nil,
			},
		},
		{
			name:   "Warning event with related object",
			reason: "ValidationError",
			action: "Validate",
			err:    errors.New("validation failed: missing field"),
			opts:   []EventOption{WithRelated(&mockObject{})},
			expected: Event{
				Type:    TypeWarning,
				Reason:  "ValidationError",
				Action:  "Validate",
				Message: "validation failed: missing field",
				Related: &mockObject{},
			},
		},
		{
			name:   "Warning event with empty error message",
			reason: "Unknown",
			action: "Process",
			err:    errors.New(""),
			opts:   nil,
			expected: Event{
				Type:    TypeWarning,
				Reason:  "Unknown",
				Action:  "Process",
				Message: "",
				Related: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Warning(tt.reason, tt.action, tt.err, tt.opts...)

			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Reason, result.Reason)
			assert.Equal(t, tt.expected.Action, result.Action)
			assert.Equal(t, tt.expected.Message, result.Message)

			if tt.expected.Related == nil {
				assert.Nil(t, result.Related)
			} else {
				assert.NotNil(t, result.Related)
				assert.IsType(t, tt.expected.Related, result.Related)
			}
		})
	}
}

func TestNewAPIRecorder(t *testing.T) {
	tests := []struct {
		name          string
		eventRecorder record.EventRecorder
		expectedNil   bool
	}{
		{
			name:          "Valid EventRecorder",
			eventRecorder: record.NewFakeRecorder(100),
			expectedNil:   false,
		},
		{
			name:          "Nil EventRecorder",
			eventRecorder: nil,
			expectedNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewAPIRecorder(tt.eventRecorder)

			if tt.expectedNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.eventRecorder, result.kube)
			}
		})
	}
}

func TestAPIRecorder_Event(t *testing.T) {
	tests := []struct {
		name  string
		obj   runtime.Object
		event Event
	}{
		{
			name: "Normal event without related object",
			obj:  &mockObject{},
			event: Event{
				Type:    TypeNormal,
				Reason:  "TestReason",
				Action:  "TestAction",
				Message: "Test message",
				Related: nil,
			},
		},
		{
			name: "Warning event with related object",
			obj:  &mockObject{},
			event: Event{
				Type:    TypeWarning,
				Reason:  "FailedOperation",
				Action:  "Update",
				Message: "Operation failed",
				Related: &mockObject{},
			},
		},
		{
			name: "Event with empty message",
			obj:  &mockObject{},
			event: Event{
				Type:    TypeNormal,
				Reason:  "EmptyMessage",
				Action:  "Test",
				Message: "",
				Related: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRec := record.NewFakeRecorder(100)
			recorder := NewAPIRecorder(mockRec)

			// Call the method
			recorder.Event(tt.obj, tt.event)
		})
	}
}

func TestNewNopRecorder(t *testing.T) {
	recorder := NewNopRecorder()

	assert.NotNil(t, recorder)
	assert.IsType(t, &NopRecorder{}, recorder)
}

func TestNopRecorder_Event(t *testing.T) {
	recorder := NewNopRecorder()
	obj := &mockObject{}
	event := Event{
		Type:    TypeNormal,
		Reason:  "TestReason",
		Action:  "TestAction",
		Message: "Test message",
	}

	// This should not panic or cause any issues
	assert.NotPanics(t, func() {
		recorder.Event(obj, event)
	})
}

func TestEventOptionMultiple(t *testing.T) {
	relatedObj1 := &mockObject{}
	relatedObj2 := &mockObject{}

	// Test applying multiple options (last one should win for Related)
	event := Event{
		Type:    TypeNormal,
		Reason:  "TestReason",
		Message: "Test message",
		Action:  "TestAction",
	}

	opt1 := WithRelated(relatedObj1)
	opt2 := WithRelated(relatedObj2)

	opt1(&event)
	opt2(&event)

	assert.Equal(t, relatedObj2, event.Related)
}

func TestEventTypes(t *testing.T) {
	// Test that custom types can be created
	customReason := Reason("CustomReason")
	customAction := Action("CustomAction")

	event := Normal(customReason, customAction, "Custom message")

	assert.Equal(t, TypeNormal, event.Type)
	assert.Equal(t, customReason, event.Reason)
	assert.Equal(t, customAction, event.Action)
	assert.Equal(t, "Custom message", event.Message)
}

func TestRecorderInterface(t *testing.T) {
	// Test that both APIRecorder and NopRecorder implement Recorder interface
	var recorder Recorder

	// Test APIRecorder
	mockRec := record.NewFakeRecorder(100)
	apiRecorder := NewAPIRecorder(mockRec)
	recorder = apiRecorder
	assert.NotNil(t, recorder)

	// Test NopRecorder
	nopRecorder := NewNopRecorder()
	recorder = nopRecorder
	assert.NotNil(t, recorder)
}
