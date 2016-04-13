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

func TestLogon(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			// unmarshal request
			env := new(envelope)
			logon := new(logonRequest)
			if err := xml.NewDecoder(req.Body).Decode(env); err != nil {
				t.Fatal("Error decoding logonRequest", err)
			}
			if err := xml.Unmarshal(env.Body.Payload, logon); err != nil {
				t.Fatal("Error unmarshaling payload", err)
			}
			if logon.Username != uname {
				t.Fatal("Unexpected data from logonRequest")
			}
			// return response
			resp.WriteHeader(http.StatusOK)
			payload, err := xml.Marshal(&logonResponse{Returnval: "000-test-000"})
			if err != nil {
				t.Fatal(err)
			}
			env = new(envelope)
			env.Body.Payload = payload
			xml.NewEncoder(resp).Encode(env)
		}),
	)
	defer server.Close()

	vb := NewVirtualBox(uname, password, server.URL)
	if err := vb.Logon(); err != nil {
		t.Fatal("Logon failed", err)
	}
	if vb.id != "000-test-000" {
		t.Fatal("Failed to get session id from logon")
	}
}
