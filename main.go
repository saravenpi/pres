package main

import (
	"github.com/saravenpi/pres/cmd/pres"
)

var version = "0.2.0"

func main() {
	pres.Execute(version)
}