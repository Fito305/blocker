// To create random blocks for testing purposes.
package util

import (
	"math/rand"
	"io"
	randc "crypto/rand"
	"time"

	"github.com/fito305/blocker/proto"
)

func RandonHash() []bytre {
	hash := make([]byte, 32)
	io.ReadFull(randc.Reader, hash)
	return hash
}

func RandomBlock() *proto.Block {
	header := &proto.header{
		Version: 1,
		Height: int32(rand.Intn(1000)),
		PrevHash: RandomHash(),
		RootHash: RandomHash(),
		Timestamp: time.Now().UnixNano(),
	}
	return &proto.Block{
		Header, header,
	}
}
