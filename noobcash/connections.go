package main

import (
	"log"
	"net"
	"strconv"
	"time"
)

func (node *Node) connectionStart(id string, remoteAddress string) {

	connection, err := net.Dial("tcp", remoteAddress)
	if err != nil {
		log.Fatal("connectionStart:", err)
	}

	if id != "id0" {
		node.sendNewConnectionMessage(connection)
	}

	node.connectionMap[id] = net.Conn(connection)

	node.monitorConnection(connection, id)
}

func (node *Node) monitorConnection(connection net.Conn, connectionID string) {
	for {
		message, err := receiveMessage(connection)
		if err != nil {
			continue
		}

		switch message.MessageType {

		case NullMessageType:
			continue

		case WelcomeMessageType:
			node.receiveWelcomeMessage(message.MessageData)

			node.sendMyInfoMessage(connection)

		case MyInfoMessageType:
			_ = node.receiveMyInfoMessage(message.MessageData)

			node.broadcastType = NodeDataMessageType

			node.broadcastLock.Lock()
			node.broadcast <- true

		case NodeDataMessageType:
			node.receiveNeighborsMessage(message.MessageData)

		case TransactionMessageType:
			node.getTransactionMessage(message.MessageData)

		case BlockMessageType:
			node.getBlockMessage(message.MessageData)

		case ResolveRequestMessageType:
			node.getResolveRequestMessage(message.MessageData, connection)

		case ResolveResponseMessageType:
			node.receiveResolveResponseMessage(message.MessageData)

		default:
			continue
		}
	}
}

func (node *Node) getResolveRequestMessage(data []byte, connection net.Conn) {
	resolveResponseMessage := node.receiveResolveRequestMessage(data)
	node.sendResolveResponseMessage(connection, resolveResponseMessage)
}

func (node *Node) getTransactionMessage(data []byte) {
	signedTransaction := node.receiveTransactionMessage(data)

	node.transactionQueue.Push(signedTransaction)
}

func (node *Node) getBlockMessage(data []byte) {
	hashedBlock := node.receiveBlockMessage(data)

	node.blockchainLock.Lock()

	if node.validateBlock(hashedBlock) {
		log.Println("Block with index", hashedBlock.Index, "validated")
	}

	node.blockchainLock.Unlock()
}

func (node *Node) broadcastMessages() {
	for {
		flag := <-node.broadcast

		if flag {

			for _, connection := range node.connectionMap {

				switch node.broadcastType {

				case NodeDataMessageType:
					node.sendNodeDataMessage(connection)

				case TransactionMessageType:
					node.sendTransactionMessage(connection, node.initiatedTransaction)

				case BlockMessageType:
					node.sendBlockMessage(connection, node.minedBlock)

				case ResolveRequestMessageType:
					node.resolvingConflict = true
					node.sendResolveRequestMessage(connection, node.resolveRequestMessage)
					time.Sleep(time.Millisecond * 200)

				default:
					log.Fatal("broadcastMessages: Trying to broadcast unknown message type")
				}
			}

			node.resolvingConflict = false
			node.broadcastLock.Unlock()
		}
	}
}

func (node *Node) acceptConnectionsBootstrap(listener net.Listener) {
	currentID := 1

	for {
		connection, err := listener.Accept()
		if err != nil {
			continue
		}

		node.connectionMap["id"+strconv.Itoa(currentID)] = connection

		go node.monitorConnection(connection, "id"+strconv.Itoa(currentID))

		node.sendWelcomeMessage(connection, "id"+strconv.Itoa(currentID))

		if currentID == NUM_NODES-1 {
			time.Sleep(time.Millisecond * 100)

			genesis := node.createGenesisBlock()
			node.broadcastBlock(genesis)

			for id := range node.nodeDataMap {
				if id == node.id {
					continue
				}
				time.Sleep(time.Second * 2)
				for i := 0; i < CAPACITY; i++ {
					if !node.sendFunds(id, 100/CAPACITY) {
						log.Fatal("Couldn't create first transaction to node", id)
					}
				}
			}
		}

		currentID++
	}
}

func (node *Node) acceptConnections(listener net.Listener) {
	for count := 0; count < NUM_NODES-2; count++ {
		connection, err := listener.Accept()
		if err != nil {
			continue
		}
		message, err := receiveMessage(connection)
		if err != nil {
			continue
		}

		id := receiveNewConnectionMessage(message.MessageData)
		node.connectionMap[id] = connection

		go node.monitorConnection(connection, id)
	}
}

func (node *Node) establishConnections() {
	myID, err := strconv.Atoi(node.id[2:])
	if err != nil {
		log.Fatal("establishConnections:", err)
	}

	for id, nodeData := range node.nodeDataMap {
		targetID, err := strconv.Atoi(id[2:])
		if err != nil {
			log.Fatal("establishConnections:", err)
		}

		if _, ok := node.connectionMap[id]; ok {
			continue

		} else if myID > targetID {
			go node.connectionStart(id, nodeData.Address)
		}
	}
}
