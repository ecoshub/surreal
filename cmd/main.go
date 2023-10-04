package main

import (
	sti "sti/internal"
)

func main() {
	c := sti.ParseConfigFlags()

	ph, err := sti.New(c)
	if err != nil {
		panic(err)
	}

	ph.StartTerminal()
}
