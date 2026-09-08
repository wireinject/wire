package main

import (
	"github.com/google/wire"
)

type A struct{}

func newA() *A {
	return &A{}
}

type Builder struct{}

func (Builder) injectA() *A {
	wire.Build(newA)
	return nil
}
