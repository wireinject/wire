//go:build wireinject

package main

import "github.com/google/wire"

func injectInt() int {
	wire.Build(nil)
	return 0
}
