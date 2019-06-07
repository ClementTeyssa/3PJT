package defs

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	gonet "net"
	"net/http"
	"sync"
	"time"

	host "github.com/libp2p/go-libp2p-host"
)

///// FLAG & VARIABLES
var Ip *string
var Port *int
var Seed *int64
var Secio *bool
var Verbose *bool

const (
	BootstrapperAddr = "https://3pjt-dnode.infux.fr/"
)

var bootstrapperPort string

var Ha host.Host

type Transaction struct {
	AccountFrom string  `json:"accountfrom"`
	AccountTo   string  `json:"accountto"`
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

func GetMyIP(ipRange string) string {
	if ipRange == "local" {
		var MyIP string
		conn, err := gonet.Dial("udp", "8.8.8.8:80")
		if err != nil {
			log.Fatalln(err)
		} else {
			localAddr := conn.LocalAddr().(*gonet.UDPAddr)
			MyIP = localAddr.IP.String()
		}
		return MyIP
	} else {
		url := "https://api.ipify.org?format=text"
		fmt.Printf("Getting IP address from ipify\n")
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		MyIP, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		return string(MyIP)
	}

}

func GenRandInt(n int) int {
	myRandSource := rand.NewSource(time.Now().UnixNano())
	myRand := rand.New(myRandSource)
	val := myRand.Intn(n)
	return val
}

func ReadFlags() {
	// Parse options from the command line
	Ip = flag.String("ip", "global", "ip address resolution")
	Port = flag.Int("p", 0, "node's port")
	Seed = flag.Int64("seed", 0, "set random seed for id generation")
	Secio = flag.Bool("secio", false, "enable secio")
	Verbose = flag.Bool("verbose", false, "enable verbose")
	flag.Parse()
}
