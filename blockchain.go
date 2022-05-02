package main

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger"
)

const dbPath = "./db/blocks"

type BlockChain struct {
	LatestHash []byte
	Database   *badger.DB
}

func (blockChain *BlockChain) AddBlock(data string) {
	var lastHash []byte

	err := blockChain.Database.View(func(txn *badger.Txn) error {
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

	newBlock := CreateBlock(data, lastHash)

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

func CreateGenesisBlock() *Block {
	return CreateBlock("Genesis Block", []byte{})
}

func InitMyChain() *BlockChain {
	var latestHash []byte
	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	db, err := badger.Open(opts)

	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("latestHash")); err == badger.ErrKeyNotFound {
			genesis := CreateGenesisBlock()
			fmt.Println("Created genesis block")
			err = txn.Set(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			err = txn.Set([]byte("latestHash"), genesis.Hash)
			latestHash = genesis.Hash
			return err
		} else {
			item, err := txn.Get([]byte("latestHash"))
			if err != nil {
				log.Panic(err)
			}
			latestHash, err = item.Value()
			return err
		}
	})
	if err != nil {
		log.Panic(err)
	}
	return &BlockChain{latestHash, db}
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
