package utils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"

	blockchain "../blockchain"
	defs "../defs"
)

// create http handlers
func makeMuxRouter() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", handleGetBlockchain).Methods("GET")
	router.HandleFunc("/", handleWriteBlock).Methods("POST")
	router.HandleFunc("/connect", handleConnect).Methods("POST")
	return router
}

// web server
func MuxServer() error {
	mux := makeMuxRouter()
	httpPort := os.Getenv("PORT")
	log.Println("HTTP Server Listening on port :", peerProfile.PeerPort) // peerProfile.PeerPort in peer-manager.go
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

// read a new transaction
func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m defs.Message

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	defs.Mutex.Lock()
	newBlock := blockchain.GenerateBlock(defs.Blockchain[len(defs.Blockchain)-1], "", "", 0)
	defs.Mutex.Unlock()

	if blockchain.IsBlockValid(newBlock, defs.Blockchain[len(defs.Blockchain)-1]) {
		defs.Mutex.Lock()
		defs.Blockchain = append(defs.Blockchain, newBlock)
		defs.Mutex.Unlock()
		spew.Dump(defs.Blockchain)
	}

	respondWithJSON(w, r, http.StatusCreated, newBlock)
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

	log.Println("mux NewTarget =", m.NewTarget)
	connect2Target(m.NewTarget)
	respondWithJSON(w, r, http.StatusCreated, m.NewTarget)
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
