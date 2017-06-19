package quasar

import (
	"fmt"
	"math/rand"
)

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

func randIntnExcluding(limit int, exclude int) int {
	// FIXME validate input is sane
	for {
		n := rand.Intn(limit)
		if n != exclude {
			return n
		}
	}
}
