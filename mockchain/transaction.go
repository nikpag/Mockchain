package main

import "crypto/rsa"

/*
Μην αλλάξετε το capitalization των λέξεων, πρέπει να ξεκινάνε με
κεφαλαίο για να γίνει σωστά το json encoding, αλλιώς δεν γίνονται
extracted τα variables
*/

type Transaction struct {
	SenderAddress      rsa.PublicKey
	ReceiverAddress    rsa.PublicKey
	Amount             uint
	TransactionID      [32]byte
	TransactionInputs  []TransactionInput
	TransactionOutputs [2]TransactionOutput
}

type SignedTransaction struct {
	SenderAddress      rsa.PublicKey
	ReceiverAddress    rsa.PublicKey
	Amount             uint
	TransactionID      [32]byte
	TransactionInputs  []TransactionInput
	TransactionOutputs [2]TransactionOutput
	Signature          []byte
}

type TransactionInput struct {
	PreviousOutputID [32]byte // reference to TransactionOutputs -> transactionId
}

type TransactionOutput struct {
	ID               [32]byte
	TransactionID    [32]byte      // the id of the transaction this output was created in
	RecipientAddress rsa.PublicKey // also known as the new owner of these coins.
	Amount           uint          // the amount of coins they own
}
