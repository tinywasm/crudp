module github.com/tinywasm/crudp

go 1.25.2

require (
	github.com/tinywasm/binary v0.3.1
	github.com/tinywasm/context v0.0.1
	github.com/tinywasm/fmt v0.12.6
)

replace github.com/tinywasm/fmt => ../fmt

replace github.com/tinywasm/binary => ../binary

replace github.com/tinywasm/time => ../time

replace github.com/tinywasm/context => ../context
