package block

// Block represents each 'item' in the blockchain
type Block struct {
	Index       int
	Timestamp   string
	Transaction Transaction
	Hash        string
	PrevHash    string
}
