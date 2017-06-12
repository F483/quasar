package quasar

import "fmt"

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
