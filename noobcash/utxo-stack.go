package main

import (
	"errors"
	"sync"
)

type UTXOStack struct {
	stackLock sync.Mutex
	UTXOs     []TransactionOutput
}

func NewStack() *UTXOStack {
	return &UTXOStack{stackLock: sync.Mutex{}, UTXOs: make([]TransactionOutput, 0)}
}

func (s *UTXOStack) Copy(original *UTXOStack) {
	s.stackLock.Lock()
	original.stackLock.Lock()

	defer s.stackLock.Unlock()
	defer original.stackLock.Unlock()

	s.UTXOs = make([]TransactionOutput, len(original.UTXOs))
	copy(s.UTXOs, original.UTXOs)
}

func (s *UTXOStack) Size() int {
	return len(s.UTXOs)
}

func (s *UTXOStack) Push(utxo TransactionOutput) {
	s.stackLock.Lock()
	defer s.stackLock.Unlock()

	s.UTXOs = append(s.UTXOs, utxo)
}

func (s *UTXOStack) Pop() (TransactionOutput, error) {
	s.stackLock.Lock()
	defer s.stackLock.Unlock()

	l := len(s.UTXOs)
	if l == 0 {
		return TransactionOutput{}, errors.New("empty UTXOStack")
	}

	res := s.UTXOs[0]
	s.UTXOs = s.UTXOs[1:]
	return res, nil
}
