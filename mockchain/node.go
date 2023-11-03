package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"sync"
	"time"
)

type Node struct {
	// Node Data
	id         string
	privateKey rsa.PrivateKey
	publicKey  rsa.PublicKey
	address    string

	nodeDataMap   map[string]*NodeData
	connectionMap map[string]net.Conn

	blockchain     []HashedBlock
	blockchainLock sync.Mutex
	mineLock       sync.Mutex

	resolvingConflict bool

	UTXOsValidated     map[[32]byte]TransactionOutput
	UTXOsSoftValidated map[[32]byte]TransactionOutput
	UTXOsCommitted     map[[32]byte]TransactionOutput

	myUTXOs UTXOStack

	transactionQueue TransactionQueue
	usedTransactions map[[32]byte]bool

	broadcast             chan bool
	broadcastLock         sync.Mutex
	broadcastType         MessageType
	initiatedTransaction  SignedTransaction
	minedBlock            HashedBlock
	resolveRequestMessage ResolveRequestMessage
}

type NodeData struct {
	PublicKey rsa.PublicKey
	Address   string
}

func (node *Node) createNode(localAddress string) {

	node.generateWallet()
	node.address = localAddress

	node.nodeDataMap = make(map[string]*NodeData)
	node.connectionMap = make(map[string]net.Conn)

	node.blockchain = make([]HashedBlock, 0)
	node.blockchainLock = sync.Mutex{}
	node.mineLock = sync.Mutex{}

	node.resolvingConflict = false

	node.UTXOsValidated = make(map[[32]byte]TransactionOutput)
	node.UTXOsSoftValidated = make(map[[32]byte]TransactionOutput)
	node.UTXOsCommitted = make(map[[32]byte]TransactionOutput)

	node.myUTXOs = *NewStack()

	node.transactionQueue = *NewTransactionQueue()
	node.usedTransactions = make(map[[32]byte]bool)

	node.broadcastLock = sync.Mutex{}
	node.broadcast = make(chan bool)
}

func (node *Node) startBootstrap(localAddress string) {
	node.nodeDataMap[node.id] = &NodeData{
		PublicKey: node.publicKey,
		Address:   node.address,
	}
	listener, _ := net.Listen("tcp", localAddress)
	defer listener.Close()
	go node.acceptConnectionsBootstrap(listener)

	node.broadcastMessages()
}

func (node *Node) startOrdinaryNode(localAddress string) {
	listener, _ := net.Listen("tcp", localAddress)
	defer listener.Close()
	go node.acceptConnections(listener)
	node.broadcastMessages()
}

func (node *Node) createGenesisBlock() HashedBlock {
	var transactionID, bootstrapTransactionOutputID, magicTransactionOutputID [32]byte

	transactionID = generateRandom32Byte()
	bootstrapTransactionOutputID = generateRandom32Byte()
	magicTransactionOutputID = generateRandom32Byte()

	bootstrapTransactionOutput := TransactionOutput{
		ID:               bootstrapTransactionOutputID,
		TransactionID:    transactionID,
		RecipientAddress: node.publicKey,
		Amount:           100 * NUM_NODES,
	}
	magicTransactionOutput := TransactionOutput{
		ID:               magicTransactionOutputID,
		TransactionID:    transactionID,
		RecipientAddress: rsa.PublicKey{N: big.NewInt(0), E: 1},
		Amount:           0,
	}

	transaction := SignedTransaction{
		SenderAddress:      rsa.PublicKey{N: big.NewInt(0), E: 1},
		ReceiverAddress:    node.publicKey,
		Amount:             100 * NUM_NODES,
		TransactionID:      transactionID,
		TransactionInputs:  []TransactionInput{},
		TransactionOutputs: [2]TransactionOutput{bootstrapTransactionOutput, magicTransactionOutput},
		Signature:          []byte{0},
	}

	var nonce, previousHash [32]byte

	for i := range nonce {
		nonce[i] = 0
		previousHash[i] = 0
	}
	previousHash[31] = 1

	block := Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		PreviousHash: previousHash,
		Transactions: [CAPACITY]SignedTransaction{transaction},
		Nonce:        nonce,
	}

	blockJSON, _ := json.Marshal(block)

	hash := sha256.Sum256(blockJSON)

	hashedBlock := HashedBlock{
		Index:        block.Index,
		Timestamp:    block.Timestamp,
		PreviousHash: block.PreviousHash,
		Transactions: block.Transactions,
		Nonce:        block.Nonce,
		Hash:         hash,
	}

	return hashedBlock
}

