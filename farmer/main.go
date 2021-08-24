package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/taigrr/log-socket/log"

	"github.com/gogrlx/grlx/api"
	"github.com/gogrlx/grlx/certs"
	. "github.com/gogrlx/grlx/config"
	"github.com/gogrlx/grlx/ingredients/test"
	"github.com/gogrlx/grlx/pki"

	// . "github.com/gogrlx/grlx/types"

	nats_server "github.com/nats-io/nats-server/v2/server"
	nats "github.com/nats-io/nats.go"
)

func init() {
	log.SetLogLevel(log.LTrace)
	createConfigRoot()
	pki.SetupPKIFarmer()
}

var s *nats_server.Server

func main() {
	defer log.Flush()
	certs.GenCert()
	certs.GenNKey(true)
	pki.ConfigureNats()
	RunNATSServer()
	StartAPIServer()
	go ConnectFarmer()
	select {}

	// Generate nkey and save or read existing
	// Post user struct to mux
	// Attempt nats auth
	// Auth nats bus
	// Cli accept key, add to config file
	// Update auth users via api

}

func createConfigRoot() {
	_, err := os.Stat(ConfigRoot)
	if err == nil {
		return
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(ConfigRoot, os.ModePerm)
		if err != nil {
			log.Panicf(err.Error())
		}
	} else {
		//TODO: work out what the other errors could be here
		log.Panicf(err.Error())
	}
}

func StartAPIServer() {
	r := api.NewRouter(BuildInfo, CertFile)
	srv := http.Server{
		//TODO: add all below settings to configuration
		Addr:         "0.0.0.0:5405",
		WriteTimeout: time.Second * 120,
		ReadTimeout:  time.Second * 120,
		IdleTimeout:  time.Second * 120,
		Handler:      r,
	}
	go func() {
		if err := srv.ListenAndServeTLS(CertFile, KeyFile); err != nil {
			log.Fatalf(err.Error())
		}
	}()

}

type logger struct {
}

func (l logger) Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// RunNATSServer starts a new Go routine based server
func RunNATSServer() {
	// Optionally override for individual debugging of tests
	// err := opts.ProcessConfigFile("config.json")
	//if err != nil {
	//		log.Panicf("Error configuring server: %v", err)
	//	}
	var err error
	pki.ReloadNKeys()
	s, err = nats_server.NewServer(&DefaultTestOptions)
	if err != nil || s == nil {
		log.Panicf("No NATS Server object returned: %v", err)
	}
	if err != nil || s == nil {
		log.Panicf("No NATS Server object returned: %v", err)
	}
	// Run server in Go routine.
	go s.Start()
	var logger log.Logger
	logger.SetInfoDepth(6)
	s.SetLogger(logger, true, true)
	// Wait for accept loop(s) to be started
	if !s.ReadyForConnections(10 * time.Second) {
		//TODO handle case where nats server port is already taken
		log.Panicf("Unable to start NATS Server in Go Routine")
	}
	//s.ReloadOptions(opts)
	pki.SetNATSServer(s)
	pki.ReloadNKeys()
}

func ConnectFarmer() {
	var connectionAttempts = 1
	var maxFarmerReconnect = 30
	var err error
	opt, err := nats.NkeyOptionFromSeed(NKeyFarmerPrivFile)
	_ = opt
	if err != nil {
		//TODO: handle error
		log.Panic(err)
	}
	certPool := x509.NewCertPool()
	rootPEM, err := ioutil.ReadFile(RootCA)
	if err != nil || rootPEM == nil {
		log.Panicf("nats: error loading or parsing rootCA file: %v", err)
	}
	ok := certPool.AppendCertsFromPEM(rootPEM)
	if !ok {
		log.Errorf("nats: failed to parse root certificate from %v", RootCA)
	}

	config := &tls.Config{
		ServerName: "localhost",
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}
	_ = config
	log.Debug("Attempting to pair Farmer to NATS bus.")
	nc, err := nats.Connect("tls://localhost:4443", //nats.RootCAs(RootCA),
		nats.Secure(config),
		opt,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(maxFarmerReconnect),
		nats.ReconnectWait(time.Second*15),
		nats.DisconnectHandler(func(_ *nats.Conn) {
			connectionAttempts++
			log.Warnf("WARN: Reconnecting Farmer to NATS bus, attempt: %d\n", connectionAttempts)
		}),
	)

	for !nc.IsConnected() {
		connectionAttempts++
		log.Debugf("Attempting to pair Farmer to NATS bus (attempt %d/%d).", connectionAttempts, maxFarmerReconnect)
		if connectionAttempts >= maxFarmerReconnect {
			//		log.Fatalf("Failed to connect Farmer to NATS %d times, exiting.", connectionAttempts)
		}
		time.Sleep(time.Second * 15)
	}
	connectionAttempts = 0
	if err != nil {
		log.Errorf("Got an error on Connect with Secure Options: %+v\n", err)
	}
	log.Debugf("Successfully joined Farmer to NATS bus")

	//	nc, err := nats.Connect(serverUrl, opt)
	//	if err != nil {
	//		//TODO: handle error
	//		panic(err)
	//	}
	ec, _ := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	test.RegisterEC(ec)
	defer ec.Close()
	select {}
}
