package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"

	"golang.org/x/crypto/ripemd160"

	"github.com/mr-tron/base58"
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func CreatePair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)

	if err != nil {
		log.Panic(err)
	}

	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pub

}

const (
	checkSumLength = 4
	version        = byte(0x00)
)

func MakeWallet() *Wallet {
	private, public := CreatePair()
	wallet := Wallet{private, public}

	return &wallet
}

func CheckSum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	return second[:checkSumLength]
}

func PublicKeyHash(pubKey []byte) []byte {
	pubKeyHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()

	_, err := hasher.Write(pubKeyHash[:])

	if err != nil {
		log.Panic(err)
	}
	publicRipemd160 := hasher.Sum(nil)

	return publicRipemd160
}
func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)
	return []byte(encode)
}

func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input[:]))
	if err != nil {
		log.Panic(err)
	}

	return decode
}

func (w Wallet) Address() []byte {
	pHash := PublicKeyHash(w.PublicKey)
	versionH := append([]byte{version}, pHash...)
	checkSum := CheckSum(versionH)

	fullH := append(versionH, checkSum...)
	address := Base58Encode(fullH)
	fmt.Printf("Pubic Key: %x\n", w.PublicKey)
	fmt.Printf("Public Key Hash: %x\n", pHash)
	fmt.Printf("Address: %s\n", address)
	return address
}
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-checkSumLength:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checkSumLength]
	targetChecksum := CheckSum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}
