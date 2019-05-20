package utils

import (
	"../blockchain"

	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
)

// create http handlers
func makeMuxRouter() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", requestGetBlockchain).Methods("GET")
	router.HandleFunc("/", handleWriteBlock).Methods("POST")
	return router
}

// web server
func muxServer() error {
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

// write blockchain when we receive an http request
func handleGetBlockchain(writter http.ResponseWriter, request *http.Request) {
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

func handleConnect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m newTarget_json

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
