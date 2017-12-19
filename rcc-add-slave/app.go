package main

import (
	"github.com/kizkoh/rcc"
)

type app struct {
	Name    string
	Version string
}

// App include application name and version
var App = app{
	Name:    "rcc-add-slave",
	Version: rcc.App.Version,
}
