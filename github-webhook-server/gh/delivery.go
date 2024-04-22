package gh

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
)

type Delivery struct {
	req  *http.Request
	Err  error
	sig  []byte
	body []byte
}

func (d Delivery) Hash() []byte {
	return []byte(d.req.Header[KeySig][0])
}

func (d Delivery) Body() []byte {
	return d.body
}

func (d Delivery) Remote() string {
	return d.req.RemoteAddr
}

func (d Delivery) Path() string {
	return d.req.URL.Path
}

func (d Delivery) Type() string {
	val, ok := d.req.Header[EventType]
	if !ok {
		// TODO: This situation should not occur, but otherwise panic on val[0]
		// Might need a more general parser for all the header fields
		val = []string{DefaultNotFound}
	}
	return val[0]
}

func (d *Delivery) VerifyHeader() bool {
	if d.req.Method != http.MethodPost {
		d.Err = fmt.Errorf("req.Method(%v) is not POST", d.req.Method)
		return false
	}
	sig, ok := d.req.Header[KeySig]
	if !ok || len(sig) != 1 || len(sig[0]) != 71 {
		d.Err = fmt.Errorf("req.%v(%v) not compatible (sha256=<64 chars hex>)", KeySig, sig)
		return false
	}
	signature, err := hex.DecodeString(sig[0][7:71])
	if err != nil {
		d.Err = fmt.Errorf("req.%v(%v) could not be decoded by hex ", KeySig, d.req.Header[KeySig])
		return false
	}
	d.req.Header[KeySig] = []string{string(signature)}
	contenttype, ok := d.req.Header[ContentType]
	if !ok || len(contenttype) != 1 || contenttype[0] != "application/json" {
		d.Err = fmt.Errorf("req.%v(%v) is not application/json", ContentType, contenttype)
		return false
	}
	if d.req.ContentLength > MaxBodyLength || d.req.ContentLength < 0 {
		d.Err = fmt.Errorf("req.ContentLength(%v) has size out of spec", d.req.ContentLength)
		return false
	}
	return true
}

func (d *Delivery) ReadBody() bool {
	// Server is defined with timeout on connection and request. Request has buffer and size limits. Headers have been checked for size. So should be safe to readall.
	bodybuffer, err := io.ReadAll(d.req.Body)
	if err != nil {
		d.Err = fmt.Errorf("ReadAll(req.Body) failed: %v", err)
		return false
	}
	if len(bodybuffer) != int(d.req.ContentLength) {
		d.Err = fmt.Errorf("ReadAll(req.Body) buffer length(%v) != ContentLength(%v)", len(bodybuffer), d.req.ContentLength)
		return false
	}
	// err == nil && len == contentlength
	d.body = bodybuffer
	return true
}

func New(req *http.Request) Delivery {
	return Delivery{req, nil, nil, nil}
}
