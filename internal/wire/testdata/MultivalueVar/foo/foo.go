package main

import (
	"github.com/google/wire"
)

func twoSets() (wire.ProviderSet, wire.ProviderSet) {
	return wire.NewSet(), wire.NewSet()
}

var A, B = twoSets()

func provideInt() int {
	return 1
}

func injectInt() int {
	wire.Build(B, provideInt)
	return 0
}
