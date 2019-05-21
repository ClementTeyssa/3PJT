package main

import (
	"log"

	defs "./defs"
	utils "./utils"
)

func init() { // Idea from https://appliedgo.net/networking/
	log.SetFlags(log.Lshortfile)
	defs.ReadFlags() // in defs.go
}

func main() {
	utils.P2pInit()              // Initialize P2P Network from Bootstrapper
	log.Fatal(utils.MuxServer()) // function is in mux.go
}
