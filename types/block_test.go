package types

import (
	"fmt"
	"testing"
	"encoding/hex"

	"github.com/Fito305/blocker/crypto"
	"github.com/Fito305/blocker/util"
	"github.com/stretchr/testify/assert"
)

func TestSignBlock(t *testing.T) {
	var (
	block = util.RandomBlock()
	privKey = crypto.GeneratePrivateKey()
	pubKey = privKey.Public()
	)
	
	// We are going to sign our block. First we get a private key because
	// we are going to sign with the private key. We are going to take the hash of the block,
	// which is going to in bytes the hash of the block is going to be a slice of bytes.
	// 32 bytes long we are going to sign the bytes with the private key which is going to return the 
	// signature, 64 bytes long. Then we are going to verify that signature by giving the public key 
	// and again the hash of that block. And then that should match.
	sig := SignBlock(privKey, block)
	assert.Equal(t, 64, len(sig.Bytes()))
	assert.True(t, sig.Verify(pubKey, HashBlock(block)))

}

func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()
	hash := HashBlock(block)
	fmt.Println(hex.EncodeToString(hash)) // Will print out the hash of the block. That is what we are going to use to retrieve the block. You can also retireve a block by it's height, but you can also query a block by it's hash on the chain. And that is basically this hash.
	assert.Equal(t, 32, len(hash))
}
