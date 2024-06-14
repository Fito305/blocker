package	node 

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/Fito305/blocker/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type Node struct {
	version string
	listenAddr string
	logger *zap.SugaredLogger

	peerLock sync.RWMutex
	peers map[proto.NodeClient]*proto.Version
	
	proto.UnimplementedNodeServer
}

func NewNode() *Node {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	logger, _ := loggerConfig.Build()
	return &Node{
		peers: make(map[proto.NodeClient]*proto.Version),
		version: "blocker-0.1",
		logger: logger.Sugar(),
	}
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	// Print out here who are we as well [%s]. Our listen Address 
	n.logger.Debugw("new peer connected", "addr:", v.ListenAddr, "height", v.Height)

	n.peers[c] = v
}

func (n *Node) deletePeer(c proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	delete(n.peers, c)
}

func (n *Node) BootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		c, err := makeNodeClient(addr)
		if err != nil {
			return err
		}
		v, err := c.Handshake(context.Background(), n.getVersion())
		if err != nil {
			n.logger.Error("handshake error:", err)
			continue
		}
		// There are a couple of things we can do. Do not return the hanshake but each time we recieve a handshake we could call handshake back.
		// But in this case, do this: because v is the version you get back from the guy. So we can add the connect 'c' with the version 'v'. 
		n.addPeer(c, v)
	}
	return nil
}

func (n *Node) Start(listenAddr string) error {
	n.listenAddr = listenAddr

	var (
	opts = []grpc.ServerOption{}
	grpcServer = grpc.NewServer(opts...)
	)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	proto.RegisterNodeServer(grpcServer, n)
	n.logger.Infow("node started...", "port", n.listenAddr)

	return grpcServer.Serve(ln)
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	c, err := makeNodeClient(v.ListenAddr) 	// We need to have the remote addr of the node that is calling. 
	if err != nil {
		return nil, err
	}

	n.addPeer(c, v)
	

	return n.getVersion(), nil // In this handshake it's going to accept the connection yes or no. Because we are going to maintain a map of grpc stuff. 
}

func (n *Node) HandleTransaction (ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	fmt.Println("recieved tx from:", peer)
	return &proto.Ack{}, nil
}

func (n *Node) getVersion() *proto.Version {	// You are going to call version on this node, it basically means its going to return it's proto.Version
	return &proto.Version{
		Version: "blocker-0.1",
		Height: 0, 
		ListenAddr: n.listenAddr,
	}
}
func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	c, err := grpc.Dial(listenAddr, grpc.WithInsecure()) // grpc.WI.. Fixes rpc error: code = Unknown desc = grpc: no transport security set.
	if err != nil {
		return nil, err
	}
	return proto.NewNodeClient(c), nil
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

// Each time we recieve the Handshake message, we are going to add we are going to add a peer to our peer map.
// perrs map[proto.NodeLient]bool.

// So basically what is going to happen is that each time group at your nodes you are going to have a list of bootstrap nodes.
// And these nodes are basically, predefined nodes you know and they are going to bootstrap your network.
// You are going to connect with them and they are going to give more peers and more peers. Every block chain does this. Every peer to peer aplication is doing that.
// bootstrapNetwork().
