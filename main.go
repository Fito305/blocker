package main

import (
	"context"
	"log"
	"time"

	"github.com/Fito305/blocker/node"
	"github.com/Fito305/blocker/proto"
	"google.golang.org/grpc"
)

func main() {
	// SO what is going to happen is, we are going to spin up :3000, he doesnt know anything. Then :4000 is going to have this :3000 that is bootstrapNode(). So what is going to happen is :4000 is going to dail (handshake/conncect) with :3000. What is going to happen is that htye are going to exchange versions, so they both are going to have each other in their peer map. So :3000 is going to connect to :4000 and :4000 is going to be connected with :3000. Look at :5000 comment.
	makeNode(":3000", []string{})
	time.Sleep(time.Second) // We need to sleep here. It's very important we need to give it time. We are making a node and then we are making another node directly after it. So it's could be that :3000 is not ready yet.
	makeNode(":4000", []string{":3000"})
	time.Sleep(4 * time.Second)
	makeNode(":5000", []string{":4000"}) // But the we spin up node :5000. And :5000 is only aware and is going to connect to :4000. So what is going to happen is :4000 is going to connect with :5000. And :5000 is going to connect with :4000. They are going to have each other in their peer map. But :5000 is not going to be connected with :3000. But thanks to the peer list, that is going to be sent by :4000 is connected all the way to :3000. :5000 has the chance the ability to also be aware of :3000. And that is how the whole network is going to be boostraped.

	select {}
}

func makeNode(listenAddr string, bootstrapNodes []string) *node.Node {
	n := node.NewNode()
	go n.Start(listenAddr, bootstrapNodes)
	return n
}

func makeTransaction() {
	client, err := grpc.Dial(":3000", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	c := proto.NewNodeClient(client)

	version := &proto.Version{
		Version: "blocker-0.1",
		Height: 1,
		ListenAddr: ":4000", // The address of the calling node.
	}

	_, err = c.Handshake(context.TODO(), version)
	if err != nil {
		log.Fatal(err)
	}
}


// NOTE: A very important aspect in blockchain is called Peer Discovery.
// For example the whole Bitcoin network has a tremendous amount of nodes. And we cannot put all these nodes into 
// our peer list (the bootstrapNodes). Each time we make a node we are going to say, "Yo these nodes I know them." And again
// bootstrapNodes, these are very important nodes, it could be nodes you created yourself. Like Bitcoin has a couple of very 
// interesting bootstrapNodes. And the main goal is that you connect the list of bootstrapNodes (a list of predefined nodes), 
// A list of predefined astrings, addresses you can connect (could be two, three etc ...). And you are going to connect to these nodes,
// but instead of four nodes, maybe you have a thousand nodes in the network. So how do you are you going to connect to all of them?
// Or at least connect to a big portion of them? Well thats with Peer Discovery. So you are going to connect to these nodes,
// and we are sending a version, and they are going to respond with their version. And a good idea is to respond with all the connected peers
// that we have at that peer list. So you are going to connect to one peer and they are going to respond to it with a version. And that is going to be for example,
// they are going to send a list of 100 peers so we can connect to these peers we are going to send the version back. Again 
// these peers are going to send back their peers they are connected to another list of 100. And that is how we are going to connect to
// eventually a lot of peers in the network. maybe all of the peers. We dont know. Peer Discovery is how you make your network healthy
// by connecting with a broad range of nodes.
