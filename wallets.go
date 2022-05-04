package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletFile = "./wallets.data"

type Wallets struct {
	Wallets map[string]*Wallet
}

func (ws *Wallets) SaveFile(nodeId string) {
	var content bytes.Buffer
	walletFile := fmt.Sprintf(walletFile, nodeId)
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)

	if err != nil {
		log.Panic(err)
	}

	err2 := ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err2 != nil {
		log.Panic(err2)
	}

}

func (ws *Wallets) LoadFile(nodeId string) error {
	walletFile := fmt.Sprintf(walletFile, nodeId)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}
	var wallets Wallets
	content, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	gob.Register(elliptic.P256())
	decode := gob.NewDecoder(bytes.NewReader(content))
	err2 := decode.Decode(&wallets)
	if err2 != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets
	return nil
}

func CreateWallets(nodeId string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFile(nodeId)
	return &wallets, err
}

func (ws *Wallets) GetAllAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

func (ws *Wallets) AddWallets() string {
	wallet := MakeWallet()
	address := fmt.Sprintf("%s", wallet.Address())
	ws.Wallets[address] = wallet

	return address
}
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}
