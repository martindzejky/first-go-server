package httpUtils

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/martindzejky/first-go-server/internal/os-signals"
)

func RunServer(port string, handler http.Handler) {
	log.Println("Creating the server on port", port)

	// configure tls
	// https://blog.cloudflare.com/exposing-go-on-the-internet/
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519, // Go 1.8 only
		},

		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	// make the server
	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 10,
		IdleTimeout:  time.Second * 60,
		TLSConfig:    tlsConfig,
		Handler:      handler,
	}

	signals := osSignals.MakeChannelWithInterruptSignal()

	log.Println("Registering request handlers")

	go func() {
		log.Println("Starting the server")
		err := server.ListenAndServe()

		if err != nil {
			log.Fatalln(err)
		}
	}()

	// block until the interrupt signal is received
	<-signals

	// make a context to wait for connections
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Gracefully shutting down the server")
	err := server.Shutdown(ctx)

	if err != nil {
		log.Fatalln("Error while shutting down the server:", err)
	}
}
