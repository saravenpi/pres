package main

import (
	"github.com/saravenpi/catalyst/cmd/catalyst"
)

var version = "0.1.2"

func main() {
	catalyst.Execute(version)
}