package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
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

func (blockChain *BlockChain) AddBlock(Transactions []*Transaction) *Block {
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

	return newBlock
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

func (chain *BlockChain) FindUnspentTxs(pubKey []byte) []Transaction {
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
				if out.KeyLocked(pubKey) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.TxInputs {
					if in.UsesKey(pubKey) {
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

func (chain *BlockChain) FindUTXOs(pubKey []byte) []TxOutput {
	var UTXOs []TxOutput
	unspentTxs := chain.FindUnspentTxs(pubKey)

	for _, tx := range unspentTxs {
		for _, out := range tx.TxOutputs {
			if out.KeyLocked(pubKey) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (chain *BlockChain) FindUTXO() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
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
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.TxInputs {
					inTxID := hex.EncodeToString(in.Id)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.OutIndex)
				}
			}
		}

		if len(block.PreviousHash) == 0 {
			break
		}
	}
	return UTXO
}

func (chain *BlockChain) FindSpendableOutputs(pubKey []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTxs(pubKey)
	accumulated := 0

Collect:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.Id)

		for outIdx, out := range tx.TxOutputs {
			if out.KeyLocked(pubKey) && accumulated < amount {
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

func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.Id, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PreviousHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

func (blockchain *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	previousTransaction := make(map[string]Transaction)
	for _, in := range tx.TxInputs {
		prevTX, err := blockchain.FindTransaction(in.Id)
		if err != nil {
			log.Panic(err)
		}
		previousTransaction[hex.EncodeToString(prevTX.Id)] = prevTX
	}

	tx.Sign(privKey, previousTransaction)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, in := range tx.TxInputs {
		prevTX, err := bc.FindTransaction(in.Id)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.Id)] = prevTX
	}
	return tx.Verify(prevTXs)
}
