package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
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
		log.Println("Received Interrupt. Exiting now.")
		cleanup(rw)
		os.Exit(1)
	}()
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
					fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
					blockchain.LastRcvdBlockchainLen = len(defs.Blockchain)
				}
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
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

	// stdReader := bufio.NewReader(os.Stdin)

	// for {
	// 	fmt.Print("> ")
	// 	sendData, err := stdReader.ReadString('\n')
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	sendData = strings.Replace(sendData, "\r", "", -1) + " (From terminal)"
	// 	acc := sendData
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	newBlock := blockchain.GenerateBlock(defs.Blockchain[len(defs.Blockchain)-1], acc, "", 0)

	// 	if blockchain.IsBlockValid(newBlock, defs.Blockchain[len(defs.Blockchain)-1]) {
	// 		defs.Mutex.Lock()
	// 		defs.Blockchain = append(defs.Blockchain, newBlock)
	// 		defs.Mutex.Unlock()
	// 	}

	// 	bytes, err := json.Marshal(defs.Blockchain)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}

	// 	spew.Dump(defs.Blockchain)

	// 	defs.Mutex.Lock()
	// 	rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	// 	rw.Flush()
	// 	blockchain.LastSentBlockchainLen = len(defs.Blockchain)
	// 	defs.Mutex.Unlock()
	// }
}
