package main

import (
	"testing"
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
	assertGreater(t, getProcessIdForExecutable("svchost.exe"), 0)
	assertEqual(t, getProcessIdForExecutable("Some Fake exe not running at all.exe"), -1)

}
