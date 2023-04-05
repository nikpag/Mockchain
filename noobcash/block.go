package main

type Block struct {
	Index        uint
	Timestamp    int64
	PreviousHash [32]byte
	Transactions [CAPACITY]SignedTransaction
	Nonce        [32]byte
}

type HashedBlock struct {
	Index        uint
	Timestamp    int64
	PreviousHash [32]byte
	Transactions [CAPACITY]SignedTransaction
	Nonce        [32]byte
	Hash         [32]byte
}
