package types

import (
	"crypto/sha256"
	
	"github.com/Fito305/blocker/crypto"
	"github.com/Fito305/blocker/proto"
	pb "google.golang.org/protobuf/runtime/protoimpl"
)

func SignBlock(pk *crypto.PrivateKey, b *proto.Block) *crypto.Signature {
	return pk.Sign(HashBlock(b))
}

// HashBlock returns a SHA256 of the header.
func HashBlock(block *proto.Block) []byte {
	b, err := pb.Marshal(block)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}



//NOTE: You could do a double SHA256 like in Bitcoin. 
// What they do is make a double SHA, double hash of the header.
// It's is not really needed. 
