package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"io"

)

// The length of the privat key is going to be 64 because its going to 
// have the first 32 bytes is going to be the private key, and then it's
// going to append the public key on the last 32 bytes. 32 + 32 = 64.
const (
	privKeyLen = 64
	pubKeyLen = 32
	seedLen = 32
	addressLen = 20
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

func NewPrivateKeyFromString(s string) *PrivateKey {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return NewPrivateKeyFromSeed(b)
}

func NewPrivateKeyFromSeed(seed []byte) *PrivateKey {
	if len(seed) != seedLen {
		panic("invalid seed length, must be 32")
	}
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
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
	return &Signature{
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

func (p *PublicKey) Address() Address {
	return Address{
		value: p.key[len(p.key)-addressLen:],
	}
}

func (p *PublicKey) Bytes() []byte {
	return p.key
}

type Signature struct {
	value []byte
}

func (s *Signature) Bytes() []byte {
	return s.value
}

func (s *Signature) Verify(pubKey *PublicKey, msg []byte) bool {
	return ed25519.Verify(pubKey.key, msg, s.value)
}

type Address struct {
	value []byte 
}

// It's always nice to have Bytes() incase you need to do 
// serilization. 
func (a Address) Bytes() []byte {
	return a.value
}

func (a Address) String() string {
	return hex.EncodeToString(a.value)
}

//NOTE:
// Public keys can be shared publicly.

// An address is basically another representation 
// of the public key. Most of the time it's some king of HEX
// representation of the bytes of the public key. Most of the
// time it's just the first 20 or something bytes.
// We will use the first 20 bytes as our address.
