package quasar

import (
	"fmt"
	"reflect"
	"time"
)

// time since unix epoch in ms
func newTimestampMS() uint64 {
	return uint64(time.Now().UnixNano()) / uint64(time.Millisecond)
}

func mustBeTrue(value bool, format string, a ...interface{}) {
	if value != true {
		panic(fmt.Sprintf(format, a...))
	}
}

func mustNotError(err error) {
	if err != nil {
		panic(fmt.Sprintf("Enexpected error: %s", err.Error()))
	}
}

func isNil(a interface{}) bool {
	defer func() { recover() }()
	return a == nil || reflect.ValueOf(a).IsNil()
}

func mustNotBeNil(a interface{}) {
	if isNil(a) {
		panic("Expected non nil value!")
	}
}
