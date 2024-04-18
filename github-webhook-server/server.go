package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	gh "github.com/rjwvandenberg/simpleservers/github-webhook-server/gh"
	"github.com/rjwvandenberg/simpleservers/github-webhook-server/secrets"
)

const (
	port            = 8080
	timeoutMs       = 5000
	maxHeaderLength = 4096 // guessed
	noStatusCode    = -1
)

// Github Webhook docs:
// https://docs.github.com/en/webhooks
func main() {
	log.Println("Starting github webhooks server...")

	validationHandlers := make(map[string]webhookValidationHandler)
	for _, obj := range []struct {
		path   string
		secret string
	}{{"/testa/", "A"}, {"/testb/", "B"}} {
		log.Printf("added: %v webhook handler ", obj.path)
		validationHandlers[obj.path] = webhookValidationHandler{obj.path, secrets.New(obj.path, []byte(obj.secret))}
	}

	// TODO: review how the timeouts interact exactly, encountered a read for EOF that would block for more data when Content-Length == body.length in delivery.ReadBody
	// In the meantime, added a TimeoutHandler to ensure the connection closes
	server := http.Server{
		Addr:         fmt.Sprintf(":%v", port),
		Handler:      http.TimeoutHandler(http.MaxBytesHandler(webhookRequestHandler{validationHandlers}, maxHeaderLength+gh.MaxBodyLength), timeoutMs*time.Millisecond, "yap timeout"),
		ReadTimeout:  timeoutMs * time.Millisecond,
		WriteTimeout: timeoutMs * time.Millisecond,
		IdleTimeout:  timeoutMs * time.Millisecond,
	}

	log.Printf("Listening to requests on :%v", port)
	server.ListenAndServe()
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
	log.Printf("accepted delivery %v", d.delivery.Type())
}
