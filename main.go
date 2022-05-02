package main

import (
	"fmt"
	"strconv"
)

func main() {
	chain := InitMyChain()

	chain.AddBlock("First")
	chain.AddBlock("Second")
	chain.AddBlock("Third")

	for _, block := range chain.Blocks {
		fmt.Printf("-----------------------------------\n")
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Data: %s\n", block.Data)
		pow := StartProofOfWork(block)
		fmt.Printf("Proof Of Work: %s\n", strconv.FormatBool(pow.Validate()))
	}
}
