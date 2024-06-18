package types

import (
	"crypto/sha256"
	
	"github.com/Fito305/blocker/crypto"
	"github.com/Fito305/blocker/proto"
	// pb "google.golang.org/protobuf/runtime/protoimpl"
	pb "github.com/golang/protobuf/proto"
)

func SignBlock(pk *crypto.PrivateKey, b *proto.Block) *crypto.Signature {
	return pk.Sign(HashBlock(b))
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
