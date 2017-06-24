package quasar

import "testing"

func TestSimulate(t *testing.T) {
	results := Simulate(&testConfig, 16, 1, 1024, false)
	if results == nil {
		// FIXME check values
		t.Errorf("Simulation failed!")
	}
}
