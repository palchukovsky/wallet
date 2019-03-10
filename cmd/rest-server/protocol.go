package main

import (
	"encoding/json"
	"log"

	"github.com/palchukovsky/wallet"
)

// Protocol encapsulates responses serialization.
type Protocol interface {
	// GetContentType returns content type for document specification.
	GetContentType() string
	// SerializeTrans serializes transaction list.
	SerializeTransList([]wallet.Trans) []byte
	// SerializeAccounts serializes account list.
	SerializeAccounts([]wallet.Account) []byte
}

type protocol struct{}

// CreateProtocol creates protocol instance for JSON.
func CreateProtocol() Protocol { return &protocol{} }

func (protocol) GetContentType() string {
	return "application/json; charset=utf-8"
}

func (p protocol) SerializeTransList(trans []wallet.Trans) []byte {
	result, err := json.Marshal(trans)
	if err != nil {
		log.Panicf(`Failed to marshal transaction list: "%s".`, err)
	}
	return result
}

func (p protocol) SerializeAccounts(accounts []wallet.Account) []byte {
	result, err := json.Marshal(accounts)
	if err != nil {
		log.Panicf(`Failed to marshal account list: "%s".`, err)
	}
	return result
}
