package main

import (
	"net/http"
	"log"
	"fmt"
	"github.com/voidshard/silo"
	"io/ioutil"
	"strings"
	"flag"
	"time"
	"crypto/tls"
)

const (
	UrlStatus = "/"
)

type App struct {
	repo *silo.Silo
}

// Serve metrics on current users & read / write stats
//
func (a *App) serveMetrics(w http.ResponseWriter, req *http.Request) {
	// TODO: Should probably actually store & return metrics, not simply "Yep I'm alive"
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ok"))
}

// Determine that a user is who they say they are
//
func (a *App) authenticate(w http.ResponseWriter, req *http.Request) *silo.Role {
	username, pass, err := a.getAuth(req)

	suser, err := a.repo.User(username, pass)
	if suser == nil || err != nil {
		log.Println("attempted authentication as user:", username)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("user unknown"))
		return nil
	}

	log.Println("authenticated user:", username)
	return suser
}

// Determine that the given user can perform this request
//
func (a *App) authorize(w http.ResponseWriter, req *http.Request, usr *silo.Role) bool {
	action := req.Method
	path := req.URL.Path

	exists, err := a.repo.Exists(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return false
	}

	if action == http.MethodDelete && usr.CanRm { // delete
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
			return false
		}

		return true
	} else if action == http.MethodGet && usr.CanGet { // read
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
			return false
		}

		return true
	} else if usr.CanPut && action == http.MethodPost { // write
		// To write something, you must use POST and have WRITE.
		// If the file exists, this should return BadRequest (you should use 400)
		if exists {
			if usr.CanRm {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("already exists: cannot overwrite with POST, use PUT"))
				return false
			} else {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("already exists: cannot overwrite"))
				return false
			}
		}

		return true
	} else if usr.CanPut && usr.CanRm && action == http.MethodPut { // overwrite
		// To overwrite something, you must use PUT and have both RM and WRITE
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
			return false
		}

		return true
	}

	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("method forbidden"))
	return false
}

// Do the actual work of the request
//  - We'll first authenticate & then authorize the client & request.
//
func (a *App) serveRequest(w http.ResponseWriter, req *http.Request) {
	suser := a.authenticate(w, req)
	if suser == nil {
		return // no idea who they are
	}

	authorized := a.authorize(w, req, suser)
	if !authorized {
		return // user / action combination not permitted -- we don't need to attempt anything
	}

	action := req.Method
	data := []byte("Ok")
	key := req.URL.Path
	var err error

	if action == http.MethodDelete {
		err = a.repo.Remove(suser, key)
	} else if action == http.MethodPost || action == http.MethodPut {
		in, err := ioutil.ReadAll(req.Body)
		if err == nil {
			err = a.repo.Store(suser, key, in)
		}
	} else if action == http.MethodGet {
		data, err = a.repo.Get(suser, key)
	}

	if err != nil {
		errstring := err.Error()
		status := http.StatusInternalServerError
		if strings.HasPrefix(errstring, silo.ForbiddenPrefix) {
			status = http.StatusForbidden
		}

		w.WriteHeader(status)
		w.Write([]byte(errstring))
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Main serve function
// Required to implement http.Handler
//
func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println(req.Method, req.URL.Path)

	if req.URL.Path == UrlStatus {
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Method Forbidden"))
			return
		}
		a.serveMetrics(w, req)
		return
	}

	a.serveRequest(w, req)
	return
}

func main() {
	// The silo service holds pretty much all the logic, so all we have to do here is read the config,
	// setup silo and proxy requests back & forth .. with a bit of translation.
	//
	configPtr := flag.String("config", "silo.ini", "Config file")
	flag.Parse()

	config, err := parseConfig(*configPtr)
	if err != nil {
		panic(err)
	}

	repo, err := silo.NewSilo(config.SiloConfig)
	if err != nil {
		panic(err)
	}

	app := App{repo: repo}

	bind := fmt.Sprintf("%s:%d", config.Server.HttpHost, config.Server.HttpPort)
	log.Println(bind)

	// Nb. The default http.ListenAndServe funcs do not allow setting of most of these vars
	srv := &http.Server{
		Handler: &app,
		Addr: bind,
		IdleTimeout: 2 * time.Second,
		ReadTimeout: 30 * time.Second,
		WriteTimeout: 30 * time.Second,
		MaxHeaderBytes: config.SiloConfig.Misc.MaxKeyBytes * 2,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		},
	}
	log.Println(srv.ListenAndServeTLS(config.Server.SSLCert, config.Server.SSLKey))
}
