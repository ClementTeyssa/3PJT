package blockchain

//Do imports
import (
	"fmt"

	"./block"

	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Blockchain is an array of Blocks
var Blockchain []block.Block

var mutex = &sync.Mutex{}

//TODO: faire en sorte de ne pas avoir à avoir le .env dans le dossier courrant
//(récupérer tout ça dans le package main) et mettre en variables
func Launch() {
	//log errors
	err := godotenv.Load()

	if err != nil {
		log.Fatal(err)
	}

	go createFirstBlock()

	log.Fatal(run())
}

// web server
func run() error {
	mux := makeMuxRouter()
	httpPort := os.Getenv("PORT")
	log.Println("HTTP Server Listening on port :", httpPort)
	s := &http.Server{
		Addr:           ":" + httpPort,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

// create firstblock
func createFirstBlock() {
	t := time.Now()
	genesisBlock := block.Block{}
	transaction := block.Transaction{}
	genesisBlock = block.Block{0, t.String(), transaction, calculateHash(genesisBlock), ""}
	spew.Dump(genesisBlock)

	mutex.Lock()
	Blockchain = append(Blockchain, genesisBlock)
	mutex.Unlock()
}

// create http handlers
func makeMuxRouter() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", requestGetBlockchain).Methods("GET")
	router.HandleFunc("/", handleWriteBlock).Methods("POST")
	return router
}

// write blockchain when we receive an http request
func requestGetBlockchain(writter http.ResponseWriter, request *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", " ")

	if err != nil {
		http.Error(writter, err.Error(), http.StatusInternalServerError)
		return
	}

	io.WriteString(writter, string(bytes))
}

// read a new transaction
func handleWriteBlock(writter http.ResponseWriter, request *http.Request) {
	writter.Header().Set("Content-Type", "application/json")
	var transaction block.Transaction

	decoder := json.NewDecoder(request.Body)

	err := decoder.Decode(&transaction)
	if err != nil {
		respondWithJSON(writter, request, http.StatusBadRequest, request.Body)
		return
	}

	defer request.Body.Close()

	mutex.Lock()
	prevBlock := Blockchain[len(Blockchain)-1]
	newBlock := generateBlock(prevBlock, transaction.AccountFrom, transaction.AccountTo, transaction.Amount)

	if isBlockValid(newBlock, prevBlock) {
		Blockchain = append(Blockchain, newBlock)
		spew.Dump(Blockchain)
	}
	mutex.Unlock()

	respondWithJSON(writter, request, http.StatusCreated, newBlock)

}

// respond with JSON
func respondWithJSON(writter http.ResponseWriter, request *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		writter.WriteHeader(http.StatusInternalServerError)
		writter.Write([]byte("HTTP 500: Internal Server Error"))
		return
	} else {
		writter.WriteHeader(code)
		writter.Write(response)
	}
}

// verify if new block is valid
func isBlockValid(newBlock, oldBlock block.Block) bool {
	if oldBlock.Index+1 != newBlock.Index || oldBlock.Hash != newBlock.PrevHash || calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// calculate SHA256 hash
func calculateHash(block block.Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + block.Transaction.AccountFrom + block.Transaction.AccountTo + fmt.Sprintf("%.2f", block.Transaction.Amount) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func generateBlock(prevBlock block.Block, accountFrom string, accountTo string, amount float32) block.Block {

	var newBlock block.Block

	t := time.Now()

	newBlock.Index = prevBlock.Index + 1
	newBlock.Timestamp = t.String()

	var transaction block.Transaction
	transaction.AccountFrom = accountFrom
	transaction.AccountTo = accountTo
	transaction.Amount = amount
	newBlock.Transaction = transaction
	newBlock.PrevHash = prevBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock
}
