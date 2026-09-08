package main

import (
	"github.com/google/wire"
)

func injectInt() int {
	wire.Build(wire.FieldsOf(new(*int), "X"))
	return 0
}
