package main

import (
	"surreal/src/core"
)

func main() {
	c := core.ParseConfigFlags()

	ph, err := core.New(c)
	if err != nil {
		panic(err)
	}

	ph.StartTerminal()
}
