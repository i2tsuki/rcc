package rcc

type app struct {
	Name    string
	Version string
}

// App include application name and version
var App = app{
	Name:    "rcc",
	Version: "0.1.0",
}
