package blockchain

//Do imports
import (
	"fmt"

	"./block"

	"crypto/sha256"
	"encoding/hex"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
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

	//log.Fatal(run())
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
