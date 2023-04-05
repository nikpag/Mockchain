package main

import (
	"fmt"
	mathrand "math/rand"
	"time"
)

func (node *Node) walletBalance(id string) uint {
	neighbor, ok := node.nodeDataMap[id]
	if !ok {
		return 0
	}

	address := neighbor.PublicKey

	var amount uint = 0
	for _, UTXO := range node.UTXOsCommitted {
		if equal(address, UTXO.RecipientAddress) {
			amount += UTXO.Amount
		}
	}
	return amount
}

func (node *Node) sendFunds(id string, coins uint) bool {
	time.Sleep(time.Millisecond*100 + time.Millisecond*time.Duration(mathrand.Intn(NUM_NODES))*200)

	node.mineLock.Lock()

	transaction, err := node.createTransaction(id, coins)
	if err != nil {
		node.mineLock.Unlock()
		return false
	}

	signedTransaction := node.signTransaction(transaction)

	node.broadcastTransaction(signedTransaction)

	node.transactionQueue.Push(signedTransaction)

	node.mineLock.Unlock()

	return true
}

func (node *Node) viewAllTransactions() {
	tempTransactions := NewTransactionQueue()
	tempTransactions.Copy(&node.transactionQueue)
	transaction, err := tempTransactions.Pop()
	for err == nil {
		fmt.Println("Transaction from", node.getId(transaction.SenderAddress), "to", node.getId(transaction.ReceiverAddress), "for", transaction.Amount)
		transaction, err = tempTransactions.Pop()
	}
}

func (node *Node) viewLastBlockTransactions() {
	lastBlock := node.blockchain[len(node.blockchain)-1]
	fmt.Println("")

	for index, transaction := range lastBlock.Transactions {
		fmt.Println("Transaction", index)
		fmt.Println("Amount:", transaction.Amount)

		var sender, receiver string
		for id, nghb := range node.nodeDataMap {
			if equal(nghb.PublicKey, transaction.ReceiverAddress) {
				receiver = id
			} else if equal(nghb.PublicKey, transaction.SenderAddress) {
				sender = id
			}
		}

		fmt.Println("From:", sender)
		fmt.Println("To:", receiver)
		fmt.Println("")
	}
}
