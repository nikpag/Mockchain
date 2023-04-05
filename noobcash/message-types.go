package main

import "crypto/rsa"

type MessageType int

const (
	NullMessageType MessageType = iota
	WelcomeMessageType
	MyInfoMessageType
	NodeDataMessageType
	NewConnectionMessageType
	TransactionMessageType
	BlockMessageType
	ResolveRequestMessageType
	ResolveResponseMessageType
)

type NullMessage struct{}

type WelcomeMessage struct {
	ID string
}

type MyInfoMessage struct {
	ID        string
	PublicKey rsa.PublicKey
	Address   string
}

type NodeDataMessage struct {
	Neighbors map[string]NodeData
}

type NewConnectionMessage struct {
	ID string
}

type TransactionMessage struct {
	Transaction SignedTransaction
}

type BlockMessage struct {
	Block HashedBlock
}

type ResolveRequestMessage struct {
	ChainSize uint
	Hashes    [][32]byte
}

type ResolveResponseMessage struct {
	Blocks []HashedBlock
}

type Message struct {
	MessageType MessageType
	MessageData []byte
}
