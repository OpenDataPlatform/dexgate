package main

import (
	"dexgate/internal/config"
	"dexgate/internal/director"
	"dexgate/internal/oidcapp"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"os"
)

var log *logrus.Entry

var tokenCount int = 0

func newToken() string {
	tokenCount++
	return fmt.Sprintf("%05d", tokenCount)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("LOGOUT"))
}

func dumpHeader(r *http.Request) {
	for name, values := range r.Header {
		for _, value := range values {
			log.Debugf("%s %s", name, value)
		}
	}
}

func main() {
	config.Setup()
	log = config.GetLog()
	log.Infof("Dexgate %s listening at '%s' to forward to '%s' (Logleve:%s)", config.GetVersion(), config.GetBindAddr(), config.GetTargetURL(), config.GetLogLevel())

	sessionManager := scs.New()

	reverseProxy := &httputil.ReverseProxy{Director: director.NewDirector(config.GetTargetURL())}

	oidcApp, err := oidcapp.NewOidcApp(config.GetOidcConfig(), true)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Unable to instanciate OIDC subsystem:%v'\n", err)
		os.Exit(2)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/logout", handleLogout)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		token := sessionManager.GetString(r.Context(), "token")
		if token == "" {
			// Fresh session. Must enter login process
			lurl := oidcApp.NewLoginURL()
			log.Debugf("Not logged. Will redirect to %s", lurl)
			http.Redirect(w, r, lurl, http.StatusSeeOther)
			//sessionManager.Put(r.Context(), "token", token)
		} else {
			reverseProxy.ServeHTTP(w, r)
		}
	})
	log.Fatal(http.ListenAndServe(config.GetBindAddr(), sessionManager.LoadAndSave(mux)))
}