func (node *Node) generateWallet() {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	node.privateKey = *privateKey
	node.publicKey = privateKey.PublicKey
}

func (node *Node) createTransaction(receiverID string, amount uint) (Transaction, error) {

	if amount <= 0 {
		return Transaction{}, errors.New("InvalidTransactionAmount")
	}
	if _, ok := node.nodeDataMap[receiverID]; !ok {
		return Transaction{}, errors.New("InvalidReceiverID")
	}

	var totalCredits uint = 0
	var UTXOs []TransactionOutput

	/*
		We need to go over our wallet to see if we have enough funds to do the transaction. If not, we have to
		return an empty transaction and re-add the UTXOs that we have checked so far.
	*/
	for totalCredits < amount {
		UTXO, err := node.myUTXOs.Pop()

		if err != nil {
			log.Println("createTransaction: Funds not enough to create transaction with", amount, "to", receiverID)
			// On god mhn ksexasete pote to _, efaga 25 lepta debugging giati nomiza oti h Pop() epestrefe int
			// enw htan to index gia to range. Mia xara htan h C, hkseran ti ekanan 50 xronia twra
			for _, UTXO := range UTXOs {
				node.myUTXOs.Push(UTXO)
			}
			return Transaction{}, errors.New("not enough coins")
		}

		totalCredits += UTXO.Amount
		UTXOs = append(UTXOs, UTXO)
	}

	recipientPublicKey := node.nodeDataMap[receiverID].PublicKey

	var transactionID, senderTransactionOutput, receiverTransactionOutput [32]byte

	transactionID = generateRandom32Byte()
	senderTransactionOutput = generateRandom32Byte()
	receiverTransactionOutput = generateRandom32Byte()

	var transactionInputs []TransactionInput
	for _, UTXO := range UTXOs {
		transactionInputs = append(transactionInputs, TransactionInput{UTXO.ID})
	}

	transactionOutputs := [2]TransactionOutput{
		{
			ID:               senderTransactionOutput,
			TransactionID:    transactionID,
			RecipientAddress: node.publicKey,
			Amount:           totalCredits - amount},
		{
			ID:               receiverTransactionOutput,
			TransactionID:    transactionID,
			RecipientAddress: recipientPublicKey,
			Amount:           amount}}

	if !node.resolvingConflict {
		node.myUTXOs.Push(transactionOutputs[0])
	}

	transaction := Transaction{
		SenderAddress:      node.publicKey,
		ReceiverAddress:    recipientPublicKey,
		Amount:             amount,
		TransactionID:      transactionID,
		TransactionInputs:  transactionInputs,
		TransactionOutputs: transactionOutputs,
	}
	return transaction, nil
}

/*
We need to sign the transaction with our private key. Whoever receives this transaction
must verify it's signature using the sender's public key.
*/
func (node *Node) signTransaction(transaction Transaction) SignedTransaction {
	transactionJSON, err := json.Marshal(transaction)
	if err != nil {
		log.Println("signTransaction: Marshal->", err)
		return SignedTransaction{}
	}

	transactionJSONHash := sha256.Sum256(transactionJSON)

	signature, _ := rsa.SignPKCS1v15(rand.Reader, &node.privateKey, crypto.SHA256, transactionJSONHash[:])
	if err != nil {
		log.Println("Setup: Sign", err)
		return SignedTransaction{}
	}

	signedTransaction := SignedTransaction{
		SenderAddress:      transaction.SenderAddress,
		ReceiverAddress:    transaction.ReceiverAddress,
		Amount:             transaction.Amount,
		TransactionID:      transaction.TransactionID,
		TransactionInputs:  transaction.TransactionInputs,
		TransactionOutputs: transaction.TransactionOutputs,
		Signature:          signature,
	}

	return signedTransaction
}

/*
Broadcast the signed transcation to every user connected in the blockchain
*/
func (node *Node) broadcastTransaction(transaction SignedTransaction) {
	node.broadcastLock.Lock()

	node.initiatedTransaction = transaction
	node.broadcastType = TransactionMessageType
	node.broadcast <- true //channel receiver
}

