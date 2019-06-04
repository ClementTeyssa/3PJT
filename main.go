package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	defs "./defs"
	utils "./utils"
)

func init() { // Idea from https://appliedgo.net/networking/
	log.SetFlags(log.Lshortfile)
	defs.ReadFlags() // in defs.go
}

func main() {
	var registered bool
	registered = false
	var email, password string
	for !registered {
		fmt.Printf("Email: ")
		fmt.Scan(&email)
		fmt.Printf("Password: ")
		fmt.Scan(&password)
		defs.MyUser.Email = email
		defs.MyUser.Password = password
		jsonValue, _ := json.Marshal(defs.MyUser)
		response, err := http.Post("https://3pjt-api.infux.fr/login", "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
			return
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			json.Unmarshal(data, &defs.MyNode)
		}
		if defs.MyNode.TxAddr != "" {
			registered = true
		}
	}

	utils.P2pInit()              // Initialize P2P Network from Bootstrapper
	log.Fatal(utils.MuxServer()) // function is in mux.go
}
