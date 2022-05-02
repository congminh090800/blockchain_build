package main

import (
	"bytes"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash         []byte
	Data         []byte
	PreviousHash []byte
	Nonce        int
}

func CreateBlock(data string, PreviousHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), PreviousHash, 0}
	pow := StartProofOfWork(block)
	nonce, hash := pow.Start()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
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
