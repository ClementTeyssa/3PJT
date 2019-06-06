package utils

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"

	blockchain "../blockchain"
	defs "../defs"
	
	"github.com/phayes/freeport"
	"github.com/davecgh/go-spew/spew"
	golog "github.com/ipfs/go-log"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	gologging "github.com/whyrusleeping/go-logging"
)

var MaxPeerPort int

type Peer struct {
	PeerAddress string `json:"PeerAddress"`
}

type PeerProfile struct { // connections of one peer
	ThisPeer  Peer   `json:"ThisPeer"`  // any node
	PeerPort  int    `json:"PeerPort"`  // port of peer
	Neighbors []Peer `json:"Neighbors"` // edges to that node
	Status    bool   `json:"Status"`    // Status: Alive or Dead
	Connected bool   `json:"Connected"` // If a node is connected or not [To be used later]
}

var ThisPeerFullAddr string
var peerProfile PeerProfile                  // used to enroll THIS peer | connectP2PNet() & enrollP2PNet)()
var PeerGraph = make(map[string]PeerProfile) // Key = Node.PeerAddress; Value.Neighbors = Edges
var graphMutex sync.RWMutex

func P2pInit() {
	// LibP2P code uses golog to log messages. They log with different
	// string IDs (i.e. "swarm"). We can control the verbosity level for
	// all loggers with:
	golog.SetAllLoggers(gologging.INFO) // Change to DEBUG for extra info

	requestPort() // request THIS peer's port from bootstrapper
	if *defs.Verbose {
		log.Println("PeerIP = ", defs.GetMyIP())
	}
	queryP2PGraph() // query graph of peers in the P2P Network

	// Make a host that listens on the given multiaddress
	makeBasicHost(peerProfile.PeerPort, *defs.Secio, *defs.Seed)
	defs.Ha.SetStreamHandler("/p2p/1.0.0", handleStream)

	log.Println("Peerstore().Peers() before connecting =", defs.Ha.Peerstore().Peers())
	connectP2PNet()
	enrollP2PNet()
	log.Println("Peerstore().Peers() after connecting =", defs.Ha.Peerstore().Peers())
	sendNodeAddr()
}

func connect2Target(newTarget string) {
	log.Println("Attempting to connect to", newTarget)
	// The following code extracts target's peer ID from the
	// given multiaddress
	ipfsaddr, err := ma.NewMultiaddr(newTarget)
	if err != nil {
		log.Fatalln(err)
	}
	if *defs.Verbose {
		log.Printf("ipfsaddr = ", ipfsaddr)
	}
	if *defs.Verbose {
		log.Printf("Target = ", newTarget)
	}

	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Fatalln(err)
	}
	if *defs.Verbose {
		log.Printf("pid = ", pid)
	}
	if *defs.Verbose {
		log.Printf("ma.P_IPFS = ", ma.P_IPFS)
	}

	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Fatalln(err)
	}
	if *defs.Verbose {
		log.Println("peerid = ", peerid)
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)
	if *defs.Verbose {
		log.Printf("targetPeerAddr = ", targetPeerAddr)
	}
	if *defs.Verbose {
		log.Printf("targetAddr = ", targetAddr)
	}

	// We have a peer ID and a targetAddr so we add it to the peerstore
	// so LibP2P knows how to contact it
	defs.Ha.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

	log.Println("opening stream to", newTarget)
	// make a new stream from host B to host A
	// it should be handled on host A by the handler we set above because
	// we use the same /p2p/1.0.0 protocol
	s, err := defs.Ha.NewStream(context.Background(), peerid, "/p2p/1.0.0")
	if err != nil {
		log.Fatalln(err)
	}
	// Create a buffered stream so that read and writes are non blocking.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// Create a thread to read and write data.
	go writeData(rw)
	go readData(rw)

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		Close(rw)
	}()
	//select {} // hang forever
}

func requestPort() { // Requesting PeerPort
	log.Println("Requesting PeerPort from Bootstrapper", *defs.BootstrapperAddr)

	response, err := http.Get(*defs.BootstrapperAddr + "port-request")
	if err != nil {
		log.Println(err)
		log.Fatalln("PANIC: Unable to requestPort() from bootstrapper. Bootstrapper may be down.")
		return
	}
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		log.Fatalln("PANIC: Unable to requestPort() from bootstrapper. Bootstrapper may be down.")
		return
	}

	json.Unmarshal(responseData, &peerProfile.PeerPort)
	if *defs.Port != 0 {
		peerProfile.PeerPort = *defs.Port
	}
	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
	} else {
		peerProfile.PeerPort = port
	}

	if *defs.Verbose {
		log.Println("PeerPort = ", peerProfile.PeerPort)
	}

	if peerProfile.PeerPort == 0 {
		log.Println("PANIC: Exiting Program. PeerPort = 0. Bootstrapper may be down.")
		os.Exit(1)
	}
}

