package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Http status codes
const ok200 = http.StatusOK
const created201 = http.StatusCreated
const err400 = http.StatusBadRequest
const err404 = http.StatusNotFound
const err500 = http.StatusInternalServerError
const err501 = http.StatusNotImplemented

var err error

func registerHanders(handlers map[string]func(http.ResponseWriter, *http.Request)) {
	for path, handlerFunc := range handlers {
		http.HandleFunc(path, allowCORS(handlerFunc))
	}
}

func allowCORS(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[request] %s | [ip] %s", r.URL.RequestURI(), r.RemoteAddr)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")
		handler(w, r)
	}
}

// symbolsHandler responds with "Not implemented - 501" as this feature is currently not planned
func respondNotImplemented(w http.ResponseWriter, r *http.Request) {
	log.Printf("[request] %s", r.URL.Path)
	respondError(w, "", err501)
}

func respondJSON(w http.ResponseWriter, content interface{}, statusCode int) {
	b, err := json.Marshal(content)
	if respondIfError(err, w, "Something went wrong!", err500) {
		return
	}
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(b)
	if err != nil {
		log.Println("[response] [error]", err)
		return
	}
	log.Printf("[response] [status%d] %s\n", statusCode, http.StatusText(statusCode))
}

func respondIfError(err error, w http.ResponseWriter, msg string, statusCode int) bool {
	if err == nil {
		return false
	}
	respondError(w, msg, statusCode)
	return true
}

func respondError(w http.ResponseWriter, msg string, statusCode int) {
	if statusCode == 0 {
		statusCode = err400
	}
	if msg == "" {
		msg = http.StatusText(statusCode)
	}
	http.Error(w, msg, statusCode)
	log.Printf("[response] [status%d] %s\n", statusCode, msg)
}

func panicIf(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %+v\n", msg, err)
		panic(err)
	}
}
