package main

import (
	"fmt"
	"log"
	"strconv"
)

func (node *Node) cli(fields []string) {
	if len(fields) == 3 && fields[0] == "t" {
		if _, ok := node.nodeDataMap[fields[1]]; !ok {
			fmt.Println("No client with id", fields[1])
			node.help()
			return
		}

		parsed, err := strconv.ParseUint(fields[2], 10, 32)
		if err != nil {
			log.Println("cli: ParseUint", err)
			node.help()
			return
		}
		amount := uint(parsed)

		node.transaction(fields[1], amount)

	} else if len(fields) == 1 {
		if fields[0] == "view" {
			node.view()
		} else if fields[0] == "balance" {
			node.balance()
		} else if fields[0] == "help" {
			node.help()
		} else if fields[0] == "txs" {
			node.viewAllTransactions()
		} else if fields[0] == "all" {
			node.allBalances()
		} else if fields[0] == "hashes" {
			node.hashes()
		} else {
			node.help()
		}
	} else {
		node.help()
	}
}

func (node *Node) transaction(id string, amount uint) {
	node.sendFunds(id, amount)
}

func (node *Node) view() {
	node.viewLastBlockTransactions()
}
func (node *Node) balance() {
	fmt.Println("\nWallet balance is:", node.walletBalance(node.id))
}
func (node *Node) help() {
	fmt.Println("\nsupported commands:")
	fmt.Println("")

	fmt.Println("t <receiverID> <amount>")
	fmt.Println("\tsends <amount> coins to client with <receiverID>")
	fmt.Println("\tExample: t id1 100")
	fmt.Println("")

	fmt.Println("view")
	fmt.Println("\tshows details of transactions in the latest block")
	fmt.Println("")

	fmt.Println("balance")
	fmt.Println("\treturns the client's remaining NBC coins")
	fmt.Println("")

	fmt.Println("txs")
	fmt.Println("\treturns all transactions received by this node in chronological order")
	fmt.Println("")

	fmt.Println("all")
	fmt.Println("\treturns the amount of coins in each of the system's nodes")
	fmt.Println("")

	fmt.Println("hashes")
	fmt.Println("\treturns the hashes of all blocks in the blockchain")
	fmt.Println("")

	fmt.Println("help")
	fmt.Println("\thelp about cli commands")
}

func (node *Node) allBalances() {
	fmt.Println("\nThe balances of all wallets are:")
	for id := range node.nodeDataMap {
		fmt.Println(id, node.walletBalance(id))
	}
}

func (node *Node) hashes() {
	fmt.Println("\nThe hashes of the blocks in the blockchain are:")
	node.blockchainLock.Lock()
	for _, block := range node.blockchain {
		fmt.Println("Block", block.Index, "with hash", block.Hash)
	}
	node.blockchainLock.Unlock()
}
