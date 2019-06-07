package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	blockchain "../blockchain"
	defs "../defs"
	net "github.com/libp2p/go-libp2p-net"
)

func handleStream(s net.Stream) {

	log.Println("Got a new stream!")
	log.Println("New list of peers =", defs.Ha.Peerstore().Peers())

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw)
	go writeData(rw)

	// stream 's' will stay open until you close it (or the other side closes it).

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		Close(rw)
	}()
}

func Close(rw *bufio.ReadWriter) {
	var i int
	for i = 0; i < len(defs.Nodes)-1; i++ {
		if defs.Nodes[i].PhAddr == defs.MyNode.PhAddr {
			break
		}
	}
	defs.Nodes = append(defs.Nodes[:i], defs.Nodes[i+1:]...)
	log.Println("Received Interrupt. Exiting now.")
	CleanAddr()
	cleanup(rw)
	os.Exit(1)
}

func CleanAddr() {
	log.Println("Cleaning address")

	jsonValue, err := json.Marshal(peerProfile)
	if err != nil {
		log.Println(err)
		return
	}

	url := defs.BootstrapperAddr + "remove-peer"
	_, err = http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Println(err)
		return
	}
}

func readData(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Println(err)
		}

		if str == "" {
			return
		}
		if str != "Exit\n" {

			chain := make([]defs.Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}

			defs.Mutex.Lock()
			if len(chain) > len(defs.Blockchain) {
				defs.Blockchain = chain
				bytes, err := json.MarshalIndent(defs.Blockchain, "", "  ")
				if err != nil {

					log.Fatal(err)
				}
				if len(defs.Blockchain) > blockchain.LastRcvdBlockchainLen {
					fmt.Printf("%s ", string(bytes))
					blockchain.LastRcvdBlockchainLen = len(defs.Blockchain)
				}
				fmt.Printf("%s ", string(bytes))
			}
			defs.Mutex.Unlock()
		}
	}
}

func writeData(rw *bufio.ReadWriter) {

	go func() {
		for {
			time.Sleep(15 * time.Second)
			defs.Mutex.Lock()
			bytes, err := json.Marshal(defs.Blockchain)
			if err != nil {
				log.Println(err)
			}
			defs.Mutex.Unlock()

			defs.Mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			if len(defs.Blockchain) > blockchain.LastSentBlockchainLen {
				fmt.Sprintf("%s\n", string(bytes))
				blockchain.LastSentBlockchainLen = len(defs.Blockchain)
			}
			defs.Mutex.Unlock()

		}
	}()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if strings.Contains(sendData, "Exit") {
			Close(rw)
		}
	}
}
