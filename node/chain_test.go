package node

import (
	"encoding/hex"
	"fmt"
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

func TestAddblockWithTx(t *testing.T) {
	var (
		chain = NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
		block = randomBlock(t, chain)
		privKey = crypto.NewPrivateKeyFromSeedStr(godSeed)
		recipient = crypto.GeneratePrivateKey().Public().Address().Bytes()
	)
	// The input of a transaction is the output of a previous transaction. 
	// In the input we need to specify the previous output.

	ftt, err := chain.txStore.Get("hashPrintedFromFilechain.go-addBlockfuncfmtPrintlnTX:") // ftt - fetch transaction transaction. Going to fetch a transaction that we stored because in that transaction there are my outputs. The outputs that I need to use for inputs below. This is nasty code.
	assert.Nil(t, err)


	inputs := []*proto.TxInput{
		{
			PrevTxHash: types.HashTransaction(ftt),
			PrevOutIndex: 0,
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
	txHash := hex.EncodeToString(types.HashTransaction(tx))

	fetchedTx, err := chain.txStore.Get(txHash)
	assert.Nil(t, err)
	assert.Equal(t, tx, fetchedTx)

	// check if their is an UTXO that is unspent.
	address := crypto.AddressFromBytes(tx.Outputs[1].Address) // Outputs[0] gets the hash for the amount spent 100 above, and Outputs[1] gets the hash for the amount 900 above sent back to us.
	key := fmt.Sprintf("%s_%s", address, txHash)

	utxo, err := chain.utxoStore.Get(key)
	assert.Nil(t, err)
	fmt.Println(utxo)
}

// So what are we doing? We create a random block, we store it into a chain, we fetch it back and then we compare if the the thing we stored the block
// is the same as we fetched. It is not a good implementation because we dont do validation. If you want to do validation you have to have
// the previous block and then make a random hash.

// If you make a transaction, if I want to send 50 bitcons to somebody, I need to query the blockchain (database), and specify unspent outputs and and use them as inputs
// in my new transaction. In this case it is a test and we cannot specify any outputs in our inputs because we don't have outputs. 

// In our Genesis Block we are sending money from nowhere (created out of thin air) to the godSeed. Then we are going to send from the our godSeed address to the recipient varaible in the function above. 
// We are sending 100 in the outputs := []*proto.TxOutput but we have 1000 so we need to specify in our inputs the output from the Genesis with is 0 (PrevOutIndex: 0). 
// In the outputs, we make another transaction back to ourselves for the 900 left out of the 1000 (we sent 100)
