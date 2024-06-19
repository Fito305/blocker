package node

import (
	"context"
	"encoding/hex"
	// "fmt"
	"net"
	"sync"

	"github.com/Fito305/blocker/proto"
	"github.com/Fito305/blocker/types"
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
	hash := hex.EncodeToString(types.HashTransaction(tx))

	n.logger.Debugw("received tx", "from", peer.Addr, "hash", hash)

	go func() { // because if we go go boradcast() we will never see the error. Here we handle the error.
		if err := n.broadcast(tx); err != nil {
			n.logger.Errorw("broadcast error", "err", err)
		}
	}()

	return &proto.Ack{}, nil
}

// The thing is because we don't have the concept of messages, we have the concept of proto types. So we need to say here if you want to broadcast something you pass in msg of any type.
func (n *Node) broadcast(msg any) error {
	for peer := range n.peers {
		// So we are going to loop through all the peers in our connection map (where ever you are keeping these proto clients), and for each client we find, we are going to call the remote procedure and it is going to be the HandleTrasaction(). Which means it is going to boradcast it again and probably again to us.
		switch v := msg.(type) {
		case *proto.Transaction:
			_, err := peer.HandleTransaction(context.Background(), v)
			if err != nil {
				return err
			}
		}
	}
	return nil
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
