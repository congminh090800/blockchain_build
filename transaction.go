package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Transaction struct {
	Id        []byte
	TxInputs  []TxInput
	TxOutputs []TxOutput
}

type TxOutput struct {
	Amount    int
	PublicKey string
}

type TxInput struct {
	Id        []byte
	OutIndex  int
	Signature string
}

func (tx *Transaction) GenerateID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.Id = hash[:]
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.TxInputs) == 1 && len(tx.TxInputs[0].Id) == 0 && tx.TxInputs[0].OutIndex == -1
}

func (txIn *TxInput) IsSigned(data string) bool {
	return txIn.Signature == data
}

func (txOut *TxOutput) CanBeUnlocked(data string) bool {
	return txOut.PublicKey == data
}

// coinbase tx is tx that has no sender, usually the first tx of genesis block or reward tx
func CreateCoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Send reward to %s", to)
	}

	txIn := TxInput{[]byte{}, -1, data}
	txOut := TxOutput{100, to}

	tx := Transaction{nil, []TxInput{txIn}, []TxOutput{txOut}}
	tx.GenerateID()

	return &tx
}

func CreateTx(from, to string, amount int, blockChain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := blockChain.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	for txid, outs := range validOutputs {
		txId, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			inputs = append(inputs, TxInput{txId, out, from})
		}
	}

	outputs = append(outputs, TxOutput{amount, to})

	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.GenerateID()

	return &tx
}
