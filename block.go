package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp    int64
	Hash         []byte
	Transactions []*Transaction
	PreviousHash []byte
	Nonce        int
	Height       int
}

func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte

	for _, tx := range block.Transactions {
		txHashes = append(txHashes, tx.Serialize())
	}
	tree := CreateMerkleTree(txHashes)

	return tree.RootNode.Data
}

func CreateBlock(txs []*Transaction, prevHash []byte, height int) *Block {
	block := &Block{time.Now().Unix(), []byte{}, txs, prevHash, 0, height}
	pow := StartProofOfWork(block)
	nonce, hash := pow.Start()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{}, 0)
}

// Serialize and deserialize data to save and load to database
func (block *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(block)

	if err != nil {
		log.Panic(err)
	}

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
