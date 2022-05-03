package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
)

type Command struct{}

func (cli *Command) printMenu() {
	fmt.Println("Commands:")
	fmt.Println("Create blockchain: initChain -address [address]")
	fmt.Println("View all blocks: print")
	fmt.Println("Send coins from one to another address: send -from [fromAddress] -to [toAddress] -amount [amount]")
	fmt.Println("Get balance of an address: getBalance -address [address]")
	fmt.Println("List all addresses: listAddresses")
	fmt.Println("Create wallets: createwallet")
}

func (cli *Command) validateArgs() {
	if len(os.Args) < 2 {
		cli.printMenu()
		runtime.Goexit()
	}

}

func (cli *Command) print() {
	chain := LoadBlockchain("")
	defer chain.Database.Close()
	iter := chain.Iterator()
	for {
		block := iter.Next()
		fmt.Printf("-----------------------------------\n")
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := StartProofOfWork(block)
		fmt.Printf("Confirmed: %s\n", strconv.FormatBool(pow.Validate()))

		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}

		if len(block.PreviousHash) == 0 {
			break
		}
	}
}

func (cli *Command) createBlockChain(address string) {
	chain := InitMyChain(address)
	chain.Database.Close()
	fmt.Println("Finished!")
}

func (cli *Command) getBalance(address string) {

	if !ValidateAddress(address) {
		log.Panic("Address is not valid")
	}

	chain := LoadBlockchain(address)
	defer chain.Database.Close()

	balance := 0

	pubKey := Base58Decode([]byte(address))
	pubKey = pubKey[1 : len(pubKey)-4]
	UTXOs := chain.FindUTXOs(pubKey)

	for _, out := range UTXOs {
		balance += out.Amount
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *Command) send(from, to string, amount int) {

	if !ValidateAddress(from) {
		log.Panic("Address is not valid")
	}

	if !ValidateAddress(to) {
		log.Panic("Address is not valid")
	}

	chain := LoadBlockchain(from)
	defer chain.Database.Close()

	tx := CreateTx(from, to, amount, chain)
	chain.AddBlock([]*Transaction{tx})
	fmt.Printf("%s sent %d to %s\n", from, amount, to)
}

func (cli *Command) listAddresses() {
	wallets, _ := CreateWallets()
	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *Command) createWallets() {
	wallets, _ := CreateWallets()
	address := wallets.AddWallets()

	wallets.SaveFile()
	fmt.Printf("Your address is: %s\n", address)
}

func (cli *Command) run() {
	cli.validateArgs()
	getBalanceCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)
	initChainCmd := flag.NewFlagSet("initChain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printCmd := flag.NewFlagSet("print", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listAddresses", flag.ExitOnError)
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := initChainCmd.String("address", "", "The address to send genesis block reward to")
	fromAddress := sendCmd.String("from", "", "Source wallet address")
	toAddress := sendCmd.String("to", "", "Destination wallet address")
	amount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "getBalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "initChain":
		err := initChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "print":
		err := printCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listAddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printMenu()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if initChainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			initChainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress)
	}

	if printCmd.Parsed() {
		cli.print()
	}

	if createWalletCmd.Parsed() {
		cli.createWallets()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if sendCmd.Parsed() {
		if *fromAddress == "" || *toAddress == "" || *amount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*fromAddress, *toAddress, *amount)
	}
}