/*
Verify the signature of the sender using his public key
*/
func (node *Node) verifySignature(signedTransaction SignedTransaction) bool {
	transaction := Transaction{
		SenderAddress:      signedTransaction.SenderAddress,
		ReceiverAddress:    signedTransaction.ReceiverAddress,
		Amount:             signedTransaction.Amount,
		TransactionID:      signedTransaction.TransactionID,
		TransactionInputs:  signedTransaction.TransactionInputs,
		TransactionOutputs: signedTransaction.TransactionOutputs,
	}

	transactionJSON, err := json.Marshal(transaction)
	if err != nil {
		log.Fatal("verifySignature: Marshal->", err)
	}

	transactionJSONHash := sha256.Sum256(transactionJSON)

	err = rsa.VerifyPKCS1v15(&transaction.SenderAddress, crypto.SHA256, transactionJSONHash[:], signedTransaction.Signature)
	if err != nil {
		return false
	} else {
		return true
	}
}

/*
Check the validity of the transaction. We need to make sure that:
 1. Signature of the sender is valid
 2. Transaction Inputs have not been already spent. This check is important to
    eliminmate "double spending".
 3. Sender and Receiver are memebers of the blockchain network
 4. Transaction Output members exist and are the correct sender and receiver
 5. Amount that is being sent actually exists in the Sender's UTXO Wallet
*/
func (node *Node) validateTransaction(signedTransaction SignedTransaction) bool {
	if !node.verifySignature(signedTransaction) {
		return false
	}

	//Check if Sender and Receiver are found in the network
	foundSender, foundReceiver := false, false

	for _, nodeData := range node.nodeDataMap {
		if equal(nodeData.PublicKey, signedTransaction.SenderAddress) {
			foundSender = true
		} else if equal(nodeData.PublicKey, signedTransaction.ReceiverAddress) {
			foundReceiver = true
		}
	}
	if !foundSender || !foundReceiver {
		return false
	}

	// Check if sender actually has the amount needed to do the transaction
	inputSum := uint(0)
	for _, transactionInput := range signedTransaction.TransactionInputs {
		if transactionOutput, ok := node.UTXOsCommitted[transactionInput.PreviousOutputID]; !ok {
			return false
		} else {
			inputSum += transactionOutput.Amount
		}
	}
	// Amounts should match
	outputSum := signedTransaction.TransactionOutputs[0].Amount + signedTransaction.TransactionOutputs[1].Amount
	if inputSum != outputSum {
		return false
	}

	// Check if UTXO's sender and receiver match
	outputsOK := false
	if equal(signedTransaction.TransactionOutputs[0].RecipientAddress, signedTransaction.ReceiverAddress) {
		if equal(signedTransaction.TransactionOutputs[1].RecipientAddress, signedTransaction.SenderAddress) {
			outputsOK = true
		}
	} else if equal(signedTransaction.TransactionOutputs[0].RecipientAddress, signedTransaction.SenderAddress) {
		if equal(signedTransaction.TransactionOutputs[1].RecipientAddress, signedTransaction.ReceiverAddress) {
			outputsOK = true
		}
	}
	if !outputsOK {
		return false
	}

	/*
		Update UTXOs. This Update must be done by everyone so that each node has a correct
		view on everyone's UTXOs
	*/

	// Delete only the ones found in the TransactionInput list :)
	for _, transactionInput := range signedTransaction.TransactionInputs {
		delete(node.UTXOsCommitted, transactionInput.PreviousOutputID)
	}

	senderTransactionOutput := &signedTransaction.TransactionOutputs[0]
	receiverTransactionOutput := &signedTransaction.TransactionOutputs[1]

	/*
		Add only the receiver transaction output to the stack. The sender transaction output has already been added
		In the transaction generation process
	*/
	if equal(node.publicKey, receiverTransactionOutput.RecipientAddress) && !node.resolvingConflict {
		node.myUTXOs.Push(*receiverTransactionOutput)
	}

	// Commit the UTXOs created by the transaction
	node.UTXOsCommitted[senderTransactionOutput.ID] = *senderTransactionOutput
	node.UTXOsCommitted[receiverTransactionOutput.ID] = *receiverTransactionOutput

	return true
}

