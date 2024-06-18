package node

import (
	"testing"

	"github.com/Fito305/blocker/types"
	"github.com/Fito305/blocker/util"
	"github.com/stretchr/testify/assert"
)

func TestChainHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())
	for i := 0; i < 100; i++ {
		b := util.RandomBlock()
		assert.Nil(t, chain.AddBlock(b))
		assert.Equal(t, chain.Height(), i)
	}
}

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())

	for i := 0; i < 100; i++ {
		var (
			block     = util.RandomBlock()
			blockHash = types.HashBlock(block)
		)
		assert.Nil(t, chain.AddBlock(block))
		fetchedBlock, err := chain.GetBlockByHash(blockHash)
		assert.Nil(t, err)
		assert.Equal(t, block, fetchedBlock)

		fetchedBlockByHeight, err := chain.GetBlockByHeight(i)
		assert.Nil(t, err)
		assert.Equal(t, block, fetchedBlockByHeight)
	}
}

// So what are we doing? We create a random block, we store it into a chain, we fetch it back and then we compare if the the thing we stored the block
// is the same as we fetched. It is not a good implementation because we dont do validation. If you want to do validation you have to have
// the previous block and then make a random hash.
