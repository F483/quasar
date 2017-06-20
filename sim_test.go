package quasar

import (
	"testing"
)

func TestSimulate(t *testing.T) {

	results := Simulate(&testConfig, 12, 12, 12)
	if results == nil {
		// FIXME check values
		t.Errorf("Simulation failed!")
	}
}
