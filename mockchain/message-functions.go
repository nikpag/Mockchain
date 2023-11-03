package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
)

func sendMessage(connection net.Conn, message Message) {
	messageJSON, _ := json.Marshal(message)

	connection.Write(messageJSON)
	connection.Write([]byte("\n"))
}

func (node *Node) sendWelcomeMessage(connection net.Conn, id string) {

	node.nodeDataMap[id] = new(NodeData)

	welcomeMessage := WelcomeMessage{id}

	welcomeMessageJSON, err := json.Marshal(welcomeMessage)
	if err != nil {
		log.Fatal("sendWelcomeMessage:", err)
	}

	message := Message{
		WelcomeMessageType,
		welcomeMessageJSON,
	}

	sendMessage(connection, message)
}

func (node *Node) sendNewConnectionMessage(connection net.Conn) {

	newConnectionMessage := NewConnectionMessage{node.id}

	newConnectionMessageJSON, err := json.Marshal(newConnectionMessage)
	if err != nil {
		log.Fatal("sendNewConnectionMessage:", err)
	}

	message := Message{
		NewConnectionMessageType,
		newConnectionMessageJSON,
	}

	sendMessage(connection, message)
}

func (node *Node) sendMyInfoMessage(connection net.Conn) {

	myInfoMessage := MyInfoMessage{
		ID:        node.id,
		PublicKey: node.publicKey,
		Address:   node.address}

	myInfoMessageJSON, err := json.Marshal(myInfoMessage)
	if err != nil {
		log.Fatal("sendMyInfoMessage:", err)
	}

	message := Message{
		MessageType: MyInfoMessageType,
		MessageData: myInfoMessageJSON,
	}

	sendMessage(connection, message)
}

func (node *Node) sendNodeDataMessage(connection net.Conn) {

	nodeDataList := make(map[string]NodeData)
	for id, nodeData := range node.nodeDataMap {
		nodeDataList[id] = NodeData{Address: nodeData.Address, PublicKey: nodeData.PublicKey}
	}

	nodeDataListMessage := NodeDataMessage{
		nodeDataList,
	}

	nodeDataListMessageJSON, err := json.Marshal(nodeDataListMessage)
	if err != nil {
		return
	}

	message := Message{
		MessageType: NodeDataMessageType,
		MessageData: nodeDataListMessageJSON,
	}

	sendMessage(connection, message)
}

func (node *Node) sendTransactionMessage(connection net.Conn, signedTranscation SignedTransaction) {

	transactionMessage := TransactionMessage{signedTranscation}

	transactionMessageJSON, err := json.Marshal(transactionMessage)
	if err != nil {
		return
	}

	message := Message{
		MessageType: TransactionMessageType,
		MessageData: transactionMessageJSON,
	}

	sendMessage(connection, message)
}

func (node *Node) sendBlockMessage(connections net.Conn, hashedBlock HashedBlock) {
	blockMessage := BlockMessage{
		hashedBlock,
	}

	blockMessageJSON, err := json.Marshal(blockMessage)
	if err != nil {
		return
	}

	message := Message{
		MessageType: BlockMessageType,
		MessageData: blockMessageJSON,
	}

	sendMessage(connections, message)
}

func (node *Node) sendResolveRequestMessage(connection net.Conn, resolveRequestMessage ResolveRequestMessage) {
	resolveRequestMessageJSON, err := json.Marshal(resolveRequestMessage)
	if err != nil {
		return
	}

	message := Message{
		MessageType: ResolveRequestMessageType,
		MessageData: resolveRequestMessageJSON,
	}

	sendMessage(connection, message)
}

func (node *Node) sendResolveResponseMessage(connection net.Conn, resolveResponseMessage ResolveResponseMessage) {
	resolveResponseMessageJSON, err := json.Marshal(resolveResponseMessage)
	if err != nil {
		return
	}

	message := Message{
		MessageType: ResolveResponseMessageType,
		MessageData: resolveResponseMessageJSON,
	}

	sendMessage(connection, message)
}

func receiveMessage(connection net.Conn) (Message, error) {
	messageBytes, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		return Message{}, err
	}

	var message Message
	err = json.Unmarshal([]byte(messageBytes), &message)
	if err != nil {
		return Message{}, err
	}

	return message, nil
}

