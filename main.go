package main

import (
	"next-terminal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
