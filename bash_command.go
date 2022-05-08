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
	nodeId := os.Getenv("NODE_ID")
	fmt.Println("NODE_ID: ", nodeId)
	fmt.Println("Commands:")
	fmt.Println("Create blockchain: initChain -address [address]")
	fmt.Println("View all blocks: print")
	fmt.Println("Send coins from one to another address, then -mine flag is set, mine off of this node: send -from [fromAddress] -to [toAddress] -amount [amount] -mine")
	fmt.Println("Get balance of an address: getBalance -address [address]")
	fmt.Println("List all addresses: listAddresses")
	fmt.Println("Create wallets: createwallet")
	fmt.Println("Rebuild UTXO set: reindexutxo")
	fmt.Println("Start a node with ID specified in NODE_ID env, -miner enables mining: startnode -miner ADDRESS")
}

func (cli *Command) validateArgs() {
	if len(os.Args) < 2 {
		cli.printMenu()
		runtime.Goexit()
	}

}

func (cli *Command) print(nodeId string) {
	chain := LoadBlockchain(nodeId)
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

func (cli *Command) createBlockChain(address, nodeId string) {
	chain := InitMyChain(address, nodeId)
	defer chain.Database.Close()

	UTXOSet := UTXOSet{chain}
	UTXOSet.Reindex()

	fmt.Println("Finished!")
}

func (cli *Command) getBalance(address, nodeId string) {

	if !ValidateAddress(address) {
		log.Panic("Address is not valid")
	}

	chain := LoadBlockchain(nodeId)
	UTXOSet := UTXOSet{chain}
	defer chain.Database.Close()

	balance := 0

	pubKey := Base58Decode([]byte(address))
	pubKey = pubKey[1 : len(pubKey)-4]
	UTXOs := UTXOSet.FindUTXO(pubKey)

	for _, out := range UTXOs {
		balance += out.Amount
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *Command) send(from, to string, amount int, nodeId string, mineNow bool) {

	if !ValidateAddress(from) {
		log.Panic("Address is not valid")
	}

	if !ValidateAddress(to) {
		log.Panic("Address is not valid")
	}

	chain := LoadBlockchain(nodeId)
	UTXOSet := UTXOSet{chain}
	defer chain.Database.Close()

	wallets, err := CreateWallets(nodeId)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := CreateTx(&wallet, to, amount, &UTXOSet)
	if mineNow {
		cbTx := CreateCoinbaseTx(from, "")
		txs := []*Transaction{cbTx, tx}
		block := chain.MineBlock(txs)
		UTXOSet.Update(block)
	} else {
		SendTx(KnownNodes[0], tx)
		fmt.Println("send tx")
	}

	fmt.Printf("%s sent %d to %s\n", from, amount, to)
}

func (cli *Command) listAddresses(nodeId string) {
	wallets, _ := CreateWallets(nodeId)
	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *Command) createWallet(nodeId string) {
	wallets, _ := CreateWallets(nodeId)
	address := wallets.AddWallets()

	wallets.SaveFile(nodeId)
	fmt.Printf("Your address is: %s\n", address)
}

func (cli *Command) run() {
	cli.validateArgs()

	nodeId := os.Getenv("NODE_ID")
	if nodeId == "" {
		fmt.Printf("NODE_ID env is not set!")
		runtime.Goexit()
	}

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
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")

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
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
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
		cli.getBalance(*getBalanceAddress, nodeId)
	}

	if initChainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			initChainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress, nodeId)
	}

	if printCmd.Parsed() {
		cli.print(nodeId)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeId)
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeId)
	}

	if sendCmd.Parsed() {
		if *fromAddress == "" || *toAddress == "" || *amount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*fromAddress, *toAddress, *amount, nodeId, *sendMine)
	}
	if reindexUTXOCmd.Parsed() {
		cli.reindexUTXO()
	}
	if startNodeCmd.Parsed() {
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			startNodeCmd.Usage()
			runtime.Goexit()
		}
		cli.StartNode(nodeID, *startNodeMiner)
	}
}

func (cli *Command) reindexUTXO() {
	chain := LoadBlockchain("")
	defer chain.Database.Close()
	UTXOSet := UTXOSet{chain}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

func (cli *Command) StartNode(nodeID, minerAddress string) {
	fmt.Printf("Starting Node %s\n", nodeID)

	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	StartServer(nodeID, minerAddress)
}
