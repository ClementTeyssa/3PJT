package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	defs "./defs"
	utils "./utils"
)

func init() {
	log.SetFlags(log.Lshortfile)
	defs.ReadFlags()
}

func main() {
	var registered bool
	registered = false
	for !registered {
		email, password := credentials()
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

func credentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err == nil {
		fmt.Println("\nPassword typed: " + string(bytePassword))
	}
	password := string(bytePassword)

	return strings.TrimSpace(username), strings.TrimSpace(password)
}
