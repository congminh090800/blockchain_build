package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const dbPath = "./db/blocks"
const dbManifest = "./db/blocks/MANIFEST"

type BlockChain struct {
	LatestHash []byte
	Database   *badger.DB
}

func isDbExisted() bool {
	if _, err := os.Stat(dbManifest); os.IsNotExist(err) {
		return false
	}

	return true
}

func LoadBlockchain(address string) *BlockChain {
	if !isDbExisted() {
		fmt.Println("Please init your blockchain first!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("latestHash"))
		if err != nil {
			log.Panic(err)
		}
		lastHash, err = item.Value()

		return err
	})
	if err != nil {
		log.Panic(err)
	}

	chain := BlockChain{lastHash, db}

	return &chain
}

func (blockChain *BlockChain) AddBlock(Transactions []*Transaction) {
	var lastestHash []byte

	err := blockChain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("latestHash"))
		if err != nil {
			log.Panic(err)
		}
		lastestHash, err = item.Value()

		return err
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := CreateBlock(Transactions, lastestHash)

	err = blockChain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = txn.Set([]byte("latestHash"), newBlock.Hash)

		blockChain.LatestHash = newBlock.Hash

		return err
	})
	if err != nil {
		log.Panic(err)
	}
}

func CreateGenesisBlock(CoinbaseTx *Transaction) *Block {
	return CreateBlock([]*Transaction{CoinbaseTx}, []byte{})
}

func InitMyChain(address string) *BlockChain {
	if isDbExisted() {
		fmt.Println("Your blockchain is already running")
		runtime.Goexit()
	}
	var latestHash []byte

	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)

	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		coinbaseTx := CreateCoinbaseTx(address, "Genesis data")
		genesis := CreateGenesisBlock(coinbaseTx)
		err = txn.Set(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = txn.Set([]byte("latestHash"), genesis.Hash)

		latestHash = genesis.Hash

		return err
	})
	if err != nil {
		log.Panic(err)
	}
	return &BlockChain{latestHash, db}
}

func (chain *BlockChain) FindUnspentTxs(address string) []Transaction {
	var unspentTxs []Transaction

	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.Id)

		Outputs:
			for outIdx, out := range tx.TxOutputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.TxInputs {
					if in.IsSigned(address) {
						inTxID := hex.EncodeToString(in.Id)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.OutIndex)
					}
				}
			}
		}

		if len(block.PreviousHash) == 0 {
			break
		}
	}
	return unspentTxs
}

func (chain *BlockChain) FindUTXOs(address string) []TxOutput {
	var UTXOs []TxOutput
	unspentTxs := chain.FindUnspentTxs(address)

	for _, tx := range unspentTxs {
		for _, out := range tx.TxOutputs {
			if out.CanBeUnlocked(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (chain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTxs(address)
	accumulated := 0

Collect:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.Id)

		for outIdx, out := range tx.TxOutputs {
			if out.CanBeUnlocked(address) && accumulated < amount {
				accumulated += out.Amount
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if accumulated >= amount {
					break Collect
				}
			}
		}
	}

	return accumulated, unspentOuts
}

// travel backward the blockchain
// source: https://www.youtube.com/watch?v=szOZ3p-5YIc&list=PLpP5MQvVi4PGmNYGEsShrlvuE2B33xV1L&index=3
type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LatestHash, chain.Database}

	return iter
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		if err != nil {
			log.Panic(err)
		}
		encodedBlock, err := item.Value()
		block = Deserialize(encodedBlock)

		return err
	})
	if err != nil {
		log.Panic(err)
	}

	iter.CurrentHash = block.PreviousHash

	return block
}