/*
This function makes a validation run on all the transactions selected to be mined. The UTXOs
that are affiliated with this transaction are not commited to the global UTXO commited map.
This is only done to make sure that the transactions that are added to the blockToBeMined are valid.
*/
func (node *Node) softValidateTransaction(signedTransaction SignedTransaction) bool {
	if !node.verifySignature(signedTransaction) {
		return false
	}

	//Check if Sender and Receiver are found in the network
	foundSender := false
	foundReceiver := false
	for _, nodeData := range node.nodeDataMap {
		if equal(nodeData.PublicKey, signedTransaction.SenderAddress) {
			foundSender = true
		} else if equal(nodeData.PublicKey, signedTransaction.ReceiverAddress) {
			foundReceiver = true
		}
	}
	if !foundSender || !foundReceiver {
		return false
	}

	// Check if sender actually has the amount needed to do the transaction
	inputSum := uint(0)
	for _, transactionInput := range signedTransaction.TransactionInputs {
		if transactionOutput, ok := node.UTXOsSoftValidated[transactionInput.PreviousOutputID]; !ok {
			return false
		} else {
			inputSum += transactionOutput.Amount
		}
	}

	// Amounts should match
	outputSum := signedTransaction.TransactionOutputs[0].Amount + signedTransaction.TransactionOutputs[1].Amount
	if inputSum != outputSum {
		return false
	}

	// Check if UTXO's sender and receiver match
	outputsOK := false
	if equal(signedTransaction.TransactionOutputs[0].RecipientAddress, signedTransaction.ReceiverAddress) {
		if equal(signedTransaction.TransactionOutputs[1].RecipientAddress, signedTransaction.SenderAddress) {
			outputsOK = true
		}
	} else if equal(signedTransaction.TransactionOutputs[0].RecipientAddress, signedTransaction.SenderAddress) {
		if equal(signedTransaction.TransactionOutputs[1].RecipientAddress, signedTransaction.ReceiverAddress) {
			outputsOK = true
		}
	}
	if !outputsOK {
		return false
	}

	/*
		Update UTXOs. This Update must be done by everyone so that each node has a correct
		view on everyone's UTXOs
	*/

	// Remove all the UTXOs that had matching TransactionInput IDs
	for _, transactionInput := range signedTransaction.TransactionInputs {
		delete(node.UTXOsSoftValidated, transactionInput.PreviousOutputID)
	}

	senderTransactionOutput := &signedTransaction.TransactionOutputs[0]
	receiverTransactionOutput := &signedTransaction.TransactionOutputs[1]

	node.UTXOsSoftValidated[senderTransactionOutput.ID] = *senderTransactionOutput
	node.UTXOsSoftValidated[receiverTransactionOutput.ID] = *receiverTransactionOutput

	return true
}

func (node *Node) mineBlock(signedTransaction []SignedTransaction, chainLength uint) HashedBlock {
	fmt.Println("\nMining block", chainLength)

	var transactions [CAPACITY]SignedTransaction
	copy(transactions[:], signedTransaction)

	/*
		Create a basis block using the transactions that we have validated and are using to mine the
		block. Nonce is <nil> in this case and will be generating various values to test the hash.
	*/
	block := Block{
		Index:        chainLength,
		Timestamp:    time.Now().Unix(),
		PreviousHash: node.blockchain[chainLength-1].Hash,
		Transactions: transactions,
	}

	i := 0
	for {
		i++
		if i%10000 == 0 {

			/*
				Periodically check if a new block has been added
				to the blockchain. Go is sadly not fully preemptive,
				we have to do this check by hand every so and often
				as to not continue mining if someone else mines the block
				first.
			*/
			node.blockchainLock.Lock()
			if chainLength < uint(len(node.blockchain)) {
				node.blockchainLock.Unlock()
				return HashedBlock{}
			}
			node.blockchainLock.Unlock()
		}

		/*
			Generate a random nonce number, hash the block and check if it
			matches our difficulty rules.
		*/
		block.Nonce = generateRandom32Byte()

		blockJSON, err := json.Marshal(block)
		if err != nil {
			log.Println("Error while marshaling block")
			continue
		}
		hash := sha256.Sum256(blockJSON)

		mined := true

		/*
			While trying to do the project we run up to this problem:
			The difficulty of the block is the amount of 0s that need to
			be found in the beginning of the SHA256 hash. That means that if
			the difficulty is 5, then the 5 MSBs of the hash must be 0. This turned
			out to be very easy for the node to calculate and caused some synchronization
			issues. Thus, we changed it from DIFFICULTY bits to DIFFICULTY hexadecimals
			(or nibbles, 4-bits).
		*/
		for _, oneByte := range hash[:DIFFICULTY/2] {
			if oneByte != 0 {
				mined = false
				break
			}
		}
		if DIFFICULTY%2 == 1 && hash[DIFFICULTY/2] > 15 {
			mined = false
		}

		if mined {
			log.Printf("Node %v found block!", node.id)

			return HashedBlock{
				Index:        block.Index,
				Timestamp:    block.Timestamp,
				PreviousHash: block.PreviousHash,
				Transactions: block.Transactions,
				Nonce:        block.Nonce,
				Hash:         hash,
			}
		}
	}
}

