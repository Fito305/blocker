package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {
	privKey := GeneratePrivateKey()
	assert.Equal(t, len(privKey.Bytes()), PrivKeyLen)
	pubKey := privKey.Public()
	assert.Equal(t, len(pubKey.Bytes()), PubKeyLen)
}

func TestNewPrivateKeyFromString(t *testing.T) {
	var (
		seed = "10b2718850eb224208fff9f437aa2f07a634e121447ae09eb2d4e90f8bba24dc"
		privKey = NewPrivateKeyFromString(seed)
		addressStr = "d5ff8f76dadd9143be31cad6896184e66289c59b"
	)
	assert.Equal(t, PrivKeyLen, len(privKey.Bytes()))
	address := privKey.Public().Address()
	assert.Equal(t, addressStr, address.String())
}

func TestPrivateKeySign(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()
	msg := []byte("foo bar baz")

	sig := privKey.Sign(msg)
	assert.True(t, sig.Verify(pubKey, msg))

	// Test with invalid msg
	assert.False(t, sig.Verify(pubKey, []byte("foo")))

	// Test with invalid pubkey
	invalidPrivKey := GeneratePrivateKey()
	invalidPubKey := invalidPrivKey.Public()
	assert.False(t, sig.Verify(invalidPubKey, msg))
}

func TestPublicKeyToAddress(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()
	address := pubKey.Address()
	assert.Equal(t, AddressLen, len(address.Bytes()))
	fmt.Println(address) // So the address is basically 
	// Just the structure but because we have a string interface
	// / method, it'll print a nice string.
}
