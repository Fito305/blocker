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
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	assert.Equal(t, 0, chain.Height())
	/*block*/ _, err := chain.GetBlockByHeight(0) // block is the genesis block. We don't care about the block, the only thing we want is that the block exists (has been created in the chain).

	assert.Nil(t, err)
}

func TestChainHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
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
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())

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

func TestAddBlockWithInsufficientFunds(t *testing.T) {
	var (
		chain = NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
		block = randomBlock(t, chain)
		privKey = crypto.NewPrivateKeyFromSeedStr(godSeed)
		recipient = crypto.GeneratePrivateKey().Public().Address().Bytes()
	)
	prevTx, err := chain.txStore.Get("hashPrintedFromFilechain.go-addBlockfuncfmtPrintlnTX:") // - fetch transaction transaction. Going to fetch a transaction that we stored because in that transaction there are my outputs. The outputs that I need to use for inputs below. This is nasty code.
	assert.Nil(t, err)
	inputs := []*proto.TxInput{
		{
			PrevTxHash: types.HashTransaction(prevTx),
			PrevOutIndex: 0, // This is output / index 0.
			PublicKey: privKey.Public().Bytes(),
		},
	}
	outputs := []*proto.TxOutput{
		{
			Amount: 10001, // Cannot send 10001 because we only have 1000 in chain.go
			Address: recipient,
		},
	}
	tx := &proto.Transaction{
		Version: 1,
		Inputs: inputs,
		Outputs: outputs,
	}
	sig := types.SignTransaction(privKey, tx)
	tx.Inputs[0].Signature = sig.Bytes()
	block.Transactions = append(block.Transactions, tx)
	require.NotNil(t, chain.AddBlock(block)) // Adding a block should fail due to not having enough funds.
}

func TestAddblockWithTx(t *testing.T) {
	var (
		chain = NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
		block = randomBlock(t, chain)
		privKey = crypto.NewPrivateKeyFromSeedStr(godSeed)
		recipient = crypto.GeneratePrivateKey().Public().Address().Bytes()
	)
	// The input of a transaction is the output of a previous transaction. 
	// In the input we need to specify the previous output.

	prevTx, err := chain.txStore.Get("hashPrintedFromFilechain.go-addBlockfuncfmtPrintlnTX:") // - fetch transaction transaction. Going to fetch a transaction that we stored because in that transaction there are my outputs. The outputs that I need to use for inputs below. This is nasty code.
	assert.Nil(t, err)


	inputs := []*proto.TxInput{
		{
			PrevTxHash: types.HashTransaction(prevTx),
			PrevOutIndex: 0, // This is output / index 0.
			PublicKey: privKey.Public().Bytes(),
		},
	}
	outputs := []*proto.TxOutput{
		{
			Amount: 100, // This output is spent output. 
			Address: recipient,
		},
		{
			Amount: 900, // We send it back to our own address. We can use this back again as an input.
			Address: privKey.Public().Address().Bytes(), // This is nasty code. And it is hard coded
		},
	}
	tx := &proto.Transaction{
		Version: 1,
		Inputs: inputs,
		Outputs: outputs,
	}

	sig := types.SignTransaction(privKey, tx)
	tx.Inputs[0].Signature = sig.Bytes()

	block.Transactions = append(block.Transactions, tx)
	require.Nil(t, chain.AddBlock(block))
}

// So what are we doing? We create a random block, we store it into a chain, we fetch it back and then we compare if the the thing we stored the block
// is the same as we fetched. It is not a good implementation because we dont do validation. If you want to do validation you have to have
// the previous block and then make a random hash.

// If you make a transaction, if I want to send 50 bitcons to somebody, I need to query the blockchain (database), and specify unspent outputs and and use them as inputs
// in my new transaction. In this case it is a test and we cannot specify any outputs in our inputs because we don't have outputs. 

// In our Genesis Block we are sending money from nowhere (created out of thin air) to the godSeed. Then we are going to send from the our godSeed address to the recipient varaible in the function above. 
// We are sending 100 in the outputs := []*proto.TxOutput but we have 1000 so we need to specify in our inputs the output from the Genesis with is 0 (PrevOutIndex: 0). 
// In the outputs, we make another transaction back to ourselves for the 900 left out of the 1000 (we sent 100)

// If I want to make a new transaction, that means that I need to provide two things in the transaction. 
// I need to provide an input or multiple inputs which is going to be outputs of previous transactions sent to me. 
// An output of a transaction is always going to be an input of the next transaction for myself. 
// So if I want to send 100 tokens to someone, I need to provide (that is going to be an output, im going to make an output of 100 tokens to your address).
// But I need to have an input or multiple inputs with the sum of atleast 100 tokens.
// In order to test this, we create a Genesis block. The GenesisBlock is the first block of a block chain. Im going to create some coins out of thin air.
// And Im going to send them the 1000 coins to the private key in chain.go createGenesisBlock().
// And the private key is going to be the godSeed variable. So it is going to be the address of the godSeed.
// And that is going to be prevTx varaible in TestAddBlockWithTx() above.
// So what we do is check the hash of the transaction. We got the hash of the transaction in prevTx and we are going to fetch that transaction 
// so we can use the output as an input for our test. This is important to understand and it is what makes it secure.


// BITCOIN USES UTXO.