func (node *Node) broadcastBlock(block HashedBlock) {
	if len(node.blockchain) > int(block.Index) {
		log.Println("broadcastBlock: Aborting broadcast - new block added to chain")
	}

	if !node.validateBlock(block) {
		log.Println("Error: Block not valid")
		return
	}

	node.broadcastLock.Lock()

	node.minedBlock = block
	node.broadcastType = BlockMessageType
	node.broadcast <- true
}

func (node *Node) validateBlock(hashedBlock HashedBlock) bool {

	currentChainLength := uint(len(node.blockchain))

	// Automatically validate genesis block, its accepted as valid every time
	if currentChainLength == 0 && hashedBlock.Index == 0 {
		for _, transaction := range hashedBlock.Transactions {

			if transaction.Amount > 0 {
				transactionOutput1 := &transaction.TransactionOutputs[0]
				transactionOutput2 := &transaction.TransactionOutputs[1]

				// only store in own unspent tokens if bootstrap
				if node.broadcastType != ResolveRequestMessageType {
					if equal(node.publicKey, transactionOutput1.RecipientAddress) && !node.resolvingConflict {
						node.myUTXOs.Push(*transactionOutput1)
					} else if equal(node.publicKey, transactionOutput2.RecipientAddress) && !node.resolvingConflict {
						node.myUTXOs.Push(*transactionOutput2)
					}
				}
				node.UTXOsCommitted[transactionOutput1.ID] = *transactionOutput1
				node.UTXOsCommitted[transactionOutput2.ID] = *transactionOutput2
			}
		}
		node.blockchain = append(node.blockchain, hashedBlock)

		return true

	} else if currentChainLength > hashedBlock.Index {
		// Reject older blocks
		log.Println("Older block received")
		return false
	} else if currentChainLength < hashedBlock.Index {
		// Resolve conflict if there is a gap in the blockchain
		log.Println("Possible gap in chain - calling resolveConflict()")
		node.resolveConflict(hashedBlock.Index)
		return false
	}

	// If the block is the correct one, check its validity by comparing the hashes and the difficulty rule
	block := Block{
		Index:        hashedBlock.Index,
		Timestamp:    hashedBlock.Timestamp,
		PreviousHash: hashedBlock.PreviousHash,
		Transactions: hashedBlock.Transactions,
		Nonce:        hashedBlock.Nonce,
	}

	blockJSON, _ := json.Marshal(block)

	hash := sha256.Sum256(blockJSON)
	if hash != hashedBlock.Hash {
		return false
	}

	for _, oneByte := range hash[:DIFFICULTY/2] {
		if oneByte != 0 {
			return false
		}
	}

	if DIFFICULTY%2 == 1 && hash[DIFFICULTY/2] > 15 {
		return false
	}

	// We possess an invalid blockchain, receive the largest one from someone else
	if hashedBlock.Index > 0 && hashedBlock.PreviousHash != node.blockchain[hashedBlock.Index-1].Hash {
		log.Println("Previous hash not matching the one in the blockchain - calling resolveConflict")
		node.resolveConflict(hashedBlock.Index)
		return false
	}

	// The block is accepted - we need to add it to the blockchain
	node.blockchain = append(node.blockchain, hashedBlock)
	for index, transaction := range hashedBlock.Transactions {
		valid := node.validateTransaction(transaction)
		if !valid {
			log.Println("validateBlock: transaction", index, "invalid")
		}
	}

	return true
}

