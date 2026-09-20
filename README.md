# Wire: Automated Initialization in Go

[![Build Status](https://github.com/wireinject/wire/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/wireinject/wire/actions)
[![godoc](https://godoc.org/github.com/google/wire?status.svg)][godoc]

> [!WARNING]
> The original `google/wire` project is no longer maintained. It was archived again
> on August 25, 2025 after a brief unarchiving period, and it is now completely
> unmaintained.
>
> This is a version maintained by [wireinject/wire](https://github.com/wireinject/wire).

Wire is a code generation tool that automates connecting components using
[dependency injection][]. Dependencies between components are represented in
Wire as function parameters, encouraging explicit initialization instead of
global variables. Because Wire operates without runtime state or reflection,
code written to be used with Wire is useful even for hand-written
initialization.

For an overview, see the [introductory blog post][].

[dependency injection]: https://en.wikipedia.org/wiki/Dependency_injection
[introductory blog post]: https://blog.golang.org/wire
[godoc]: https://godoc.org/github.com/google/wire
[travis]: https://travis-ci.com/google/wire

## Installing

Install Wire by running:

```shell
go install github.com/wireinject/wire/cmd/wire@latest
```

and ensuring that `$GOPATH/bin` is added to your `$PATH`.

## Documentation

- [Tutorial][]
- [User Guide][]
- [Command Line Reference][]
- [Best Practices][]
- [FAQ][]

[Tutorial]: ./_tutorial/README.md
[Best Practices]: ./docs/best-practices.md
[Command Line Reference]: ./docs/cli.md
[FAQ]: ./docs/faq.md
[User Guide]: ./docs/guide.md

## Project status

As of version v0.3.0, Wire is *beta* and is considered feature complete. It
works well for the tasks it was designed to perform, and we prefer to keep it
as simple as possible.

Maintenance is now handled by [wireinject/wire](https://github.com/wireinject/wire),
with a deliberately narrow scope:

- We fix bugs, including compatibility problems with new Go toolchains and their
  dependencies.
- We add small, scoped features when they are clearly useful and low-risk.
- We do not overhaul the project. The core design, public API, and command-line
  interface stay compatible with `google/wire`.
- We do not modify the golden tests that pin the generated output. The code `wire`
  produces for existing projects keeps working exactly as before.

Stability comes first: every change is weighed against the risk it poses to
existing users.

## Community

For questions, please use [GitHub Discussions](https://github.com/wireinject/wire/discussions).

This project is covered by the Go [Code of Conduct][].

[Code of Conduct]: ./CODE_OF_CONDUCT.md
[go-cloud mailing list]: https://groups.google.com/forum/#!forum/go-cloud
