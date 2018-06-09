package main

import (
	"net/http"
	"crypto/tls"
	"fmt"
	"encoding/base64"
	"time"
	"io/ioutil"
	"crypto/x509"
)

const (
	// symlinked into etc/ from docker/silo/dist
	SSLPem = "etc/ssl.pem"
)

// build client url
func Url(suffix string, port int) string {
	return fmt.Sprintf("https://localhost:%d/%s", port, suffix)
}

// find the read only user, read from the config file
func ReadOnlyRole(cfg *fileConfig) *entity {
	for _, u := range cfg.Role {
		if u.Get && !u.Put && !u.Del {
			return u
		}
	}
	return nil
}

// find the read/write user, read from the config file
func ReadWriteRole(cfg *fileConfig) *entity {
	for _, u := range cfg.Role {
		if u.Get && u.Put && !u.Del {
			return u
		}
	}
	return nil
}

// find the read/write/delete user, read from the config file
func AllRole(cfg *fileConfig) *entity {
	for _, u := range cfg.Role {
		if u.Get && u.Put && u.Del {
			return u
		}
	}
	return nil
}

// build basic auth authorization header with the given username / password
func BasicAuth(username, password string) string {
	token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	return fmt.Sprintf("Basic %s", token)
}

// Setup & return http client
//
func NewClient() *http.Client {
	// .. Much reading ..
	// https://stackoverflow.com/questions/12122159/golang-how-to-do-a-https-request-with-bad-certificate
	// https://gist.github.com/denji/12b3a568f092ab951456
	// https://github.com/jcbsmpsn/golang-https-example

	cert, err := ioutil.ReadFile(SSLPem)
	if err != nil {
		panic(err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(cert)

	transport := &http.Transport{
		MaxIdleConns:          5,
		IdleConnTimeout:       20 * time.Second,
		TLSHandshakeTimeout:   20 * time.Second,
		TLSClientConfig: &tls.Config{
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			RootCAs: certPool,
			PreferServerCipherSuites: true,
			MinVersion:               tls.VersionTLS12,
		},
	}

	return &http.Client{Transport: transport}
}
