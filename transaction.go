package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

type Transaction struct {
	Id        []byte
	TxInputs  []TxInput
	TxOutputs []TxOutput
}

type TxOutput struct {
	Amount    int
	PublicKey []byte
}

type TxOutputs struct {
	Outputs []TxOutput
}

type TxInput struct {
	Id        []byte
	OutIndex  int
	Signature []byte
	PublicKey []byte
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

// func (txIn *TxInput) IsSigned(data string) bool {
// 	return txIn.Signature == data
// }

// func (txOut *TxOutput) CanBeUnlocked(data string) bool {
// 	return txOut.PublicKey == data
// }

func (txtIn *TxInput) UsesKey(pubKey []byte) bool {
	lockhash := PublicKeyHash(txtIn.PublicKey)
	return bytes.Compare(lockhash, pubKey) == 0
}

func (txOut *TxOutput) Lock(address []byte) {
	pubKey := Base58Decode(address)
	pubKey = pubKey[1 : len(pubKey)-4]
	txOut.PublicKey = pubKey
}

func (txOut *TxOutput) KeyLocked(pubKey []byte) bool {
	return bytes.Compare(txOut.PublicKey, pubKey) == 0
}

func NewTxOut(val int, address string) *TxOutput {
	txo := TxOutput{val, nil}
	txo.Lock([]byte(address))

	return &txo
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.Id = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}
func (tx Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, in := range tx.TxInputs {
		if prevTXs[hex.EncodeToString(in.Id)].Id == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inId, in := range txCopy.TxInputs {
		prevTXs := prevTXs[hex.EncodeToString(in.Id)]
		txCopy.TxInputs[inId].Signature = nil
		txCopy.TxInputs[inId].PublicKey = prevTXs.TxOutputs[in.OutIndex].PublicKey
		txCopy.Id = txCopy.Hash()
		txCopy.TxInputs[inId].PublicKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.Id)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.TxInputs[inId].Signature = signature
	}
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, in := range tx.TxInputs {
		inputs = append(inputs, TxInput{in.Id, in.OutIndex, nil, nil})
	}

	for _, out := range tx.TxOutputs {
		outputs = append(outputs, TxOutput{out.Amount, out.PublicKey})
	}

	txCopy := Transaction{tx.Id, inputs, outputs}

	return txCopy
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, in := range tx.TxInputs {
		if prevTXs[hex.EncodeToString(in.Id)].Id == nil {
			log.Panic("ERROR: Previous transaction does not exist")
		}
	}
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()
	for inId, in := range tx.TxInputs {
		prevTXs := prevTXs[hex.EncodeToString(in.Id)]
		txCopy.TxInputs[inId].Signature = nil
		txCopy.TxInputs[inId].PublicKey = prevTXs.TxOutputs[in.OutIndex].PublicKey
		txCopy.Id = txCopy.Hash()
		txCopy.TxInputs[inId].PublicKey = nil
		r := big.Int{}
		s := big.Int{}
		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.Signature)
		r.SetBytes(in.Signature[:(keyLen / 2)])
		s.SetBytes(in.Signature[(keyLen / 2):])
		x.SetBytes(in.PublicKey[:(keyLen / 2)])
		y.SetBytes(in.PublicKey[(keyLen / 2):])
		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.Id, &r, &s) == false {
			return false
		}

	}

	return true
}

func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.Id))
	for i, input := range tx.TxInputs {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:     %x", input.Id))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.OutIndex))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PublicKey))
	}

	for i, output := range tx.TxOutputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Amount))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PublicKey))
	}

	return strings.Join(lines, "\n")
}

// coinbase tx is tx that has no sender, usually the first tx of genesis block or reward tx
func CreateCoinbaseTx(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randData)
	}

	txIn := TxInput{[]byte{}, -1, nil, []byte(data)}
	txOut := NewTxOut(100, to)

	tx := Transaction{nil, []TxInput{txIn}, []TxOutput{*txOut}}
	tx.Id = tx.Hash()

	return &tx
}

func CreateTx(from, to string, amount int, UTXO *UTXOSet) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	wallets, errW := CreateWallets()
	if errW != nil {
		log.Panic(errW)
	}

	w := wallets.GetWallet(from)
	pubKey := PublicKeyHash(w.PublicKey)

	acc, validOutputs := UTXO.FindSpendableOutputs(pubKey, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	for txid, outs := range validOutputs {
		txId, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			inputs = append(inputs, TxInput{txId, out, nil, pubKey})
		}
	}

	outputs = append(outputs, *NewTxOut(amount, to))

	if acc > amount {
		outputs = append(outputs, *NewTxOut(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.Id = tx.Hash()

	UTXO.Blockchain.SignTransaction(&tx, w.PrivateKey)
	return &tx
}

func (outs TxOutputs) Serialize() []byte {
	var buffer bytes.Buffer

	encode := gob.NewEncoder(&buffer)
	err := encode.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buffer.Bytes()
}

func DeserializeOutputs(data []byte) TxOutputs {
	var outputs TxOutputs

	decode := gob.NewDecoder(bytes.NewReader(data))
	err := decode.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}
