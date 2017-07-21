package rca

import (
	"github.com/kizkoh/rca"
)

type app struct {
	Name    string
	Version string
}

// App include application name and version
var App = app{
	Name:    "rca",
	Version: rca.App.Version,
}
