package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const NUM_NODES = 5
const DIFFICULTY = 4 // number of hex digits to be zero on the start of a block's hash
const CAPACITY = 1

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing arguments. Run './noobcash.elf -h' for help")
		return
	}

	var node Node

	localAddress := flag.String("localAddress", "localhost:50000", "This node's ip:port")
	bootstrapAddress := flag.String("bootstrapAddress", "localhost:50000", "Bootstrap node's ip:port")
	transactionFile := flag.String("transactionFile", "none", "File containing this node's transactions")
	isBootstrap := flag.Bool("isBootstrap", false, "True if this node is bootstrap")

	flag.Parse()

	node.createNode(*localAddress)

	go node.collectTransactions()

	if *isBootstrap {
		node.id = "id0"
		go node.startBootstrap(*localAddress)
	} else {
		go node.startOrdinaryNode(*localAddress)
		go node.connectionStart("id0", *bootstrapAddress)
	}

	for {
		node.blockchainLock.Lock()
		if len(node.blockchain) > 0 {
			node.blockchainLock.Unlock()
			break
		}
		node.blockchainLock.Unlock()
	}

	startTime := time.Now().Unix()

	for len(node.blockchain) < NUM_NODES {
		continue
	}

	fmt.Println("All clients have 100 coins")

	if *transactionFile != "none" {

		file, err := os.Open(*transactionFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			text := scanner.Text()
			fields := strings.Fields(text)

			if len(fields) < 2 {
				break
			}

			if _, ok := node.nodeDataMap[fields[0]]; !ok {
				fmt.Println("No node with id", fields[0])
				break
			}

			parsed, err := strconv.ParseUint(fields[1], 10, 32)
			if err != nil {
				log.Println("cli: ParseUint", err)
				break
			}
			amount := uint(parsed)

			node.transaction(fields[0], amount)
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		endTime := time.Now().Unix()
		blockTime := float32(endTime-startTime) / float32(len(node.blockchain))
		fmt.Println("\n\n##########################################################")
		fmt.Println("Blocks:", len(node.blockchain))
		fmt.Println("Time:", endTime-startTime, "sec")
		fmt.Println("Average block time:", blockTime, "sec")
		fmt.Println("Throughput:", float32(CAPACITY)/blockTime)
		fmt.Println("##########################################################")
		fmt.Println("")
		fmt.Println("")
	}

	node.test()

	for {
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		fields := strings.Fields(line)
		node.cli(fields)
	}
}
