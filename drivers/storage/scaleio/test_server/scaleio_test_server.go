// +build run

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/gotil"

	// load the driver
	"github.com/emccode/libstorage"
	_ "github.com/emccode/libstorage/drivers/storage/scaleio"
	"github.com/emccode/libstorage/drivers/storage/scaleio/executor"
)

func main() {

	// make sure all servers get closed even if the test is abrubptly aborted
	trapAbort()

	if debug, _ := strconv.ParseBool(os.Getenv("LIBSTORAGE_DEBUG")); debug {
		log.SetLevel(log.DebugLevel)
		os.Setenv("LIBSTORAGE_SERVER_HTTP_LOGGING_ENABLED", "true")
		os.Setenv("LIBSTORAGE_SERVER_HTTP_LOGGING_LOGREQUEST", "true")
		os.Setenv("LIBSTORAGE_SERVER_HTTP_LOGGING_LOGRESPONSE", "true")
		os.Setenv("LIBSTORAGE_CLIENT_HTTP_LOGGING_ENABLED", "true")
		os.Setenv("LIBSTORAGE_CLIENT_HTTP_LOGGING_LOGREQUEST", "true")
		os.Setenv("LIBSTORAGE_CLIENT_HTTP_LOGGING_LOGRESPONSE", "true")
	}

	serve("", false)
}

func trapAbort() {
	// make sure all servers get closed even if the test is abrubptly aborted
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		fmt.Println("received abort signal")
		closeAllServers()
		fmt.Println("all servers closed")
		os.Exit(1)
	}()
}

var servers []io.Closer

func closeAllServers() bool {
	noErrors := true
	for _, server := range servers {
		if err := server.Close(); err != nil {
			noErrors = false
			fmt.Printf("error closing server: %v\n", err)
		}
	}
	return noErrors
}

func serve(host string, tls bool) {

	if host == "" {
		host = fmt.Sprintf("tcp://localhost:%d", gotil.RandomTCPPort())
	}
	config := getConfig(host, tls)
	server, errs := libstorage.Serve(config)
	if server != nil {
		servers = append(servers, server)
	}
	<-errs
}

func getConfig(host string, tls bool) gofig.Config {
	if host == "" {
		host = "tcp://127.0.0.1:7979"
	}
	config := gofig.New()

	scaleioConfig := map[string]interface{}{
		"endpoint":             "https://192.168.50.12/api",
		"insecure":             true,
		"useCerts":             false,
		"userName":             "admin",
		"password":             "Scaleio123",
		"systemID":             "6cfe25856a90658d",
		"systemName":           "cluster1",
		"protectionDomainID":   "6d13747300000000",
		"protectionDomainName": "pdomain",
		"storagePoolID":        "672d836d00000000",
		"storagePoolName":      "pool1",
		"thinOrThick":          "ThinProvisioned",
		"version":              "2.0",
	}
	config.Set("scaleio", scaleioConfig)

	var clientTLS, serverTLS string
	if tls {
		clientTLS = fmt.Sprintf(
			libStorageConfigClientTLS,
			clientCrt, clientKey, trustedCerts)
		serverTLS = fmt.Sprintf(
			libStorageConfigServerTLS,
			serverCrt, serverKey, trustedCerts)
	}
	configYaml := fmt.Sprintf(
		libStorageConfigBase,
		host,
    "/tmp/libstorage/executors",
		clientTLS,
    serverTLS,
		serviceName,
    executor.Name)
	fmt.Printf("Config YML %+v", configYaml)

	configYamlBuf := []byte(configYaml)
	if err := config.ReadConfig(bytes.NewReader(configYamlBuf)); err != nil {
		panic(err)
	}
  return config
}

var (
	tlsPath = fmt.Sprintf(
		"%s/src/github.com/emccode/libstorage/.tls", os.Getenv("GOPATH"))
	serverCrt    = fmt.Sprintf("%s/libstorage-server.crt", tlsPath)
	serverKey    = fmt.Sprintf("%s/libstorage-server.key", tlsPath)
	clientCrt    = fmt.Sprintf("%s/libstorage-client.crt", tlsPath)
	clientKey    = fmt.Sprintf("%s/libstorage-client.key", tlsPath)
	trustedCerts = fmt.Sprintf("%s/libstorage-ca.crt", tlsPath)
)

const (
	serviceName = executor.Name

	/*
	   libStorageConfigBase is the base config for tests
	   01 - the host address to server and which the client uses
	   02 - the executors directory
	   03 - the client TLS section. use an empty string if TLS is disabled
	   04 - the server TLS section. use an empty string if TLS is disabled
	   05 - the first service name
	   06 - the first service's driver type
	*/
	libStorageConfigBase = `
libstorage:
  host: %[1]s
  driver: invalidDriverName
  executorsDir: %[2]s
  profiles:
    enabled: true
    groups:
    - local=127.0.0.1%[3]s
  server:
    endpoints:
      localhost:
        address: tcp://192.168.50.12:9000
    services:
      %[5]s:
        libstorage:
          driver: %[6]s
          profiles:
            enabled: true
            groups:
            - remote=127.0.0.1
`

	libStorageConfigClientTLS = `
    tls:
      serverName: libstorage-server
      certFile: %s
      keyFile: %s
      trustedCertsFile: %s
`

	libStorageConfigServerTLS = `
        tls:
          serverName: libstorage-server
          certFile: %s
          keyFile: %s
          trustedCertsFile: %s
          clientCertRequired: true
`
)
