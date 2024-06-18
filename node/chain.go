package node

import (
	"encoding/hex"
	"fmt"

	"github.com/Fito305/blocker/proto"
	"github.com/Fito305/blocker/types"
)

type HeaderList struct {
	headers []*proto.Header
}

func NewHeaderList() *HeaderList {
	return &HeaderList{
		headers: []*proto.Header{},
	}
}

func (list *HeaderList) Add(h *proto.Header) {
	list.headers = append(list.headers, h)
}

func (list *HeaderList) Get(index int) *proto.Header {
	if index > list.Height() {
		panic("index too high!")
	}
	return list.headers[index]
}

func (list *HeaderList) Height() int {
	return list.Len() - 1
}

// [A, B, C, D, E] Len = 5 height = 4
func (list *HeaderList) Len() int {
	return len(list.headers)
}

type Chain struct {
	blockStore BlockStorer
	headers    *HeaderList
}

// Constructor
func NewChain(bs BlockStorer) *Chain {
	return &Chain{
		blockStore: bs,
		headers: NewHeaderList(),
	}
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) AddBlock(b *proto.Block) error {
	// Add the header to the list of headers.
	c.headers.Add(b.Header)
	// validation
	return c.blockStore.Put(b)
}

func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hashHex := hex.EncodeToString(hash)
	return c.blockStore.Get(hashHex)
}

func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	// We are going to check if we want to get a block by the height, it is going to check if we have it.
	if c.Height() < height {
		return nil, fmt.Errorf("given height (%d) too high - height (%d)", height, c.Height())
	}
	header := c.headers.Get(height)
	hash := types.HashHeader(header)
	return c.GetBlockByHash(hash)
}

// NOTE: We need to store a lot of stuff. We need to create our memory store. We need to store blocks, transactions, a lot of stuff.
// And most of these blockchains they do that by using some kind of level DB over ROX Db which is a embedded key value store.
// And an embedded key value store is something that is not running on your machine. It is embedded into your application.
// It is running into your applciation. So ppl dont need to have if they want to put up a node, they dont need to have something else
// installed. The only thing they need to do is run your program, run the node and it is all good. And everything is getting boot up in the same binary,
// in the same compiled executable program. But before we are going to use these key value stores, these embedded stuff,
// we are going to make our own memory implementation. So we are going to store everything in memory. Everything is eventually a key value store
// embedded (not always they do store it into disk). We are going to store everything in memory just so we have something that we
// can simply use to test. And because it is an interface if we later on want to swap that out with a level DB or a blocks DB,
// we just need to make the interface implementation. Swap it out and we dont need to change anything from a business logic. because
// the interface will do its job. That is the power of interfaces.

// There should always be one block because we are always going to have the genesis block.

// So what is going to happen is that each time we add a block, we are going to add the header to the list of headers. 
