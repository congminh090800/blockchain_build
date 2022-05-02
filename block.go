package main

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
