package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"io"

)

// The length of the privat key is going to be 64 because its going to 
// have the first 32 bytes is going to be the private key, and then it's
// going to append the public key on the last 32 bytes. 32 + 32 = 64.
const (
	privKeyLen = 64
	pubKeyLen = 32
	seedLen = 32
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

func GeneratePrivateKey() *PrivateKey {
	seed := make([]byte, seedLen)
	_, err := io.ReadFull(rand.Reader, seed)
	if err != nil {
		panic(err) // If your program cannot continue, there is no need to handle the error. Just panic().
	}
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

// Pirvate keys are the keys we are going to hold on our system.
// We can never publish them and never make them public. But you need
// them to sign stuff. Such as ypour transactions.
func (p *PrivateKey) Bytes() []byte {
	return p.key
}

func (p *PrivateKey) Sign(msg []byte) *Signature {
	return Signature{
		value: ed25519.Sign(p.key, msg),
	}
}

func (p *PrivateKey) Public() *PublicKey {
	b := make([]byte, pubKeyLen) // b is a buffer
	copy(b, p.key[32:]) 

	return &PublicKey{
		key: b,
	}
}

type PublicKey struct {
	key ed25519.PublicKey
}

func (p *PublicKey) Bytes() []byte {
	return p.key
}

type Signature struct {
	value []byte
}
