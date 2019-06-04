package defs

import (
	"flag"
	"log"
	"math/rand"
	gonet "net"
	"sync"
	"time"

	host "github.com/libp2p/go-libp2p-host"
)

///// FLAG & VARIABLES
var Seed *int64
var Secio *bool
var Verbose *bool
var BootstrapperAddr *string

const (
	bootstrapperPort = "51000"
)

type GoodResult struct {
	Good string `json:"good"`
}

type MyError struct {
	Error string `json:"error"`
}

var Ha host.Host

type Transaction struct {
	AccountFrom string  `json:"accfrom"`
	AccountTo   string  `json:"accto"`
	Amount      float32 `json:"amount"`
}

type Status struct {
	Status string `json:"status"`
}

type Node struct {
	PhAddr string `json:"ipAdress"`
	TxAddr string `json:"adress"`
}

var MyNode Node

var Nodes []Node

// Block represents each 'item' in the blockchain
type Block struct {
	Index       int    `json:"id"`
	Timestamp   string `json:"timestamp"`
	Transaction Transaction
	Hash        string `json:"hash"`
	PrevHash    string `json:"prevhash"`
}

// Blockchain is a series of validated Blocks
var Blockchain []Block

// Message takes incoming JSON payload for writing heart rate
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var MyUser User

type Message struct {
	transaction Transaction
}

type NewTargetJson struct {
	NewTarget string
}

var Mutex = &sync.Mutex{}

////////  HELPER FUNCTIONS

func GetMyIP() string {
	var MyIP string

	conn, err := gonet.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatalln(err)
	} else {
		localAddr := conn.LocalAddr().(*gonet.UDPAddr)
		MyIP = localAddr.IP.String()
	}
	return MyIP
}

func GenRandInt(n int) int {
	myRandSource := rand.NewSource(time.Now().UnixNano())
	myRand := rand.New(myRandSource)
	val := myRand.Intn(n)
	return val
}

func ReadFlags() {
	// Parse options from the command line
	//ListenF = flag.Int("l", 0, "wait for incoming connections")
	//Target = flag.String("d", "", "target peer to dial")
	Seed = flag.Int64("seed", 0, "set random seed for id generation")
	Secio = flag.Bool("secio", false, "enable secio")
	Verbose = flag.Bool("verbose", false, "enable verbose")
	BootstrapperAddr = flag.String("b", "local", "Address of bootstrapper")
	flag.Parse()

	if *BootstrapperAddr == "local" {
		*BootstrapperAddr = "http://localhost:" + bootstrapperPort + "/"
	} else {
		*BootstrapperAddr = "http://" + *BootstrapperAddr + ":" + bootstrapperPort + "/"
	}
	//*BootstrapperAddr = "https://3pjt-dnode.infux.fr/"
}
