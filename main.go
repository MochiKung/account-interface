package main

import (
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"

	"dev.corp.extreme.co.th/exe-account/account-interface/config"
	"dev.corp.extreme.co.th/exe-account/account-interface/handler/oauth2/handler/token"
)

func main() {
	var status int = 0
	var err error
	signalStopFlag := false
	terminatedFlag := false

	// parse configs
	conf := config.Default

	// setup request handlers
	http.Handle(token.PrefixPath, token.New())

	// listen for termination signals
	termsig := make(chan os.Signal, 1)
	signal.Notify(termsig, os.Interrupt)

	// listen requests
	if conf.Server == nil {
		log.Println("no server config")
		signal.Stop(termsig)
		close(termsig)
		os.Exit(1)
	}

	listener, terminateChannel, err := startServer(conf.Server)
	if err != nil {
		signal.Stop(termsig)
		log.Println(err)
		status |= int(1)
		os.Exit(status)
	} else {
		log.Println("requests listener started")
	}

	// wait for termination
	for !terminatedFlag {
		select {
		case <-termsig:
			if !signalStopFlag {
				signal.Stop(termsig)
				signalStopFlag = true
				log.Println("received termination request. stopping")
				if err := listener.Close(); err != nil {
					log.Printf("failed to close requests listener: %v\n", err)
					close(termsig)
					terminatedFlag = true
					status |= int(2)
				}
				listener = nil
			}
		case terminate := <-terminateChannel:
			log.Printf("requests listener stopped with status %v\n", terminate)
			status |= terminate
			if listener != nil {
				listener.Close()
				listener = nil
			}
			terminatedFlag = true
		}
	}

	if signalStopFlag != true {
		signal.Stop(termsig)
	}

	// clean up
	close(terminateChannel)
	close(termsig)

	os.Exit(status)
}

func startServer(serverConfig *config.Server) (net.Listener, chan int, error) {
	// create server
	terminate := make(chan int, 1)

	log.Printf("starting requests listener on %v\n", serverConfig.Address)
	// create listener
	listener, err := createListener(serverConfig)
	if err != nil {
		return nil, terminate, err
	}

	// start server
	var wg sync.WaitGroup
	server := &http.Server{
		ConnState: func(conn net.Conn, state http.ConnState) {
			switch state {
			case http.StateNew:
				wg.Add(1)
			case http.StateClosed:
				wg.Done()
			}
		},
	}
	server.SetKeepAlivesEnabled(false)

	go func(server *http.Server, listener net.Listener, terminate chan<- int) {
		var status int
		if err := server.Serve(listener); err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Println(err)
				status = 1
			} else {
				status = 0
			}
		} else {
			status = 0
		}

		log.Printf("performing graceful close on all clients\n")
		wg.Wait()

		terminate <- status
	}(server, listener, terminate)
	return listener, terminate, nil
}

func createListener(config *config.Server) (net.Listener, error) {
	if config.Tls == nil {
		return nil, errors.New("invalid server config: missing TLS setup")
	}
	listener, err := net.Listen("tcp", config.Address)
	if err != nil {
		return nil, err
	}

	if config.Tls.Enable {
		// use TLS connection
		if config.Tls.CertificateFile == "" || config.Tls.CertificateKeyFile == "" {
			return nil, errors.New("missing tls certificate file")
		}
		tlsConfig := &tls.Config{}
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(config.Tls.CertificateFile, config.Tls.CertificateKeyFile)
		if err != nil {
			listener.Close()
			listener = nil
			return nil, err
		}
		return tls.NewListener(listener, tlsConfig), nil
	} else {
		return listener, nil
	}
}