func queryP2PGraph() { // Query the graph of peers in the P2P Network from the Bootstrapper
	log.Println("Querying graph of peers from Bootstrapper", *defs.BootstrapperAddr)

	response, err := http.Get(*defs.BootstrapperAddr + "query-p2p-graph")
	if err != nil {
		log.Println(err)
		return
	}
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return
	}

	graphMutex.RLock()
	json.Unmarshal(responseData, &PeerGraph)
	if *defs.Verbose {
		log.Println("PeerGraph = ", PeerGraph)
		spew.Dump(PeerGraph)
	}
	graphMutex.RUnlock()
}

func connectP2PNet() {
	peerProfile.ThisPeer = Peer{PeerAddress: ThisPeerFullAddr}

	if len(PeerGraph) == 0 && len(defs.Nodes) == 0 { // first node in the network
		log.Println("I'm first peer. Creating Genesis Block.")
		defs.Blockchain = append(defs.Blockchain, blockchain.GenerateGenesisBlock())
		spew.Dump(defs.Blockchain)
		log.Println("I'm first peer. Listening for connections.")
	} else {
		log.Println("Connecting to P2P network")

		if *defs.Verbose {
			log.Println("Cardinality of PeerGraph = ", len(PeerGraph))
		}

		// make connection with peers[choice]
		choice := defs.GenRandInt(len(PeerGraph))
		log.Println("Connecting choice = ", choice)

		peers := make([]string, 0, len(PeerGraph))
		for p, _ := range PeerGraph {
			peers = append(peers, p)
		}
		log.Println("Connecting to", peers[choice])
		connect2Target(peers[choice])
		peerProfile.Neighbors = append(peerProfile.Neighbors, Peer{PeerAddress: peers[choice]})
		if *defs.Verbose {
			log.Println("peers[choice] = ", peers[choice])
		}
	}
}

func enrollP2PNet() { // Enroll to the P2P Network by adding THIS peer with Bootstrapper
	log.Println("Enrolling in P2P network at Bootstrapper", *defs.BootstrapperAddr)

	jsonValue, err := json.Marshal(peerProfile)
	if err != nil {
		log.Println(err)
		return
	}

	url := *defs.BootstrapperAddr + "enroll-p2p-net"
	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Println(err)
		return
	}
	defer response.Body.Close()

	if *defs.Verbose {
		log.Println("response Status:", response.Status)
	}
	if *defs.Verbose {
		log.Println("response Headers:", response.Header)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("response Body:", string(body))
}

// makeBasicHost creates a LibP2P host with a random peer ID listening on the
// given multiaddress. It will use secio if secio is true.
func makeBasicHost(listenPort int, secio bool, randseed int64) {

	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	if *defs.Verbose {
		log.Printf("r = ", r)
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		log.Fatal(err)
	}
	if *defs.Verbose {
		log.Printf("priv = ", priv)
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/"+defs.GetMyIP()+"/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	if *defs.Verbose {
		log.Printf("opts = ", opts)
	}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		log.Fatal(err)
	}

	if *defs.Verbose {
		log.Printf("basicHost = ", basicHost)
		log.Printf("basicHost.ID() = ", basicHost.ID())
		log.Printf("basicHost.ID().Pretty() = ", basicHost.ID().Pretty())
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	if *defs.Verbose {
		log.Printf("hostAddr = ", hostAddr)
	}

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	if *defs.Verbose {
		log.Printf("addr = ", addr)
	}
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("My fullAddr = %s\n", fullAddr)
	var re = regexp.MustCompile(`/ip4/*.*.*.*/tcp`)
	ThisPeerFullAddr = re.ReplaceAllString(fullAddr.String(), "/ip4/"+GetMyIP()+"/tcp")
	//ThisPeerFullAddr = fullAddr.String()
	//return basicHost, nil
	defs.Ha = basicHost // ha defined in defs.go
	if *defs.Verbose {
		log.Printf("basicHost = ", defs.Ha)
	}
}

func sendNodeAddr() {
	log.Println("Sending addresses to Bootstrapper", *defs.BootstrapperAddr)
	defs.MyNode.PhAddr = peerProfile.ThisPeer.PeerAddress
	defs.Nodes = append(defs.Nodes, defs.MyNode)
	jsonValue, err := json.Marshal(defs.MyNode)
	if err != nil {
		log.Println(err)
		return
	}

	url := *defs.BootstrapperAddr + "node-addr"
	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Println(err)
		return
	}
	defer response.Body.Close()

	if *defs.Verbose {
		log.Println("response Status:", response.Status)
	}
	if *defs.Verbose {
		log.Println("response Headers:", response.Header)
	}
}

func cleanup(rw *bufio.ReadWriter) {
	fmt.Println("cleanup")
	defs.Mutex.Lock()
	rw.WriteString("Exit\n")
	rw.Flush()
	defs.Mutex.Unlock()
}
