package node

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/Fito305/blocker/proto"
	"github.com/Fito305/blocker/types"
)

type BlockStorer interface {
	Put(*proto.Block) error
	Get(string) (*proto.Block, error)
}

type MemoryBlockStore struct {
	lock   sync.RWMutex // To make concurent safe
	blocks map[string]*proto.Block
}

func NewMemoryBlockStore() *MemoryBlockStore {
	return &MemoryBlockStore{
		blocks: make(map[string]*proto.Block),
	}
}

func (s *MemoryBlockStore) Put(b *proto.Block) error {
	s.lock.Lock() // makes is concurrency safe.
	defer s.lock.Unlock()
	hash := hex.EncodeToString(types.HashBlock(b))
	s.blocks[hash] = b
	return nil
}

func (s *MemoryBlockStore) Get(hash string) (*proto.Block, error) {
	s.lock.RLock() // Read lock
	defer s.lock.RUnlock()
	block, ok := s.blocks[hash]
	if !ok {
		return nil, fmt.Errorf("block with hash [%s] does not exists", hash)
	}
	return block, nil
}

// NOTE: If you want to do this with a good implementation that is not in memory you would serialized that into bytes.
