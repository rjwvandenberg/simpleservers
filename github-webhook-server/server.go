package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	gh "github.com/rjwvandenberg/simpleservers/github-webhook-server/gh"
	"github.com/rjwvandenberg/simpleservers/github-webhook-server/secrets"
)

const (
	port                = 8080
	connectionTimeoutMs = 10000
	requestTimeoutMs    = 5000
	maxHeaderLength     = 4096 // guessed
	noStatusCode        = -1
)

// Github Webhook docs:
// https://docs.github.com/en/webhooks
// Assumption loadbalancer is transcoding http2 to http1
func main() {
	log.Println("Starting github webhooks server (http1)...")

	validationHandlers := getHandlers()

	// TimeoutHandler will limit the time in serveHTTP and return 503 <msg> if exceeded, by attaching a cancel context to the http.Request
	// MaxBytesHandler uses a MaxBytesReader to wrap the request.Body io.Reader to limit the size of a request body. It is a 32kb buffered reader
	// A good writeup on timeout considerations https://adam-p.ca/blog/2022/01/golang-http-server-timeouts/
	// and check the source by following server.ListenAndServe
	server := http.Server{
		Addr:         fmt.Sprintf(":%v", port),
		Handler:      http.TimeoutHandler(http.MaxBytesHandler(webhookRequestHandler{validationHandlers}, maxHeaderLength+gh.MaxBodyLength), requestTimeoutMs*time.Millisecond, "<!DOCTYPE html><html><body>connection timed out</body></html>"),
		ReadTimeout:  connectionTimeoutMs * time.Millisecond, // Max time for header+body read (per connection)
		WriteTimeout: connectionTimeoutMs * time.Millisecond, // Max time for response write (per connection) resets on receiving new request
		// IdleTimeout:  timeoutMs * time.Millisecond,	// Max wait time for next request, uses ReadTimeout if 0.
		MaxHeaderBytes: maxHeaderLength, // Default 1MB
		// ErrorLog: ,									// Default is log.Print* functions
	}

	log.Printf("Listening to requests on :%v", port)
	server.ListenAndServe()
}

// env var SECRETS is a comma seperated list of path,base64(secret),path,base64(secret),.... webhooks
// so SECRETS=testa,A,testb,B gets converted into
// <domain>/testa/  authorized with secret A
// <domain>/testb/  authorized with secret B
func getHandlers() map[string]webhookValidationHandler {
	defer os.Clearenv()

	for _, ev := range os.Environ() {
		if s, found := strings.CutPrefix(ev, "SECRETS="); found {
			return processSecrets(strings.Split(s, ","))
		}
	}

	log.Panic("failed: did not find SECRETS env variable")
	return nil
}

func processSecrets(envSecrets []string) map[string]webhookValidationHandler {
	if len(envSecrets)%2 != 0 {
		log.Panic("failed: secrets slice length is odd")
	}

	handlers := make(map[string]webhookValidationHandler)
	for i := range len(envSecrets) / 2 {
		path := fmt.Sprintf("/%v/", strings.TrimSpace(envSecrets[2*i]))
		secret := strings.TrimSpace(envSecrets[2*i+1])
		log.Printf("added: %v webhook handler", path)
		handlers[path] = webhookValidationHandler{path, secrets.New(path, []byte(secret))}
	}

	return handlers
}

type webhookRequestHandler struct {
	validationHandlers map[string]webhookValidationHandler
}

func (w webhookRequestHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	validationHandler, ok := w.validationHandlers[req.URL.Path]
	if !ok {
		w.refuse(res, req, fmt.Errorf("req.URL.Path(%v) not registered as a path", req.URL.Path), http.StatusNotFound)
		return
	}

	delivery := gh.New(req)
	if !delivery.VerifyHeader() {
		w.refuse(res, req, delivery.Err, http.StatusNotFound)
		return
	}

	res.WriteHeader(http.StatusAccepted)
	if !delivery.ReadBody() {
		w.refuse(res, req, delivery.Err, noStatusCode)
		return
	}

	go validationHandler.validate(delivery)
}

func (w webhookRequestHandler) refuse(res http.ResponseWriter, req *http.Request, err error, statuscode int) {
	log.Printf("refused: %v at %v, because: %v", req.RemoteAddr, req.URL.Path, err)
	if statuscode != noStatusCode {
		res.WriteHeader(http.StatusNotFound)
	}
}

type webhookValidationHandler struct {
	path   string
	secret *secrets.SecretHandler
}

func (w webhookValidationHandler) validate(delivery gh.Delivery) {
	if !w.secret.Validate(delivery.Hash(), delivery.Body()) {
		log.Printf("refused: %v requesting %v, because invalid signature %v", delivery.Remote(), delivery.Path(), hex.EncodeToString(delivery.Hash()))
	} else {
		go DeliveryHandler{delivery}.process()
	}
}

type DeliveryHandler struct {
	delivery gh.Delivery
}

func (d DeliveryHandler) process() {
	log.Printf("accepted: delivery '%v'", d.delivery.Type())
}
