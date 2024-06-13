package	node 

import (
	"context"
	"fmt"

	"github.com/Fito305/blocker/proto"
	"google.golang.org/grpc/peer"
)

type Node struct {
	version string
	proto.UnimplementedNodeServer
}

func NewNode() *Node {
	return &Node{
		version: "blocker-0.1",
	}
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	ourVersion := &proto.Version{
		Version: n.version,
		Height: 100,
	}

	p, _ := peer.FromContext(ctx)
	
	fmt.Printf("receieved version from %s: %+v\n",v, p.Addr)

	return ourVersion, nil // In this handshake it's going to accept the connection yes or no. Because we are going to maintain a map of grpc stuff. 
}

func (n *Node) HandleTransaction (ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	fmt.Println("recieved tx from:", peer)
	return &proto.Ack{}, nil
}

// NOTE ctx because we want to get our peer later on from this context.
// A lot of people don't know that but you can extract information
// about the peer that is calling this thing from the context.

// peers map[] so basically when we handle a transaction, we can boradcast this to all the peers. We have all these
// grpc connections. and how we are going to put all these peers into a map, it's because in each time we HandleTransaction we are going
// to check if the thing calling our HandleTransaction is actually in the peer map and if it is it gives the connection because we want only 
// to give connection to know peers to the server. We will also ad some json rpc server so everybody we can just publish the transaction into
// the network. 

// In a grpc enviroment you want to respond. So each time it's an rpc you send something and you expect to recieve a response
// but in this case it is not going to happen yet. Maybe instead of None we can say an aquired 'Ack'. "Like hey yo, it's fine."
// We got your transaction we handled it, it's good instead of None.
