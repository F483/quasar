package quasar

import (
	"fmt"
	"testing"
)

func TestMustNotBeNilPtr(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("mustNotBeNil did not panic as expected")
		}
	}()

	var i *int
	mustNotBeNil(i)
}

func TestMustNotBeNilDirect(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("mustNotBeNil did not panic as expected")
		}
	}()
	mustNotBeNil(nil)
}

func TestMustBeTrue(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("mustBeTrue did not panic as expected")
		}
	}()

	mustBeTrue(false, "must panic")
}

func TestMustNotError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("mustNotError did not panic as expected")
		}
	}()
	mustNotError(fmt.Errorf("must panic"))
}
