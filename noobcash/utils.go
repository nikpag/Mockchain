package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func equal(publicKey1, publicKey2 rsa.PublicKey) bool {
	if publicKey1.N.Cmp(publicKey2.N) == 0 && publicKey1.E == publicKey2.E {
		return true
	} else {
		return false
	}
}

func generateRandom32Byte() [32]byte {
	var nonce [32]byte
	_, _ = rand.Read(nonce[:])
	return nonce
}

func (node *Node) test() {
	if node.id == "id0" {
		http.HandleFunc("/transaction/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Transaction hit")
			pathArgs := strings.Split(r.URL.Path, "/")

			receiver := pathArgs[3]
			amount, err := strconv.Atoi(pathArgs[4])
			if err != nil {
				w.WriteHeader(500)
				return
			}

			node.transaction(receiver, uint(amount))

			w.WriteHeader(200)
		})

		http.HandleFunc("/balance", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Balance hit")
			w.WriteHeader(200)
			w.Write([]byte(strconv.Itoa(int(node.walletBalance(node.id)))))
		})

		http.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) {
			tempTransactions := NewTransactionQueue()
			tempTransactions.Copy(&node.transactionQueue)
			w.WriteHeader(200)

			response := ""

			response += "{\"transactionList\": ["

			transaction, err := tempTransactions.Pop()

			for i := 0; err == nil; i++ {
				var _fromTo string
				var _node string

				if node.getId(transaction.SenderAddress) == "id0" {
					_fromTo = "to"
					_node = node.getId(transaction.ReceiverAddress)
				} else if node.getId(transaction.ReceiverAddress) == "id0" {
					_fromTo = "from"
					_node = node.getId(transaction.SenderAddress)
				} else {
					continue
				}

				_amount := transaction.Amount

				response += fmt.Sprintf("{\"fromTo\": \"%s\", \"node\": \"%s\", \"amount\": %d},", _fromTo, _node, _amount)
				transaction, err = tempTransactions.Pop()
			}

			response = response[:len(response)-1]

			response += "]}"

			w.Write([]byte(response))

		})

		go http.ListenAndServe(":58080", nil)
	}
}
