package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type DataSend struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
	Mine   bool   `json:"mine"`
}

func main() {
	os.Setenv("NODE_ID", "3000")
	nodeId := os.Getenv("NODE_ID")
	wallets, _ := CreateWallets(nodeId)
	address := wallets.AddWallets()

	wallets.SaveFile(nodeId)
	chain := InitMyChain(address, nodeId)

	u := UTXOSet{chain}
	u.Reindex()
	chain.Database.Close()
	r := gin.Default()
	r.Use(cors.Default())
	r.POST("/createwallet", func(c *gin.Context) {
		nodeId := os.Getenv("NODE_ID")
		wallets, _ := CreateWallets(nodeId)
		address := wallets.AddWallets()

		wallets.SaveFile(nodeId)
		c.JSON(200, gin.H{
			"data": address,
		})
	})
	r.POST("/send", func(c *gin.Context) {
		nodeId := os.Getenv("NODE_ID")
		var data DataSend
		if err := c.BindJSON(&data); err != nil {
			fmt.Println(err)
			return
		}

		if !ValidateAddress(data.From) {
			log.Panic("Address is not valid")
		}

		if !ValidateAddress(data.To) {
			log.Panic("Address is not valid")
		}

		chain := LoadBlockchain(nodeId)
		UTXOSet := UTXOSet{chain}
		defer chain.Database.Close()

		wallets, err := CreateWallets(nodeId)
		if err != nil {
			log.Panic(err)
		}
		wallet := wallets.GetWallet(data.From)
		amount, _ := strconv.ParseInt(data.Amount, 10, 64)
		tx := CreateTx(&wallet, data.To, int(amount), &UTXOSet)
		if data.Mine {
			cbTx := CreateCoinbaseTx(data.From, "")
			txs := []*Transaction{cbTx, tx}
			block := chain.MineBlock(txs)
			UTXOSet.Update(block)
		} else {
			txs := []*Transaction{tx}
			block := chain.MineBlock(txs)
			UTXOSet.Update(block)
		}
	})
	r.GET("/listaddresses", func(c *gin.Context) {
		nodeId := os.Getenv("NODE_ID")
		wallets, _ := CreateWallets(nodeId)
		addresses := wallets.GetAllAddresses()
		c.JSON(200, gin.H{
			"data": addresses,
		})
	})
	r.GET("/getbalance", func(c *gin.Context) {
		nodeId := os.Getenv("NODE_ID")
		address := c.Query("address")
		if !ValidateAddress(address) {
			c.JSON(400, gin.H{
				"message": "wallet is not valid",
			})
		}
		chain := LoadBlockchain(nodeId)
		utxoSet := UTXOSet{chain}
		defer chain.Database.Close()

		balance := 0

		pubKey := Base58Decode([]byte(address))
		pubKey = pubKey[1 : len(pubKey)-4]
		UTXOs := utxoSet.FindUTXO(pubKey)

		for _, out := range UTXOs {
			balance += out.Amount
		}

		c.JSON(200, gin.H{
			"address": address,
			"balance": balance,
		})
	})
	r.GET("/print", func(c *gin.Context) {
		nodeId := os.Getenv("NODE_ID")
		chain := LoadBlockchain(nodeId)
		defer chain.Database.Close()
		iter := chain.Iterator()
		var data []Block
		for {
			block := iter.Next()
			data = append(data, *block)
			if len(block.PreviousHash) == 0 {
				break
			}
		}
		c.JSON(200, gin.H{
			"data": data,
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
