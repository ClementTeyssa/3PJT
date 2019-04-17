package block

// Transaction represents each transaction in a block
type Transaction struct {
	AccountFrom string
	AccountTo   string
	Amount      float32
}
