package main

import "testing"
import "github.com/stretchr/testify/assert"

func Test_isProcessRunning(t *testing.T) {
	// Start Automobilista 2 before
	assert.Greater(t, getProcessIdForExceutable("AMS2AVX.exe"), 0)
	assert.Equal(t, getProcessIdForExceutable("Some Fake exe not running at all.exe"), -1)
}
