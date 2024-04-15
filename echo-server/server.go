package main

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const port = 8080
const maxReqLength = 1024
const read_ms = 1000
const write_ms = 1000
const salt_len = 128

func main() {
	server := http.Server{
		Addr:           fmt.Sprintf(":%v", port),
		Handler:        http.MaxBytesHandler(EchoHandler{magic: magic_number(salt_len)}, maxReqLength),
		ReadTimeout:    read_ms * time.Millisecond,
		WriteTimeout:   write_ms * time.Millisecond,
		MaxHeaderBytes: maxReqLength,
	}
	log.Println("Starting echo-server...")
	err := server.ListenAndServe()
	log.Fatalf("%v", err)
}

type EchoHandler struct {
	magic []byte
}

func (h EchoHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	split := strings.Split(req.RemoteAddr, ":")
	port := split[len(split)-1]
	quickunsafehash := fmt.Sprintf("%x", md5.Sum(append([]byte(strings.Join(split[:len(split)-1], "")), h.magic...)))

	log.Printf("served %v:%v - url length %v ", quickunsafehash, port, len(req.URL.Path))
	url := req.URL
	res.Write([]byte(url.Path))
}

func magic_number(salt_len int) []byte {
	random := make([]byte, salt_len)
	_, err := rand.Read(random)
	if err != nil {
		log.Fatalln("Could not generate random numbers.")
	}
	return random
}
