package main

import (
	"example.com/other"

	"github.com/google/wire"
)

func provideString() string {
	return "x"
}

func injectS() other.S {
	wire.Build(provideString, wire.Struct(new(other.S), "priv"))
	return other.S{}
}
