package node

import (
	"testing"

	"github.com/Fito305/blocker/crypto"
	"github.com/Fito305/blocker/proto"
	"github.com/Fito305/blocker/types"
	"github.com/Fito305/blocker/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func randomBlock(t *testing.T, chain *Chain) *proto.Block {
	privKey := crypto.GeneratePrivateKey()
	b := util.RandomBlock()
	prevBlock, err := chain.GetBlockByHeight(chain.Height())
	require.Nil(t, err)
	b.Header.PrevHash = types.HashBlock(prevBlock)
	types.SignBlock(privKey, b)
	return b
}

// Check if the Genesis Block was created.
func TestNewChain(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())
	assert.Equal(t, 0, chain.Height())
	/*block*/ _, err := chain.GetBlockByHeight(0) // block is the genesis block. We don't care about the block, the only thing we want is that the block exists (has been created in the chain).

	assert.Nil(t, err)
}

func TestChainHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())
	for i := 0; i < 100; i++ {
		b := randomBlock(t, chain)
		// b := util.RandomBlock() // These commented lines are replaced by the helper function randomBlock
		// prevBlock, err := chain.GetBlockByHeight(chain.Height())
		// require.Nil(t, err)
		// b.Header.PrevHash = types.HashBlock(prevBlock)

		require.Nil(t, chain.AddBlock(b))
		require.Equal(t, chain.Height(), i+1) // We added the genisis block so we have to do i+1.
	}
}

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())

	for i := 0; i < 100; i++ {

		block := randomBlock(t, chain)

		// prevBlock, err := chain.GetBlockByHeight(chain.Height())
		// require.Nil(t, err)
		// block.Header.PrevHash = types.HashBlock(prevBlock)

		blockHash := types.HashBlock(block)
		require.Nil(t, chain.AddBlock(block))

		fetchedBlock, err := chain.GetBlockByHash(blockHash)
		require.Nil(t, err)
		require.Equal(t, block, fetchedBlock)

		fetchedBlockByHeight, err := chain.GetBlockByHeight(i + 1)
		require.Nil(t, err)
		require.Equal(t, block, fetchedBlockByHeight)
	}
}

// So what are we doing? We create a random block, we store it into a chain, we fetch it back and then we compare if the the thing we stored the block
// is the same as we fetched. It is not a good implementation because we dont do validation. If you want to do validation you have to have
// the previous block and then make a random hash.
