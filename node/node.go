package node

import (
	"context"
	"encoding/hex"
	// "fmt"
	"net"
	"sync"
	"time"

	"github.com/Fito305/blocker/crypto"
	"github.com/Fito305/blocker/proto"
	"github.com/Fito305/blocker/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const blockTime = time.Second * 5

// A Mempool is just a pool in memory of known transactions. For example, if we are playing a game, and we are playing a numbers game, if I'm telling you numbers and we need to go from 1 - 10 but you cannot have any duplicates. What are you going to do, you are going to remember the numbers you choose because you cannot choose the same one again.
// So a Mempool is each time I'm sending a transaction, I'm going to remember that transaction in my memory. So the next time some other dude is sending me the same transaction, because it's a peer to peer protocol it could be that there is some delay and I already recieved a transaction from Bob but Alice transaction takes a longer round trip, I aleady have a transaction from Bob so i don't need to have the same transaction from alice so I can just drop it.
// You can make a Mempool as compact as you want.
type Mempool struct {
	lock sync.RWMutex
	txx map[string]*proto.Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		txx: make(map[string]*proto.Transaction),
	}
}

func (pool *Mempool) Clear() []*proto.Transaction { // We are going to clear the Mempool but we are also going to return a slice of proto.Transactions because we need it to add more blocks.
	pool.lock.Lock()
	defer pool.lock.Unlock()

	txx := make([]*proto.Transaction, len(pool.txx))
	it := 0
	for k, v := range pool.txx {
		delete(pool.txx, k)
		txx[it] = v
		it++
	}
	return txx 
}

func (pool *Mempool) Len() int {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	return len(pool.txx)
}

func (pool *Mempool) Has(tx *proto.Transaction) bool {
	pool.lock.RLock() // read lock
	defer pool.lock.RUnlock()

	// So what is happening here, we are making the hash representation string from the transaction hash, you are going to hash it. make it a nice hash string so we can use it a map (you cannot use bytes as a key in a map in go). So we are going to use a string. And we are going to check if we already have it yes or no. And we return that value.
	hash := hex.EncodeToString(types.HashTransaction(tx))
	_, ok := pool.txx[hash]
	return ok
}

func (pool *Mempool) Add(tx *proto.Transaction) bool {
	if pool.Has(tx) {
		return false // if we already have it return false because we did Add() it.
	}

	pool.lock.Lock()
	defer pool.lock.Unlock()

	hash := hex.EncodeToString(types.HashTransaction(tx))
	pool.txx[hash] = tx
	return true // In this case if we don't have it, we are going to Add() it. These bools allows use a skip in checks in HandleTransaction()
}

type ServerConfig struct {
	Version    string
	ListenAddr string
	PrivateKey *crypto.PrivateKey // Validator key
}

type Node struct {
	ServerConfig // embedd it.
	logger       *zap.SugaredLogger

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version
	mempool  *Mempool

	proto.UnimplementedNodeServer
}

func NewNode(cfg ServerConfig) *Node {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	logger, _ := loggerConfig.Build()
	return &Node{
		peers:        make(map[proto.NodeClient]*proto.Version),
		logger:       logger.Sugar(),
		mempool:      NewMempool(),
		ServerConfig: cfg,
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
		"ourNode:", n.ListenAddr,
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
		if addr == n.ListenAddr {
			continue
		}
		n.logger.Debugw("dialing remote node", "we", n.ListenAddr, "remoteNode", addr)
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
	n.ListenAddr = listenAddr

	var (
		opts       = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opts...)
	)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	proto.RegisterNodeServer(grpcServer, n)
	n.logger.Infow("node started...", "port", n.ListenAddr)

	//  bootstrap the network with a list of already known nodes
	// in the network.
	if len(bootstrapNodes) > 0 {
		go n.bootstrapNetwork(bootstrapNodes)
	}

	if n.PrivateKey != nil {
		go n.validatorLoop()
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

	// Wrap with this if statement to fix infinite loop, before we actually broadcast we are going to check .
	if n.mempool.Add(tx) { // If we added the mempool then we are going to broadcast it.
		n.logger.Debugw("received tx", "from", peer.Addr, "hash", hash, "we", n.ListenAddr)
		go func() { // because if we go go boradcast() we will never see the error. Here we handle the error.
			if err := n.broadcast(tx); err != nil {
				n.logger.Errorw("broadcast error", "err", err)
			}
		}()
	}

	return &proto.Ack{}, nil
}

func (n *Node) validatorLoop() {
	n.logger.Infow("starting validator loop", "pubkey", n.PrivateKey.Public(), "blockTime", blockTime)
	ticker := time.NewTicker(blockTime)
	for {
		<-ticker.C

		txx := n.mempool.Clear() // We are going to clear the mempool, with the transactions, and these transactions we are going to froge into a block.
		n.logger.Debugw("time to create a new block", "lenTx", len(txx))
	}
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
		ListenAddr: n.ListenAddr,
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
	if n.ListenAddr == addr {
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

// NOTE: Mutexes are slow. So how far you can go without using them.
// Who is making transactions? We the normal people are making transactions. People are going to make transactions
// by posting it in wallets. By posting it in wallets through JSON API. We post that to a node and that node(server) is going to
// validate it and its going to broadcast it and put it into the mempool and then we have validators.
// Validators normally in a proof of work, the first to solve the puzzle will be able to forge the block and get the reward.
// in our case in a proof of stake, we are going to have a consensus mechanism that is basically going to
// determine who for this round is going to forge a block. Because we cannot all forge a block. Let's say if we are in a room with
// 5 people and only the leader can do something, only the leader can press the red button, to escape in a an escape room. Only the
// leader can do that but who is the leader? Is everybody going directly to this button we are going to clap each other's cheeks
// because you can't. There needs to be a leader. So how do you come to a decision of who is a leader? Well it is by consensus.
// You are going to discuss with each other and at a certain point of time they do a vote and choose a leader. That is the same thing here
// that node can forge a block.

// Not everybody can forge a block out of the mempool. We need to have some sort of a consensus. What we going to do to mimick that we are going to
// make a config and we are going to put a rpivate key in the config and the guy with the private key in his config, because he needs a private key to sign the blocks,
// that is going to be the validator. Additionally, each time we cealr the block we need to delete the transaction from the mempool.

// Tx is short for transaction.
