package types

import (
	"bytes"
	"crypto/sha256"

	"github.com/Fito305/blocker/crypto"
	"github.com/Fito305/blocker/proto"
	"github.com/cbergoon/merkletree"
	// pb "google.golang.org/protobuf/runtime/protoimpl"
	pb "github.com/golang/protobuf/proto"
)

type TxHash struct {
	hash []byte
}

func NewTxHash(hash []byte) TxHash {
	return TxHash{hash: hash}
}

func (h TxHash) CalculateHash() ([]byte, error) {
	return h.hash, nil
}

func (h TxHash) Equals(other merkletree.Content) (bool, error) {
	equals := bytes.Equal(h.hash, other.(TxHash).hash)
	return equals, nil
}

func VerifyBlock(b *proto.Block) bool { // Normally we can attach these functions on a receiver (form the block receiver), but in this case we cannot because it is on a generated proto buffer. But it actually the same concept.
	if len(b.Transactions) > 0 {
		if !VerifyRootHash(b) {
			return false
		}
	}

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

	if len(b.Transactions) > 0 {
		tree, err := GetMerkleTree(b)
		if err != nil {
			panic(err)
		}

		b.Header.RootHash = tree.MerkleRoot()
	}
	return sig
}

func VerifyRootHash(b *proto.Block) bool {
	// We are going to verify the block, if it is verified, and if we don't have the root hash we are going to set it.
	tree, err := GetMerkleTree(b)
	if err != nil {
		return false
	}
	valid, err := tree.VerifyTree()
	if err != nil {
		return false
	}

	if !valid {
		return false
	}

	return bytes.Equal(b.Header.RootHash, tree.MerkleRoot())
}

func GetMerkleTree(b *proto.Block) (*merkletree.MerkleTree, error) {
	list := make([]merkletree.Content, len(b.Transactions))
	for i := 0; i < len(b.Transactions); i++ {
		list[i] = NewTxHash(HashTransaction(b.Transactions[i]))
	}

	t, err := merkletree.NewTree(list)
	if err != nil {
		return nil, err
	}

	return t, nil
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
