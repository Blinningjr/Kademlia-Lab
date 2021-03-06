package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func PostObjectRest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	data := vars["data"]
	if len(data) > 255 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "\nCan't send data that is longer than 255 characters (got %d characters).\n", len(data))
	} else {
		getRpcCh := make(chan GetRPCConfig)
		defer close(getRpcCh)
		go GetRPCMessageStarter(getRpcCh)

		sendToCh := make(chan SendToStruct)
		go SendToStarter(sendToCh)
		defer close(sendToCh)

		hash := sha1.Sum([]byte(data))
		hash_string := hex.EncodeToString(hash[:])

		w.WriteHeader(http.StatusCreated)

		payload := Payload{hash_string, []byte(data), maxExpire, nil, false}
		SendMessage(getRpcCh, sendToCh, CliPut, payload, w)
	}
}

func GetObjectRest(w http.ResponseWriter, r *http.Request) {
	getRpcCh := make(chan GetRPCConfig)
	defer close(getRpcCh)
	go GetRPCMessageStarter(getRpcCh)

	sendToCh := make(chan SendToStruct)
	go SendToStarter(sendToCh)
	defer close(sendToCh)

	vars := mux.Vars(r)
	hash := vars["hash"]
	payload := Payload{hash, []byte(hash), maxExpire, nil, false}
	SendMessage(getRpcCh, sendToCh, CliGet, payload, w)
	w.WriteHeader(http.StatusCreated)
}

func RunRestServer() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/objects/{hash}", GetObjectRest).Methods("GET")
	router.HandleFunc("/objects/{data}", PostObjectRest).Methods("POST")
	fmt.Printf("\nStarting Rest API Server...\n")
	http.ListenAndServe(resolveHostIp(":8081"), router)
}
