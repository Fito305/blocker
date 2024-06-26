package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

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
			fmt.Println("INVALID root hash")
			return false
		}
	}

	if len(b.PublicKey) != crypto.PubKeyLen {
		fmt.Println("Invalid public key length")
		return false
	}
	if len(b.Signature) != crypto.SignatureLen {
		fmt.Println("Invalid signature length")
		return false
	}
	var (
		sig    = crypto.SignatureFromBytes(b.Signature)
		pubKey = crypto.PublicKeyFromBytes(b.PublicKey)
		hash   = HashBlock(b)
	)
	fmt.Println(hex.EncodeToString(hash))
	if !sig.Verify(pubKey, hash) {
		fmt.Printf("%v\n", b.Header) // delete all the logs
		fmt.Println(hex.EncodeToString(sig.Bytes()))
		fmt.Println("Invalid block signature")
		return false
	}
	return true
}

func SignBlock(pk *crypto.PrivateKey, b *proto.Block) *crypto.Signature {
	if len(b.Transactions) > 0 {
		tree, err := GetMerkleTree(b)
		if err != nil {
			panic(err)
		}

		b.Header.RootHash = tree.MerkleRoot()
	}
	hash := HashBlock(b)
	fmt.Println("Hash before signature", hex.EncodeToString(hash))
	sig := pk.Sign(hash)
	b.PublicKey = pk.Public().Bytes()
	b.Signature = sig.Bytes()

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


// If you have a wrong signature it means that something is wrong.
//BUG* FIXED - That was in SignBlock - the only thing that gets hashed is the message Header {} in types.proto. The rootHash in Header provides us with all
// the validation fo the transactions. The rootHash is the merkleTree for all the transactions. 
// What we were doing is we hashed the block `hash := HashBlock(b)`, then we signed the block `sig := pk.Sign(Hash)
// then we set the public key which doesn't really matter because we only have the header `b.PublicKey = pk.Public().Bytes()
// The problem is the rootHash and why? Because we sign the block in the sig but we set the rootHash below the signature (code lines below), in 
// b.Header.RootHash = tree.MerkleRoot(). So we did not provide a signature for the rootHash. We signed everything except the rootHash. So to fix it,
// we placed the if statement with b.Header.Roothash = tree.merkleRoot() above hash := HashBlock(b). sig, b.PublicKey and b.Singature.

//Why not a proof of work / stake but instead a proof of consensus? 
