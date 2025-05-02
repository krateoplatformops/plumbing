package env

import (
	"os"
	"testing"
	"time"
)

func TestString(t *testing.T) {
	key := "TEST_STRING"
	defaultValue := "default"

	result := String(key, defaultValue)
	if result != defaultValue {
		t.Errorf("expected %q, got %q", defaultValue, result)
	}

	os.Setenv(key, "custom")
	defer os.Unsetenv(key)

	result = String(key, defaultValue)
	if result != "custom" {
		t.Errorf("expected %q, got %q", "custom", result)
	}
}

func TestInt(t *testing.T) {
	key := "TEST_INT"
	defaultValue := 42

	result := Int(key, defaultValue)
	if result != defaultValue {
		t.Errorf("expected %d, got %d", defaultValue, result)
	}

	os.Setenv(key, "100")
	defer os.Unsetenv(key)

	result = Int(key, defaultValue)
	if result != 100 {
		t.Errorf("expected 100, got %d", result)
	}

	os.Setenv(key, "invalid")
	result = Int(key, defaultValue)
	if result != defaultValue {
		t.Errorf("expected %d, got %d", defaultValue, result)
	}
}

func TestBool(t *testing.T) {
	key := "TEST_BOOL"

	if Bool(key, true) != true {
		t.Errorf("expected true, got false")
	}

	os.Setenv(key, "true")
	defer os.Unsetenv(key)

	if Bool(key, false) != true {
		t.Errorf("expected true, got false")
	}

	os.Setenv(key, "false")
	if Bool(key, true) != false {
		t.Errorf("expected false, got true")
	}

	os.Setenv(key, "invalid")
	if Bool(key, true) != true {
		t.Errorf("expected default true, got false")
	}
}

func TestDuration(t *testing.T) {
	key := "TEST_DURATION"
	defaultValue := 5 * time.Second

	if Duration(key, defaultValue) != defaultValue {
		t.Errorf("expected %v, got %v", defaultValue, Duration(key, defaultValue))
	}

	os.Setenv(key, "10s")
	defer os.Unsetenv(key)

	if Duration(key, defaultValue) != 10*time.Second {
		t.Errorf("expected 10s, got %v", Duration(key, defaultValue))
	}

	os.Setenv(key, "invalid")
	if Duration(key, defaultValue) != defaultValue {
		t.Errorf("expected %v, got %v", defaultValue, Duration(key, defaultValue))
	}
}

func TestServicePort(t *testing.T) {
	key := "TEST_PORT"
	defaultValue := 8080

	if ServicePort(key, defaultValue) != defaultValue {
		t.Errorf("expected %d, got %d", defaultValue, ServicePort(key, defaultValue))
	}

	// Caso: Valore numerico
	os.Setenv(key, "9090")
	defer os.Unsetenv(key)

	if ServicePort(key, defaultValue) != 9090 {
		t.Errorf("expected 9090, got %d", ServicePort(key, defaultValue))
	}

	os.Setenv(key, "tcp://10.0.0.1:7070")
	if ServicePort(key, defaultValue) != 7070 {
		t.Errorf("expected 7070, got %d", ServicePort(key, defaultValue))
	}

	os.Setenv(key, "tcp://10.0.0.1:")
	if ServicePort(key, defaultValue) != defaultValue {
		t.Errorf("expected default %d, got %d", defaultValue, ServicePort(key, defaultValue))
	}

	os.Setenv(key, "invalid")
	if ServicePort(key, defaultValue) != defaultValue {
		t.Errorf("expected default %d, got %d", defaultValue, ServicePort(key, defaultValue))
	}
}

func TestTestMode(t *testing.T) {
	os.Unsetenv("TEST_MODE")
	if TestMode() != false {
		t.Errorf("expected false, got true")
	}

	SetTestMode(true)
	if TestMode() != true {
		t.Errorf("expected true, got false")
	}

	SetTestMode(false)
	if TestMode() != false {
		t.Errorf("expected false, got true")
	}
}
