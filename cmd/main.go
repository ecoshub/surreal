package main

import (
	"fmt"
	sti "sti/internal"
)

func main() {
	c := sti.ParseConfigFlags()

	ph, err := sti.New(c)
	if err != nil {
		fmt.Println(err)
		return
	}

	ph.StartTerminal()
}
