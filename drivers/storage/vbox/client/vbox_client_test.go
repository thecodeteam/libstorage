package client

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	uname    = ""
	password = ""
)

type testTag struct {
	Value string `xml:"val"`
}

func TestNewVirtualBox(t *testing.T) {
	vb := NewVirtualBox("uname", "password", "http://test/")
	if vb.username != "uname" {
		t.Fatal("Username not set")
	}
	if vb.password != "password" {
		t.Fatal("Password not set")
	}
	if vb.vbURL != "http://test/" {
		t.Fatal("URL not set")
	}
}

func TestSend_NotOK(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusForbidden)
		}),
	)
	defer server.Close()
	vb := NewVirtualBox(uname, password, server.URL)
	resp := new(string)
	if err := vb.send("test", resp); err == nil {
		t.Fatal("Expected failure")
	}
}
func TestSend_OK(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			payload, err := xml.Marshal(&testTag{Value: "Test"})
			if err != nil {
				t.Fatal(err)
			}
			env := new(envelope)
			env.Body.Payload = payload
			xml.NewEncoder(resp).Encode(env)
		}),
	)
	defer server.Close()

	vb := NewVirtualBox(uname, password, server.URL)
	resp := new(testTag)
	if err := vb.send("test", resp); err != nil {
		t.Fatal("Unexpected failure:", err)
	}
	if resp.Value != "Test" {
		t.Fatal("Failed to process xml response properly")
	}
}
