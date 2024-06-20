**Everytime you make a change to types.proto, you have to run make proto.
NOTES: place all notes here along with the file in which they were copied from.

We need to have some kind of a chain, a blockchain mechanism to store and retireive blocks and validate all that stuff.
Because then we can actually start filling up the blockchain, validating blocks, and actually making transactions. 

Main.go notes
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


// To create blocks you need to do it via proof of stake. Proof of work is old. 
// We are going to boradcast transactions.
Normally if you want to send a transaction, we going to do that but our json API we also need ot build. But because we have
grpc we could do it by just connecting with a grpc client to some of the nodes and push the transaction to it. 
And then it can validate the transaction and it can broadcast that to its known peers. 

node.go notes
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
