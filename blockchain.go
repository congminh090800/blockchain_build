package main

type BlockChain struct {
	Blocks []*Block
}

func (blockChain *BlockChain) AddBlock(data string) {
	latestBlock := blockChain.Blocks[len(blockChain.Blocks)-1]
	newBlock := CreateBlock(data, latestBlock.Hash)
	blockChain.Blocks = append(blockChain.Blocks, newBlock)
}

func CreateGenesisBlock() *Block {
	return CreateBlock("Genesis Block", []byte{})
}

func InitMyChain() *BlockChain {
	return &BlockChain{[]*Block{CreateGenesisBlock()}}
}
