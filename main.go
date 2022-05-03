package main

import (
	"os"
)

func main() {
	defer os.Exit(0)
	cli := Command{}
	cli.run()
}
