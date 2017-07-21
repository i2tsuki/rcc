package main

import (
	"github.com/kizkoh/rca"
)

type app struct {
	Name    string
	Version string
}

// App include application name and version
var App = app{
	Name:    "rca-cluster-tree",
	Version: rca.App.Version,
}