func (node *Node) validateChain() bool {

	/*
		Check if every block of the blockchain is valid
		This is only called when we receive a blockchain
		while calling resolve conflict.
	*/

	blockchainCopy := make([]HashedBlock, len(node.blockchain))
	copy(blockchainCopy, node.blockchain)
	node.blockchain = make([]HashedBlock, 0)

	node.UTXOsSoftValidated = make(map[[32]byte]TransactionOutput)
	node.UTXOsSoftValidated = make(map[[32]byte]TransactionOutput)
	node.UTXOsCommitted = make(map[[32]byte]TransactionOutput)
	node.usedTransactions = make(map[[32]byte]bool)

	valid := true
	for _, block := range blockchainCopy {
		valid = node.validateBlock(block)
		if !valid {
			break
		}
	}

	if valid {
		log.Println("validateChain: blockchain is valid")
	} else {
		log.Println("validateChain: blockchain is not valid")
	}

	return valid
}

func (node *Node) resolveConflict(index uint) {
	hashes := make([][32]byte, 0)
	for _, block := range node.blockchain {
		hashes = append(hashes, block.Hash)
	}

	length := uint(len(node.blockchain))

	node.broadcastLock.Lock()
	node.resolveRequestMessage = ResolveRequestMessage{
		ChainSize: length,
		Hashes:    hashes,
	}
	node.broadcastType = ResolveRequestMessageType
	node.broadcast <- true
	for int(index) > len(node.blockchain) {
		continue
	}
}

/*
This is state in the program where we wait to receive the necessary *capacity* transactions
to start mining the block.
*/
func (node *Node) collectTransactions() {
	for {
		node.blockchainLock.Lock()
		if len(node.blockchain) > 0 {
			node.blockchainLock.Unlock()
			break
		}
		node.blockchainLock.Unlock()
	}

	node.updateUncommitted()

	for {
		chainLength := uint(len(node.blockchain))
		collectedTransactions := []SignedTransaction{}

		allTransactions := *NewTransactionQueue()
		allTransactions.Copy(&node.transactionQueue)

		/*
			Collect enough valid transactions to start mining the block
		*/
		skip := false
		for len(collectedTransactions) < CAPACITY {

			if node.broadcastType == ResolveRequestMessageType {
				time.Sleep(time.Second * 4)
				skip = true
				break
			}

			// Keep popping transactions until we have reached the CAPACITY amount
			transaction, err := allTransactions.Pop()

			if err != nil {
				allTransactions.Copy(&node.transactionQueue)
				continue
			}

			/*
				Check if the transaction has already been used to mine another block.
				If not, then use it in the current block
			*/
			alreadyUsed, exists := node.usedTransactions[transaction.TransactionID]
			if !exists || !alreadyUsed {
				node.blockchainLock.Lock()
				transactionOK := node.softValidateTransaction(transaction)
				node.blockchainLock.Unlock()
				if !transactionOK {
					continue
				}

				node.usedTransactions[transaction.TransactionID] = true
				collectedTransactions = append(collectedTransactions, transaction)
			}
		}

		/*
			We have collected enough transactions, start the clock and start mining.
			Mining stops only if we find the correct nonce or someone else finds it
			first and broadcasts it to the rest of the network.
		*/

		startTime := time.Now().Unix()

		if skip {
			skip = false
		} else {
			node.mineLock.Lock()
			block := node.mineBlock(collectedTransactions, chainLength)
			node.mineLock.Unlock()

			node.blockchainLock.Lock()
			if chainLength == uint(len(node.blockchain)) {
				node.broadcastBlock(block)
			}
			node.blockchainLock.Unlock()

			stopTime := time.Now().Unix()
			blockTime := stopTime - startTime
			fmt.Println(">>Block time:", blockTime)
			fmt.Println("")
		}

		node.updateUncommitted()
	}
}

func (node *Node) updateUncommitted() {

	node.UTXOsSoftValidated = make(map[[32]byte]TransactionOutput)
	for transactionID, UTXOs := range node.UTXOsCommitted {
		node.UTXOsSoftValidated[transactionID] = UTXOs
	}

	node.usedTransactions = make(map[[32]byte]bool)

	node.blockchainLock.Lock()
	for _, block := range node.blockchain {
		for _, transaction := range block.Transactions {
			node.usedTransactions[transaction.TransactionID] = true
		}
	}
	node.blockchainLock.Unlock()
}

func (node *Node) getId(publicKey rsa.PublicKey) string {
	for id, nodeData := range node.nodeDataMap {
		if equal(nodeData.PublicKey, publicKey) {
			return id
		}
	}
	return ""
}
