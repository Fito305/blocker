package node

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/Fito305/blocker/crypto"
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
	chain :=  &Chain{
		blockStore: bs,
		headers: NewHeaderList(),
	}
	chain.addBlock(createGenesisBlock()) // Create the genesis block without validation. Now we have that genesis block each time we create our new blockchain.
	return chain
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) AddBlock(b *proto.Block) error {
	if err := c.ValidateBlock(b); err != nil {
		return err
	}
	return c.addBlock(b) // block with validation.
}

func (c *Chain) addBlock(b *proto.Block) error {
	// Add the header to the list of headers.
	c.headers.Add(b.Header)
	// validation
	return c.blockStore.Put(b) // block without validation. The genisis block is not validated.
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

func (c *Chain) ValidateBlock(b *proto.Block) error { // the b passed in the parameter is a new block.
	// Validate the signature of the block.
	if !types.VerifyBlock(b) {
		return fmt.Errorf("invalide block signature")
	}

	// validate if the previous hash is the actual hash of the current block.
	currentBlock, err := c.GetBlockByHeight(c.Height()) // The hash of the new block b, will be the has of the current block.
	if err != nil {
		return err
	}
	hash := types.HashBlock(currentBlock)
	if !bytes.Equal(hash, b.Header.PrevHash) {
		return fmt.Errorf("invalid previous block hash")
	}
	return nil
}

func createGenesisBlock() *proto.Block {
	privKey := crypto.GeneratePrivateKey()
	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
	}
	types.SignBlock(privKey, block)
	return block
}


// The Genesis Block right now will not be deterministic at all times because each time we are going to createGenesisBlock it is going to create a new privatekey
// and it should be one from a seed. A we are going to create the Genesis Block with NewChain(). The problem is that we cannot validate the genisis block like any other block
// because it is the genesis block. We don't have a previous hash due to the genesis block being the first block.
