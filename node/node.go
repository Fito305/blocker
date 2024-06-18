package node

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
	version    string
	listenAddr string
	logger     *zap.SugaredLogger

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version

	proto.UnimplementedNodeServer
}

func NewNode() *Node {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	logger, _ := loggerConfig.Build()
	return &Node{
		peers:   make(map[proto.NodeClient]*proto.Version),
		version: "blocker-0.1",
		logger:  logger.Sugar(),
	}
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	// Handle the Logic where we decide to accept or drop
	// the incoming node connection. it's simple logic to see if we have max connections or not.

	n.peers[c] = v // Here we add the peer. It's basically accepted.

	// Connect to all peers in the recieved list of peers
	// loop through the list of peers, see if you have the version and connect with them.
	if len(v.PeerList) > 0 {
		go n.bootstrapNetwork(v.PeerList)
	}

	// Print out here who are we as well [%s]. Our listen Address
	n.logger.Debugw("new peer successfully connected",
		"ourNode:", n.listenAddr,
		"remoteNode:", v.ListenAddr,
		"height:", v.Height) // You'll see the address and nodes connected to each other due to PEER DISCOVERY.
}

func (n *Node) deletePeer(c proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	delete(n.peers, c)
}

func (n *Node) bootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		if !n.canConnectWith(addr) {
			continue
		}
		if addr == n.listenAddr {
			continue
		}
	n.logger.Debugw("dialing remote node", "we", n.listenAddr, "remoteNode", addr) 
		c, v, err := n.dialRemoteNode(addr)
		if err != nil {
			return err
		}
		// There are a couple of things we can do. Do not return the hanshake but each time we recieve a handshake we could call handshake back.
		// But in this case, do this: because v is the version you get back from the guy. So we can add the connect 'c' with the version 'v'.
		n.addPeer(c, v)
	}
	return nil
}

func (n *Node) Start(listenAddr string, bootstrapNodes []string) error {
	n.listenAddr = listenAddr

	var (
		opts       = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opts...)
	)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	proto.RegisterNodeServer(grpcServer, n)
	n.logger.Infow("node started...", "port", n.listenAddr)

	//  bootstrap the network with a list of already known nodes
	// in the network.
	if len(bootstrapNodes) > 0 {
		go n.bootstrapNetwork(bootstrapNodes)
	}

	return grpcServer.Serve(ln)
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	c, err := makeNodeClient(v.ListenAddr) // We need to have the remote addr of the node that is calling.
	if err != nil {
		return nil, err
	}

	n.addPeer(c, v)

	return n.getVersion(), nil // In this handshake it's going to accept the connection yes or no. Because we are going to maintain a map of grpc stuff.
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	fmt.Println("recieved tx from:", peer)
	return &proto.Ack{}, nil
}

func (n *Node) dialRemoteNode(addr string) (proto.NodeClient, *proto.Version, error) {

	c, err := makeNodeClient(addr)
	if err != nil {
		return nil, nil, err
	}
	v, err := c.Handshake(context.Background(), n.getVersion())
	if err != nil {
		return nil, nil, err
	}
	return c, v, nil
}

func (n *Node) getVersion() *proto.Version { // You are going to call version on this node, it basically means its going to return it's proto.Version
	return &proto.Version{
		Version:    "blocker-0.1",
		Height:     0,
		ListenAddr: n.listenAddr,
		PeerList:   n.getPeerList(),
	}
}

func (n *Node) getPeerList() []string { // returns a slice of strings because our peerlist is a map above
	n.peerLock.RLock() // Read Lock
	defer n.peerLock.RUnlock()

	peers := []string{}
	for _, version := range n.peers {
		peers = append(peers, version.ListenAddr)
	}
	return peers
}

func (n *Node) canConnectWith(addr string) bool {
	// We are going to check if we can connect to a certain address. We are going to check is this address the same as ours? Then we don't connect. Then we are going to loop through all of our connected peers and check if the address we are trying to connect to is already in that list. If its already in that list we don't connect. If it is not in that list we can connect. 
	if n.listenAddr == addr {
		return false
	}

	connectedPeers := n.getPeerList()

	for _, connectedAddr := range connectedPeers {
		if addr == connectedAddr {
			return false
		}
	}
	return true
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

// Each time we are going to recieve a peer, the Handshake() is when someone is dailing to us. They are going to call Handshake().
// And we are going to makeNodeClient connection out of the context. The handshake is only getting handled with nodes that are hand shaking with us
// doing an outbounding dail to us. So you don't want to do logic code inside the handshake() because it will only run when
// an outside node dails us. Not when we dail them. That is bad. In this case addPeer() is a good place for that logic because the other functions
// like bootstrapNode() and others are calling it. So by doing the logic in addPeers you are going to handle both sides of the case. When we dail and when we get a dail.
// **THATS VERY IMPORTANT**

// bootstrapNetwork() should be private. So what is going to happen is basically, is that we are going to make a config object passed as a paramter bootstrapNodes.
// and instead of calling it in main.go, we will pass it via n.Start() in makeNode() in main and use it here in Node.go.

// NOTE: If you need to return more than 3 variables from a function it could be a very good use case to wrap them in a struct. So you can return just that struct and an error.

// Peers and remote nodes are the samething. A peer is something we connect to that is connected to you. And a remote node (what is a node?) is actually a sever
// So there is a lot of terminology for the samethings. Peer/server/Remote node. 

// It's going to be an asynchonous envirnment where we are going to send these messages async and you need to be prepared. 
// You can never assume that a message is coming at a certain point you have to expect the message coming. An ideal timeframe for when the message is coming.
// A message can come at any time at a certain point of time. It does it async. We don't know in what order it is coming in. So what is going to happen is
// we are going to have peer lists with ourselves in it. And it is also going to have pper list recieved with nodes in them that we are already connected to. 
