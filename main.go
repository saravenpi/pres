package main

import (
	"github.com/saravenpi/catalyst/cmd/catalyst"
)

var version = "0.1.3"

func main() {
	catalyst.Execute(version)
}