package quasar

import "testing"

func TestSimulate(t *testing.T) {
	results := Simulate(&testConfig, 4, 1, 1)
	if results == nil {
		// FIXME check values
		t.Errorf("Simulation failed!")
	}
}
