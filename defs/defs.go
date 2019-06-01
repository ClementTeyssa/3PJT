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

//var ListenF *int
//var Target *string
var Secio *bool
var Verbose *bool
var Seed *int64
var DataDir *string
var BootstrapperAddr *string

const (
	bootstrapperPort = "51000"
)

var Ha host.Host

type Transaction struct {
	AccountFrom string
	AccountTo   string
	Amount      float32
}

// Block represents each 'item' in the blockchain
type Block struct {
	Index       int
	Timestamp   string
	Transaction Transaction
	Hash        string
	PrevHash    string
	Validator   string
}

// Blockchain is a series of validated Blocks
var Blockchain []Block

var TempBlocks []Block

// validators keeps track of open validators and balances
var Validators = make(map[string]int)

var Winner string

// Message takes incoming JSON payload for writing heart rate
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
	Secio = flag.Bool("secio", false, "enable secio")
	Verbose = flag.Bool("verbose", false, "enable verbose")
	Seed = flag.Int64("seed", 0, "set random seed for id generation")
	DataDir = flag.String("data", "data", "pathname of data directory")
	BootstrapperAddr = flag.String("b", "local", "Address of bootstrapper")
	flag.Parse()

	if *BootstrapperAddr == "local" {
		*BootstrapperAddr = "http://localhost:" + bootstrapperPort + "/"
	} else {
		*BootstrapperAddr = "http://" + *BootstrapperAddr + ":" + bootstrapperPort + "/"
	}
}
