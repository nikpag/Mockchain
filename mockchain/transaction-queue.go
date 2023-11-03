package main

import (
	"errors"
	"sync"
)

/*
This is the stack in which we control the flow of the UTXOs. Everyone updates their own and the
"global" UTXO Stack to know how many funds each user has. This update happens from everyone, in their
own copy of the structure. This only works because we make sure that the transactions run as long as
everyone is connected in the network and no one ever leaves.
*/
type TransactionQueue struct {
	queueLock    sync.Mutex
	Transactions []SignedTransaction
}

func NewTransactionQueue() *TransactionQueue {
	return &TransactionQueue{queueLock: sync.Mutex{}, Transactions: make([]SignedTransaction, 0)}
}

func (transactionQueue *TransactionQueue) Copy(original *TransactionQueue) {
	transactionQueue.queueLock.Lock()
	original.queueLock.Lock()

	defer transactionQueue.queueLock.Unlock()
	defer original.queueLock.Unlock()

	transactionQueue.Transactions = make([]SignedTransaction, len(original.Transactions))
	copy(transactionQueue.Transactions, original.Transactions)
}

func (transactionQueue *TransactionQueue) Size() int {
	return len(transactionQueue.Transactions)
}

func (transactionQueue *TransactionQueue) Push(signedTransaction SignedTransaction) {
	transactionQueue.queueLock.Lock()
	defer transactionQueue.queueLock.Unlock()

	transactionQueue.Transactions = append(transactionQueue.Transactions, signedTransaction)
}

func (transactionQueue *TransactionQueue) Pop() (SignedTransaction, error) {
	transactionQueue.queueLock.Lock()
	defer transactionQueue.queueLock.Unlock()

	length := len(transactionQueue.Transactions)
	if length == 0 {
		return SignedTransaction{}, errors.New("empty transactionQueue")
	}

	result := transactionQueue.Transactions[0]
	transactionQueue.Transactions = transactionQueue.Transactions[1:]
	return result, nil
}
