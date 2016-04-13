package client

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"time"
)

// VirtualBox Represents a virtualbox sesion
type VirtualBox struct {
	ConnectionTimeout time.Duration
	username          string
	password          string
	vbURL             string
	client            *http.Client
	transport         *http.Transport
}

type envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		XMLName xml.Name `xml:"Body"`
		Payload []byte   `xml:",innerxml"`
	}
}

// NewVirtualBox returns a reference to a VirtualBox value.
func NewVirtualBox(uname, pwd, url string) *VirtualBox {
	vb := &VirtualBox{
		username: uname,
		password: pwd,
		vbURL:    url,
		client: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   15 * time.Second,
			ExpectContinueTimeout: 3 * time.Second,
		},
		},
	}
	return vb
}

// WithTimeout sets connection timeout
func (vb *VirtualBox) WithTimeout(dur time.Duration) *VirtualBox {
	vb.client.Transport.(*http.Transport).Dial = (&net.Dialer{
		Timeout:   dur,
		KeepAlive: 30 * time.Second,
	}).Dial
	return vb
}

func (vb *VirtualBox) send(request, response interface{}) error {
	// encode request
	payload, err := xml.Marshal(request)

	if err != nil {
		return err
	}

	reqData := new(bytes.Buffer)
	env := new(envelope)
	env.Body.Payload = payload
	err = xml.NewEncoder(reqData).Encode(env)

	if err != nil {
		return err
	}

	// send req as http
	httpReq, err := http.NewRequest("POST", vb.vbURL, reqData)
	if err != nil {
		return err
	}
	httpReq.Header.Add("Content-Type", "text/xml; charset=\"utf-8\"")
	httpReq.Header.Set("User-Agent", "libstorage/0.1")

	httpResp, err := vb.client.Do(httpReq)
	if err != nil {
		return err
	}
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected http status: %d", httpResp.StatusCode)
	}
	defer httpResp.Body.Close()

	// decode response
	env = new(envelope)
	err = xml.NewDecoder(httpResp.Body).Decode(env)
	if err != nil {
		return fmt.Errorf("Error decoding response: %s", err)
	}

	err = xml.Unmarshal(env.Body.Payload, response)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal payload: %s", err)
	}
	return nil
}
