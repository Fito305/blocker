package node

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/Fito305/blocker/crypto"
	"github.com/Fito305/blocker/proto"
	"github.com/Fito305/blocker/types"
)

const godSeed = "printedHashOnTheTerminalViafmt.PrintlnFromTheSeedVaraibleInTheFileMain.goInMainFunc" // It is to make deterministic privateKey so we can have coins or some kind of a genesis input. The output that we can use as input in our transactions.

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

type UTXO struct {
	Hash     string
	OutIndex int
	Amount   uint64
	Spent    bool
}

type Chain struct {
	txStore    TXStorer
	blockStore BlockStorer
	utxoStore  UTXOStorer
	headers    *HeaderList
}

// Constructor
func NewChain(bs BlockStorer, txStore TXStorer) *Chain {
	chain := &Chain{
		blockStore: bs,
		txStore:    txStore,
		utxoStore:  NewMemoryUTXOStore(), // hard code in because we will refactor this later.
		headers:    NewHeaderList(),
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

	for _, tx := range b.Transactions {
		// fmt.Println("NEW X: ", hex.EncodeToString(types.HashTransaction(tx)))
		if err := c.txStore.Put(tx); err != nil {
			return err
		}
		hash := hex.EncodeToString(types.HashTransaction(tx))

		// address_txhash - concat address with txhash
		// that way we can fetch all the uspent address.
		for it, output := range tx.Outputs { // We have to loop over this because we have to make it for each output.
			utxo := &UTXO{
				Hash:        hash,
				Amount:      output.Amount,
				OutIndex: it,
				Spent:       false, // go will make this false by default but this is to make it more verbose.
			}
			address := crypto.AddressFromBytes(output.Address)
			key := fmt.Sprintf("%s_%s", address, hash)
			if err := c.utxoStore.Put(key, utxo); err != nil {
				return err
			}
		}
	}
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

	for _, tx := range b.Transactions {
		if !types.VerifyTransaction(tx) {
			return fmt.Errorf("invalid tx signature")
		}

		// for _, input := range tx.Inputs { // we need to check if this input is going to be an unspent output.
		// 	// we currently have access to the hash but we don't have acces to the address. So if we don't have access to the address we cannot fetch it.
		// }
	}
	return nil
}

func createGenesisBlock() *proto.Block {
	// godSeed is global. We have a godSeed we can always use.
	privKey := crypto.NewPrivateKeyFromSeedStr(godSeed) // Now we will always have a deterministic privateKey.

	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
	}

	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{}, // We don't need the input only the outputs.Because we are not going to validate this. Because it is the genesis (first) block, we don't give a shit. We are going to create tokens out of thin air.
		Outputs: []*proto.TxOutput{ // We are gong to output the seed above so we can use it in our test.
			{
				Amount:  1000,
				Address: privKey.Public().Address().Bytes(),
			},
		},
	}

	block.Transactions = append(block.Transactions, tx)
	types.SignBlock(privKey, block)
	// We also need a Merkle tree, a merkle root of all the transactions inside of it and put it in a block so we can sign that so we have a deterministic way. The merkle root of the tx we need to hash. We don't hash transactions we need to hash the root. Otherwise we can tamper with these transaction which is a security hazard.
	return block
}

// The Genesis Block right now will not be deterministic at all times because each time we are going to createGenesisBlock it is going to create a new privatekey
// and it should be one from a seed. A we are going to create the Genesis Block with NewChain(). The problem is that we cannot validate the genisis block like any other block
// because it is the genesis block. We don't have a previous hash due to the genesis block being the first block.

// In order to check if the amount has been spent or not spent, we are going to keep track of the unspent transaction outputs.

// utxoStore -  is to store utxo. We will store it when we are creating it as an output. If you are using it as an input then we know it is going to be spent. But when it is an
// output it is going to be unspent unless it is going to be used as an input.

// it - is iterator
