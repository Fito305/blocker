package types

import (
	"crypto/sha256"
	
	"github.com/Fito305/blocker/crypto"
	"github.com/Fito305/blocker/proto"
	// pb "google.golang.org/protobuf/runtime/protoimpl"
	pb "github.com/golang/protobuf/proto"
)

func VerifyBlock(b *proto.Block) bool { // Normally we can attach these functions on a receiver (form the block receiver), but in this case we cannot because it is on a generated proto buffer. But it actually the same concept.
	if len(b.PublicKey) != crypto.PubKeyLen {
		return false
	}
	if len(b.Signature) != crypto.SignatureLen {
		return false
	}
	sig := crypto.SignatureFromBytes(b.Signature)
	pubKey := crypto.PublicKeyFromBytes(b.PublicKey)
	hash := HashBlock(b)
	return sig.Verify(pubKey, hash)
}

func SignBlock(pk *crypto.PrivateKey, b *proto.Block) *crypto.Signature {
	hash := HashBlock(b)
	sig := pk.Sign(hash)
	b.PublicKey = pk.Public().Bytes()
	b.Signature = sig.Bytes()
	return sig
}

// HashBlock returns a SHA256 of the header.
func HashBlock(block *proto.Block) []byte {
	return HashHeader(block.Header)
}

func HashHeader(header *proto.Header) []byte {
	b, err := pb.Marshal(header) // Marshal the header because it is the only thing we want to sign.
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]

}


//NOTE: You could do a double SHA256 like in Bitcoin. 
// What they do is make a double SHA, double hash of the header.
// It's is not really needed. 
