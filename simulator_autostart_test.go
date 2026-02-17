package simulator_autostart_test

import (
	"testing"

	"simulator_autostart/lib/autostart"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v, got %v", a, b)
	}
}

func assertGreater(t *testing.T, a int, b int) {
	if a <= b {
		t.Errorf("Expected %v > %v", a, b)
	}
}

func Test_isProcessRunning(t *testing.T) {
	assertGreater(t, autostart.GetProcessIDForExecutable("svchost.exe"), 0)
	assertEqual(t, autostart.GetProcessIDForExecutable("Some Fake exe not running at all.exe"), -1)
}
