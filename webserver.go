package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/mreliasen/swi-server/game"
	"github.com/mreliasen/swi-server/internal/logger"
	"golang.org/x/crypto/acme/autocert"
)

var server *http.Server

func getSelfSignedOrLetsEncryptCert(certManager *autocert.Manager, domain *string) func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		dirCache, ok := certManager.Cache.(autocert.DirCache)
		if !ok {
			dirCache = "certs"
		}

		keyFile := filepath.Join(string(dirCache), *domain+".key")
		crtFile := filepath.Join(string(dirCache), *domain+".crt")

		certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
		if err != nil {
			logger.Logger.Warn(fmt.Sprintf("%s\nFalling back to Letsencrypt\n", err))
			return certManager.GetCertificate(hello)
		}

		logger.Logger.Trace("Loaded selfsigned certificate.")
		return &certificate, err
	}
}

func StartWebServer(domain *string, gameInstance *game.Game) {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", CORS(gameInstance.HandleRegistration))
	mux.HandleFunc("/check-name-taken", CORS(gameInstance.CheckNameTaken))
	mux.HandleFunc("/", CORS(func(w http.ResponseWriter, r *http.Request) {
		game.HandleWsClient(gameInstance, w, r)
	}))

	logger.Logger.Info(fmt.Sprintf("TLS domain: %s", *domain))
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(*domain),
		Cache:      autocert.DirCache("certs"),
	}

	tlsConfig := certManager.TLSConfig()
	tlsConfig.GetCertificate = getSelfSignedOrLetsEncryptCert(&certManager, domain)
	addr := ":8081"

	server = &http.Server{
		Addr:      addr,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	go http.ListenAndServe(":8080", certManager.HTTPHandler(nil))
	logger.Logger.Info(fmt.Sprintf("Server listening on: %s", server.Addr))

	if err := server.ListenAndServeTLS("", ""); err != nil {
		logger.Logger.Info(err.Error())
	}
}
