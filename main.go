package main

import (
	"dexgate/internal/config"
	"dexgate/internal/director"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
)

var log *logrus.Entry

var tokenCount int = 0

func newToken() string {
	tokenCount++
	return fmt.Sprintf("%05d", tokenCount)
}

func main() {
	config.Setup()
	log = config.GetLog()
	log.Infof("Dexgate %s listening at '%s' to forward to '%s' (Logleve:%s)", config.GetVersion(), config.GetBindAddr(), config.GetTargetURL(), config.GetLogLevel())

	sessionManager := scs.New()

	reverseProxy := &httputil.ReverseProxy{Director: director.NewDirector(config.GetTargetURL())}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//for name, values := range r.Header {
		//	for _, value := range values {
		//		log.Debugf("%s %s", name, value)
		//	}
		//}
		token := sessionManager.GetString(r.Context(), "token")
		if token == "" {
			token = newToken()
			sessionManager.Put(r.Context(), "token", token)
			log.Debugf("New token:%s", token)
		} else {
			log.Debugf("token:%s", token)
		}
		reverseProxy.ServeHTTP(w, r)
	})
	log.Fatal(http.ListenAndServe(config.GetBindAddr(), sessionManager.LoadAndSave(mux)))
}
