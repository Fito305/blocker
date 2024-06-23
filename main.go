package main

import (
	"context"
	"log"
	"time"

	"github.com/Fito305/blocker/crypto"
	"github.com/Fito305/blocker/node"
	"github.com/Fito305/blocker/proto"
	"github.com/Fito305/blocker/util"
	"google.golang.org/grpc"
)

func main() {
	// seed := util.RandomHash() // Delete this after you grab the hash from the terminal.
	// fmt.Println(hex.EncodeToString(seed))
	// return // use to return early and copy the seed hash. Use that hash for the Genesis block.
	// SO what is going to happen is, we are going to spin up :3000, he doesnt know anything. Then :4000 is going to have this :3000 that is bootstrapNode(). So what is going to happen is :4000 is going to dail (handshake/conncect) with :3000. What is going to happen is that htye are going to exchange versions, so they both are going to have each other in their peer map. So :3000 is going to connect to :4000 and :4000 is going to be connected with :3000. Look at :5000 comment.
	makeNode(":3000", []string{}, true) // The bool is validator
	time.Sleep(time.Second) // We need to sleep here. It's very important we need to give it time. We are making a node and then we are making another node directly after it. So it's could be that :3000 is not ready yet.
	makeNode(":4000", []string{":3000"}, false)
	time.Sleep(4 * time.Second)
	makeNode(":5000", []string{":4000"}, false) // But the we spin up node :5000. And :5000 is only aware and is going to connect to :4000. So what is going to happen is :4000 is going to connect with :5000. And :5000 is going to connect with :4000. They are going to have each other in their peer map. But :5000 is not going to be connected with :3000. But thanks to the peer list, that is going to be sent by :4000 is connected all the way to :3000. :5000 has the chance the ability to also be aware of :3000. And that is how the whole network is going to be boostraped.

	for { // going to keep sending transactions every 2 seconds. if you have a block time of 5, you are going to have 2 transactions each block.
		time.Sleep(time.Second * 2)
		makeTransaction()
	}

}

func makeNode(listenAddr string, bootstrapNodes []string, isValidator bool) *node.Node {
	cfg := node.ServerConfig{
		Version: "Blocker-1",
		ListenAddr: listenAddr,
	}
	if isValidator {
		cfg.PrivateKey = crypto.GeneratePrivateKey()
	}
	n := node.NewNode(cfg)
	go n.Start(listenAddr, bootstrapNodes)
	return n
}

func makeTransaction() {
	privKey := crypto.GeneratePrivateKey()
	client, err := grpc.Dial(":3000", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	c := proto.NewNodeClient(client)

	// We are going to make a simple transaction that is not valid but we are going to put some values into it jsut for the sake of it.
	// Because we are not validating right now.
	tx := &proto.Transaction{
		Version: 1,
		Inputs: []*proto.TxInput{
			{
				PrevTxHash:   util.RandomHash(),
				PrevOutIndex: 0,
				PublicKey:    privKey.Public().Bytes(),
			},
		},
		Outputs: []*proto.TxOutput{
			{
				Amount:  99,
				Address: privKey.Public().Address().Bytes(),
			},
		},
	}

	_, err = c.HandleTransaction(context.TODO(), tx)
	if err != nil {
		log.Fatal(err)
	}
}
