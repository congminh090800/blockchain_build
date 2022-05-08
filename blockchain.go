package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dgraph-io/badger"
)

const dbPath = "./db/blocks_%s"
const genesisData = "First Transaction from Genesis"

type BlockChain struct {
	LatestHash []byte
	Database   *badger.DB
}

func isDbExisted(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}

	return true
}

func InitBlockChain(address, nodeId string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeId)
	if isDbExisted(path) {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}
	var lastHash []byte
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path

	db, err := openDB(path, opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CreateCoinbaseTx(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = txn.Set([]byte("latestHash"), genesis.Hash)

		lastHash = genesis.Hash

		return err

	})

	if err != nil {
		log.Panic(err)
	}

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func LoadBlockchain(address string) *BlockChain {
	path := fmt.Sprintf(dbPath, address)
	if !isDbExisted(path) {
		fmt.Println("Please init your blockchain first!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path

	db, err := openDB(path, opts)
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

func (chain *BlockChain) AddBlock(block *Block) {
	err := chain.Database.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}

		blockData := block.Serialize()
		err := txn.Set(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		item, err := txn.Get([]byte("latestHash"))
		if err != nil {
			log.Panic(err)
		}
		lastHash, _ := item.Value()

		item, err = txn.Get(lastHash)
		if err != nil {
			log.Panic(err)
		}
		lastBlockData, _ := item.Value()

		lastBlock := Deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = txn.Set([]byte("latestHash"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			chain.LatestHash = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (chain *BlockChain) GetBestHeight() int {
	var lastBlock Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("latestHash"))
		if err != nil {
			log.Panic(err)
		}
		lastHash, _ := item.Value()

		item, err = txn.Get(lastHash)
		if err != nil {
			log.Panic(err)
		}
		lastBlockData, _ := item.Value()

		lastBlock = *Deserialize(lastBlockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

func (chain *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("Block is not found")
		} else {
			blockData, _ := item.Value()

			block = *Deserialize(blockData)
		}
		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

func (chain *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte

	iter := chain.Iterator()

	for {
		block := iter.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PreviousHash) == 0 {
			break
		}
	}

	return blocks
}

func (chain *BlockChain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		if chain.VerifyTransaction(tx) != true {
			log.Panic("Invalid Transaction")
		}
	}

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("latestHash"))
		if err != nil {
			log.Panic(err)
		}
		lastHash, err = item.Value()

		item, err = txn.Get(lastHash)
		if err != nil {
			log.Panic(err)
		}
		lastBlockData, _ := item.Value()

		lastBlock := Deserialize(lastBlockData)

		lastHeight = lastBlock.Height

		return err
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := CreateBlock(transactions, lastHash, lastHeight+1)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = txn.Set([]byte("latestHash"), newBlock.Hash)

		chain.LatestHash = newBlock.Hash

		return err
	})
	if err != nil {
		log.Panic(err)
	}

	return newBlock
}

func CreateGenesisBlock(CoinbaseTx *Transaction) *Block {
	return CreateBlock([]*Transaction{CoinbaseTx}, []byte{}, 0)
}

func InitMyChain(address, nodeId string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeId)
	if isDbExisted(path) {
		fmt.Println("Your blockchain is already running")
		runtime.Goexit()
	}
	var latestHash []byte

	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path

	db, err := openDB(path, opts)

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

func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	retryOpts.Truncate = true
	db, err := badger.Open(retryOpts)
	return db, err
}

func openDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				log.Println("database unlocked, value log truncated")
				return db, nil
			}
			log.Println("could not unlock database:", err)
		}
		return nil, err
	} else {
		return db, nil
	}
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
