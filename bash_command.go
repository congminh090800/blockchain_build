package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
)

type Command struct {
	blockchain *BlockChain
}

func (cli *Command) printMenu() {
	fmt.Println("Commands:")
	fmt.Println("Append block: add -block [block_data]")
	fmt.Println("View all blocks: print")
}

func (cli *Command) validateArgs() {
	if len(os.Args) < 2 {
		cli.printMenu()
		runtime.Goexit()
	}

}

func (cli *Command) addBlock(data string) {
	cli.blockchain.AddBlock(data)
	fmt.Println("Block added")
}

func (cli *Command) print() {
	iter := cli.blockchain.Iterator()
	for {
		block := iter.Next()
		fmt.Printf("-----------------------------------\n")
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Data: %s\n", block.Data)
		pow := StartProofOfWork(block)
		fmt.Printf("Confirmed: %s\n", strconv.FormatBool(pow.Validate()))
		if len(block.PreviousHash) == 0 {
			break
		}
	}
}

func (cli *Command) run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")

	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	default:
		cli.printMenu()
		runtime.Goexit()
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.print()
	}
}