func (node *Node) receiveNeighborsMessage(neighborsMessageBytes []byte) {

	var neighborsMessage NodeDataMessage
	err := json.Unmarshal(neighborsMessageBytes, &neighborsMessage)
	if err != nil {
		return
	}

	for id, nodeData := range neighborsMessage.Neighbors {

		if _, ok := node.nodeDataMap[id]; ok {
			continue

		} else {
			node.nodeDataMap[id] = &NodeData{Address: nodeData.Address, PublicKey: nodeData.PublicKey}
		}
	}

	node.establishConnections()
}

func (node *Node) receiveMyInfoMessage(myInfoMessageBytes []byte) string {

	var myInfoMessage MyInfoMessage
	err := json.Unmarshal(myInfoMessageBytes, &myInfoMessage)
	if err != nil {
		return ""
	}

	node.nodeDataMap[myInfoMessage.ID].PublicKey = myInfoMessage.PublicKey
	node.nodeDataMap[myInfoMessage.ID].Address = myInfoMessage.Address

	return myInfoMessage.ID
}

func (node *Node) receiveWelcomeMessage(welcomeMessageBytes []byte) {

	var welcomeMessage WelcomeMessage
	err := json.Unmarshal(welcomeMessageBytes, &welcomeMessage)
	if err != nil {
		return
	}

	node.id = welcomeMessage.ID

	node.nodeDataMap[node.id] = &NodeData{
		PublicKey: node.publicKey,
		Address:   node.address,
	}
}

func (node *Node) receiveTransactionMessage(transactionMessageBytes []byte) SignedTransaction {
	var transactionMessage TransactionMessage
	err := json.Unmarshal(transactionMessageBytes, &transactionMessage)
	if err != nil {
		return SignedTransaction{}
	}

	return transactionMessage.Transaction
}

func (node *Node) receiveBlockMessage(blockMessageBytes []byte) HashedBlock {
	var blockMessage BlockMessage
	err := json.Unmarshal(blockMessageBytes, &blockMessage)
	if err != nil {
		return HashedBlock{}
	}

	return blockMessage.Block
}

func (node *Node) receiveResolveRequestMessage(receivedResolveRequestMessageBytes []byte) ResolveResponseMessage {
	var resolveRequestMessage ResolveRequestMessage

	err := json.Unmarshal(receivedResolveRequestMessageBytes, &resolveRequestMessage)

	if err != nil {
		return ResolveResponseMessage{}
	}

	node.blockchainLock.Lock()

	if resolveRequestMessage.ChainSize < uint(len(node.blockchain)) {
		blocks := make([]HashedBlock, 0)

		var i int
		for i = range resolveRequestMessage.Hashes {
			if resolveRequestMessage.Hashes[i] != node.blockchain[i].Hash {
				break
			}
		}
		for ; i < len(node.blockchain); i++ {
			blocks = append(blocks, node.blockchain[i])
		}

		resolveResponseMessage := ResolveResponseMessage{
			Blocks: blocks,
		}

		node.blockchainLock.Unlock()
		return resolveResponseMessage
	}

	node.blockchainLock.Unlock()
	return ResolveResponseMessage{}
}

func (node *Node) receiveResolveResponseMessage(receiveResolveResponseMessageBytes []byte) {
	var resolveResponseMessage ResolveResponseMessage
	err := json.Unmarshal(receiveResolveResponseMessageBytes, &resolveResponseMessage)
	if err != nil {
		return
	}

	if len(resolveResponseMessage.Blocks) == 0 {
		return
	}

	firstIndex := resolveResponseMessage.Blocks[0].Index
	lastIndex := resolveResponseMessage.Blocks[len(resolveResponseMessage.Blocks)-1].Index

	if lastIndex >= uint(len(node.blockchain)) {
		temp := make([]HashedBlock, firstIndex)
		copy(temp[0:firstIndex], node.blockchain)

		temp = append(temp, resolveResponseMessage.Blocks...)

		node.blockchain = make([]HashedBlock, lastIndex+1)
		copy(node.blockchain, temp)

		node.validateChain()
	}

}

func receiveNewConnectionMessage(newConnectionMessageBytes []byte) string {
	var newConnectionMessage NewConnectionMessage
	err := json.Unmarshal(newConnectionMessageBytes, &newConnectionMessage)
	if err != nil {
		return ""
	}

	return newConnectionMessage.ID
}
