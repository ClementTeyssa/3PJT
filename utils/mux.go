package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	gonet "net"
	"net/http"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"

	blockchain "../blockchain"
	defs "../defs"
)

type GoodResult struct {
	Good string `json:"good"`
}

type MyError struct {
	Error string `json:"error"`
}

type Transaction struct {
	AccountFrom string  `json:"accountfrom"`
	AccountTo   string  `json:"accountto"`
	Amount      float32 `json:"amount"`
	PrivateKey  []byte  `json:"privatekey"`
}

// create http handlers
func makeMuxRouter() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", handleGetBlockchain).Methods("GET")
	router.HandleFunc("/connect", handleConnect).Methods("POST")
	router.HandleFunc("/verif-transac", verifTransaction).Methods("POST")
	router.HandleFunc("/gen-block", generateBlock).Methods("POST")
	return router
}

// web server
func MuxServer() error {
	mux := makeMuxRouter()
	log.Println("HTTP Server Listening on " + GetMyIP() + ":" + strconv.Itoa(peerProfile.PeerPort)) // peerProfile.PeerPort in peer-manager.go
	s := &http.Server{
		Addr:           ":" + strconv.Itoa(peerProfile.PeerPort),
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

// write blockchain when we receive an http request
func handleGetBlockchain(writter http.ResponseWriter, request *http.Request) {
	defs.Mutex.Lock()
	bytes, err := json.MarshalIndent(defs.Blockchain, "", " ")

	if err != nil {
		http.Error(writter, err.Error(), http.StatusInternalServerError)
		return
	}
	defs.Mutex.Unlock()
	io.WriteString(writter, string(bytes))
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m defs.NewTargetJson

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		log.Println(err)
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}

	defer r.Body.Close()

	log.Println("MUX NewTarget =", m.NewTarget)
	connect2Target(m.NewTarget)
	respondWithJSON(w, r, http.StatusCreated, m.NewTarget)
}

func verifTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var transac Transaction
	var goodRes GoodResult
	var myErr MyError
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&transac); err != nil {
		log.Println(err)
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()
	data, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(data, &transac)
	jsonValue, _ := json.Marshal(transac)
	response, err := http.Post("https://3pjt-api.infux.fr/transactions/verify", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		err = json.Unmarshal(data, &goodRes)
		if goodRes.Good == "" {
			json.Unmarshal(data, &myErr)
			respondWithJSON(w, r, http.StatusCreated, myErr)
		} else {
			respondWithJSON(w, r, http.StatusCreated, goodRes)
		}
	}
}

func generateBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var transac Transaction
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&transac); err != nil {
		log.Println(err)
		return
	}
	defer r.Body.Close()
	data, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(data, &transac)
	newBlock := blockchain.GenerateBlock(defs.Blockchain[len(defs.Blockchain)-1], transac.AccountFrom, transac.AccountTo, transac.Amount)

	if blockchain.IsBlockValid(newBlock, defs.Blockchain[len(defs.Blockchain)-1]) {
		defs.Mutex.Lock()
		defs.Blockchain = append(defs.Blockchain, newBlock)
		defs.Mutex.Unlock()
	}

	_, err := json.Marshal(defs.Blockchain)
	if err != nil {
		log.Println(err)
	}

	spew.Dump(defs.Blockchain)

	defs.Mutex.Lock()
	blockchain.LastSentBlockchainLen = len(defs.Blockchain)
	defs.Mutex.Unlock()

	respondWithJSON(w, r, http.StatusCreated, newBlock)
}

// respond with JSON
func respondWithJSON(writter http.ResponseWriter, request *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		writter.WriteHeader(http.StatusInternalServerError)
		writter.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	writter.WriteHeader(code)
	writter.Write(response)
}

func GetMyIP() string {
	var MyIP string

	conn, err := gonet.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatalln(err)
	} else {
		localAddr := conn.LocalAddr().(*gonet.UDPAddr)
		MyIP = localAddr.IP.String()
	}
	return MyIP
}
