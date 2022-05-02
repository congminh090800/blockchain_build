package main

import (
	"os"
)

func main() {
	defer os.Exit(0)
	chain := InitMyChain()
	defer chain.Database.Close()

	cli := Command{chain}
	cli.run()
}
