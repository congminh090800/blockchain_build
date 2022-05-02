package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

const Difficulty = 18

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func (pow *ProofOfWork) InitData(nonce int) []byte {
	return bytes.Join([][]byte{pow.Block.PreviousHash, pow.Block.Data, ToHex(int64(nonce)), ToHex(int64(Difficulty))}, []byte{})
}

// this proof of work is from algorithm is from: https://www.youtube.com/watch?v=aE4eDTUAE70&list=PLpP5MQvVi4PGmNYGEsShrlvuE2B33xV1L&index=2
func (pow *ProofOfWork) Start() (int, []byte) {
	var hashDecimal big.Int
	var hash [32]byte
	nonce := 0
	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		// convert hash to decimal
		hashDecimal.SetBytes(hash[:])

		// this means hash had more zeros than the difficulty required
		if hashDecimal.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Println()
	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var hashDecimal big.Int
	data := pow.InitData(pow.Block.Nonce)
	hash := sha256.Sum256(data)
	hashDecimal.SetBytes(hash[:])
	return hashDecimal.Cmp(pow.Target) == -1
}

func StartProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty))
	pow := &ProofOfWork{block, target}
	return pow
}

func ToHex(num int64) []byte {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buffer.Bytes()
}
