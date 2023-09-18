package main

import (
	"fmt"
	"surreal/surreal"
)

func main() {
	c := surreal.ParseConfigFlags()

	ph, err := surreal.New(c)
	if err != nil {
		fmt.Println(err)
		return
	}

	ph.StartTerminal()
}
