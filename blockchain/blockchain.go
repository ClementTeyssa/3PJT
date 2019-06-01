package blockchain

//Do imports
import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"time"

	defs "../defs"
)

var LastSentBlockchainLen = 0
var LastRcvdBlockchainLen = 0

// verify if new block is valid
func IsBlockValid(newBlock, oldBlock defs.Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		log.Println("BLOCKCHAIN ERROR: Index Mismatch")
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		log.Println("BLOCKCHAIN ERROR: Hash Inconsistent")
		return false
	}

	if CalculateHash(newBlock) != newBlock.Hash {
		log.Println("BLOCKCHAIN ERROR: Hash Mismatch")
		return false
	}

	return true
}

// calculate SHA256 hash
func CalculateHash(block defs.Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + block.Transaction.AccountFrom + block.Transaction.AccountTo + fmt.Sprintf("%.2f", block.Transaction.Amount) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func GenerateBlock(oldBlock defs.Block, accountFrom string, accountTo string, amount float32, address string) defs.Block {

	var newBlock defs.Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()

	var transaction defs.Transaction
	transaction.AccountTo = accountTo
	transaction.Amount = amount
	newBlock.Transaction = transaction
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = CalculateHash(newBlock)
	newBlock.Validator = address

	return newBlock
}

func GenerateGenesisBlock() defs.Block {
	genesisBlock := defs.Block{}
	var transaction defs.Transaction
	transaction.AccountTo = ""
	transaction.AccountFrom = ""
	transaction.Amount = 0
	genesisBlock = defs.Block{0, time.Now().String(), transaction, CalculateHash(genesisBlock), "", ""}
	genesisBlock.Hash = CalculateHash(genesisBlock)
	return genesisBlock
}
