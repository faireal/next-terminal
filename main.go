package main

import (
	"next-terminal/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
		os.Exit(-1)
	}
}
