package types

import (
	"crypto/sha256"


	"github.com/Fito305/blocker/crypto"
	"github.com/Fito305/blocker/proto"
	// pb "google.golang.org/protobuf/runtime/protoimpl" // In the video this path is different.
	pb "github.com/golang/protobuf/proto"
)

func SignTransaction(pk *crypto.PrivateKey, tx *proto.Transaction) *crypto.Signature {
	return pk.Sign(HashTransaction(tx))
}

func HashTransaction(tx *proto.Transaction) []byte {
	b, err := pb.Marshal(tx)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:] // Specify it as a slice.
}

func VerifyTransaction(tx *proto.Transaction) bool {
	for _, input := range tx.Inputs {
		if len(input.Signature) == 0 {
			panic("the transaction has no signature")
		}

		var (
			sig = crypto.SignatureFromBytes(input.Signature)
			pubKey = crypto.PublicKeyFromBytes(input.PublicKey)
		)
		// TODO: make sure we dont run into problems after verification
		// cause we have set the signature to nil.
		tempSig := input.Signature
		input.Signature = nil // We don't hash the signature, we hash the transaction without the signature.
		// That is why we set the signature to nil. And then we are going to hash the transaction and then verify.
		if !sig.Verify(pubKey, HashTransaction(tx)) {
			return false
		}
		input.Signature = tempSig
	}
	return true
}


// NOTE: the problem with hashTransaction is that we are going to hash the
// signature itself, which we cannot do that in VerifyTransaction. We do HashTransaction()
// which is basically going to hash the whole transaction including the signature which
// we don't currently have at the moment your going ot sign it. So what we are going to do
// is some dirty stuff. What we can do is inputSignature = nil



// NOTE: The output is only valid if the corresponding previous inputs
// was verified.

// NOTE:: so if we really want to make transactions later on, we need to have access to the previous
// transaction. Because i need to find my output. I need to specify the output I want to spend.
// That is how we are going ot track. because there is actually no database of balances
// kind of like in Ethereum. Everything in Ethereum is getting stored in an account
// model so we can just book update the address, transition the state and update the account balance
// and then we know it. But in Bitcoin there is no such thing as that. In Bitcoin we basically
// to hand over our whole balance like my bank account it 100 coins so I'm going to
// spend my complete 100 coins, send 5 coins, but send 95 coins back to myself. And that way
// we can completely track the total balance and also the total market cap the coin has (coins in circulation). 
