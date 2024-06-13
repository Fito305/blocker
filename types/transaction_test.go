package types

import (
	"fmt"
	"testing"

	"github.com/Fito305/blocker/crypto"
	"github.com/Fito305/blocker/proto"
	"github.com/Fito305/blocker/util"
	"github.com/stretchr/testify/assert"
)

func TestNewTransaction(t *testing.T) {
	fromPrivKey := crypto.GeneratePrivateKey()
	fromAddress := fromPrivKey.Public().Address().Bytes()

	toPrivKey := crypto.GeneratePrivateKey()
	toAddress := toPrivKey.Public().Address().Bytes()

	input := &proto.TxInput{
		PrevTxHas:    util.RandomHash(),
		PrevOutIndex: 0,
		PublicKey:    fromPrivKey.Public.Bytes(),
	}

	output1 := &proto.TxOutput{
		Amount:  5, // The amount we want to spend
		Address: toAddress,
	}
	output2 := &proto.TxOutput{
		Amount:  95,
		Address: fromAddress, // We are going to send it back.
	}

	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{input},
		Outputs: []*proto.TxOutput{output1, output2},
	}
	sig := SignTransaction(fromPrivKey, tx) // We send it, we need to sign it.
	input.Signature = sig.Bytes()

	assert.True(t, VerifyTransaction(tx))

	fmt.Printf("%+v\n", tx)

}

// NOTE: In Bitcoin let's say for example, I want to send 5 coins,
// by my balance is 100 coins. And I want to send 5 coins to the address AAA, what
// I need to do first is create an input and that input I have to specify the output of our previous transaction
// where we recieved coins. We need to specify the previous hash and previous index of the output, and the specify a publicKey and a signature.
// Input is basically telling us about our self (okay what was our previous output
// we recieved because that is what we need to spend. And what is our publicKey and our signature).
// The input is the info about oursleves. And the output is basically going to specify
// the destination of the coins we are going to spend.
// We have 100 coins, and we are going to send 5 coins to toAddress/ But then we need
// to spend the other resting 95 of the coins. We are going to send the 95 back to
// fromAddress. We recieved 100 coins from a previous transaction, we want to send 5 but we need to
// actually completely spend the whole 100 coins. So we send 5 to our destination,
// and we are going to send 95 back to ourselves. But we are going to send our whole balance. Or to someother adresses
// and then send the remainder back to yourself (or to another address of ours).

//If you have a ledger, you can see that they basically continuosly create new addresses
// and disribute that but the balance stays the same.
