package gh

import (
	"encoding/hex"
	"errors"
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
	var totalBytesRead int64 = 0
	bodybuffer := make([]byte, MaxBodyLength)
	var err error
	for err != io.EOF && totalBytesRead < d.req.ContentLength { // at read==contentlength, need to read again to reach EOF
		// TODO: review Read, might need to set a timeout (on top of the Server timeout?)
		// because i'm unsure where to take care of this at the moment
		length, err := d.req.Body.Read(bodybuffer[totalBytesRead:])
		totalBytesRead += int64(length)

		if err != io.EOF && err != nil {
			d.Err = errors.Join(errors.New("readbody io error: "), err)
			return false
		}
	}
	if (err == io.EOF && totalBytesRead != d.req.ContentLength) ||
		(err != io.EOF && totalBytesRead > d.req.ContentLength) {
		d.Err = fmt.Errorf("read body invalid state (%v)", err)
		return false
	}

	d.body = append(make([]byte, 0), bodybuffer[:totalBytesRead]...) // TODO: review slice memory, copy slice to free memory?
	return true
}

func New(req *http.Request) Delivery {
	return Delivery{req, nil, nil, nil}
}
